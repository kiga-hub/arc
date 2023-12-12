package micro

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"
	"syscall"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/labstack/echo-contrib/pprof"
	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	nacosConstant "github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/model"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"github.com/pangpanglabs/echoswagger/v2"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"golang.org/x/net/netutil"
	"golang.org/x/time/rate"

	platformConf "github.com/kiga-hub/arc/conf"
	"github.com/kiga-hub/arc/configuration"
	"github.com/kiga-hub/arc/constant"
	"github.com/kiga-hub/arc/logging"
	"github.com/kiga-hub/arc/micro/conf"
	"github.com/kiga-hub/arc/utils"
)

const (
	// PlatformConfigDataID is dataID of platform config in nacos
	PlatformConfigDataID = "PLATFORM_CONF"
	// PlatformConfigGroup is group of platform in nacos
	PlatformConfigGroup = "PLATFORM_GROUP"

	urlStatus = "/status"
	urlHealth = "/health"
)

var (
	// ErrNeedRestart will notify micro to restart when onConfigChanged
	ErrNeedRestart = errors.New("need restart")
)

// Status is status of micro
type Status struct {
	IsOK       bool                        `json:"is_ok,omitempty"`
	Basic      *conf.BasicConfig           `json:"basic,omitempty"`
	Components map[string]*ComponentStatus `json:"components,omitempty"`
}

// ComponentStatus is status of component
type ComponentStatus struct {
	IsOK   bool                   `json:"is_ok,omitempty"`
	Params map[string]interface{} `json:"params,omitempty"`
}

// IComponent is interface of component
type IComponent interface {
	// Name of the component
	Name() string

	// Status of the component
	Status() *ComponentStatus

	// PreInit called before Init()
	PreInit(ctx context.Context) error

	// Init the component
	Init(server *Server) error

	// PostInit called after Init()
	PostInit(ctx context.Context) error

	// SetDynamicConfig called when get dynamic config for the first time
	SetDynamicConfig(*platformConf.NodeConfig) error

	// OnConfigChanged called when dynamic config changed
	OnConfigChanged(*platformConf.NodeConfig) error

	// GetSubscribeServiceList returns the service that the component need to subscribe
	GetSubscribeServiceList() []string

	// OnServiceChanged called when subscribe service changed
	OnServiceChanged(services []model.SubscribeService, err error)

	// SetupHandler of echo if the component need
	SetupHandler(root echoswagger.ApiRoot, base string) error

	// PreStart called before Start()
	PreStart(ctx context.Context) error

	// Start the component
	Start(ctx context.Context) error

	// PostStart called after Start()
	PostStart(ctx context.Context) error

	// PreStop called before Stop()
	PreStop(ctx context.Context) error

	// Stop the component
	Stop(ctx context.Context) error

	// PostStop called after Stop()
	PostStop(ctx context.Context) error
}

var regkey int = 0

// Server serve as a server composed by components
type Server struct {
	AppName        string
	AppVersion     string
	components     []IComponent
	ctx            context.Context
	e              *echo.Echo
	apiRoot        echoswagger.ApiRoot
	httpServer     http.Server
	nacosClient    *configuration.NacosClient
	PrivateIP      net.IP
	PrivateCluster string
	GlobalIP       net.IP
	GlobalCluster  string
	FlagSet        *pflag.FlagSet
	stopSignal     chan os.Signal

	GzipSkipper func(uri string) bool
	// APIRateSkipper 定义限流器Skipper
	APIRateSkipper func(uri string) bool
	// APIBodySkipper 定义request Body Content-length限制的Skipper
	APIBodySkipper func(uri string) bool
	// APITimeOutSkipper 定义超时Skipper
	APITimeOutSkipper func(uri string) bool

	loggerMicroFunc   func() logging.ILogger
	loggerHTTPFunc    func() logging.ILogger
	loggerConfigFunc  func() logging.ILogger
	loggerMonitorFunc func() logging.ILogger
}

// NewServer create a MicroServer
func NewServer(appname, appversion string, components []IComponent) (*Server, error) {
	server := &Server{
		AppName:    appname,
		AppVersion: appversion,
		components: components,
		ctx:        context.WithValue(context.Background(), &regkey, map[interface{}]interface{}{}),
		stopSignal: make(chan os.Signal, 1),
	}
	return server, nil
}

// GetStatus of the micro server
func (m *Server) GetStatus() Status {
	ok := true
	cs := map[string]*ComponentStatus{}
	for _, v := range m.components {
		status := v.Status()
		cs[v.Name()] = status
		ok = ok && status.IsOK
	}
	return Status{
		IsOK:       ok,
		Basic:      conf.GetBasicConfig(),
		Components: cs,
	}
}

func (m *Server) preInit() error {
	//config
	// from default
	conf.SetDefaultBasicConfig()
	for _, v := range m.components {
		err := v.PreInit(m.ctx)
		if err != nil {
			return err
		}
	}

	// from file
	viper.SetConfigType("toml")
	viper.SetConfigName(m.AppName) // name of config file (without extension)
	viper.AddConfigPath(".")       // optionally look for config in the working directory
	viper.AddConfigPath("./conf")
	viper.AddConfigPath("../conf")

	err := viper.ReadInConfig()
	if err != nil {
		switch err.(type) {
		case viper.ConfigFileNotFoundError:
		default:
			return err
		}
	}

	// from env
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv() // read in environment variables that match

	// from flag
	if m.FlagSet != nil {
		pflag.CommandLine = m.FlagSet
	}
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	err = viper.BindPFlags(pflag.CommandLine)
	if err != nil {
		return err
	}

	// from dynamic
	basicConf := conf.GetBasicConfig()
	if basicConf.IsDynamicConfig {
		configuration.SetDefaultNacosClientConfig()
		configuration.SetDefaultNacosServerConfig()
		nacosClientConfig := configuration.GetNacosClientConfig()
		nacosServerConfig := configuration.GetNacosServerConfig()
		m.nacosClient, err = configuration.NewNacos(*nacosClientConfig, []nacosConstant.ServerConfig{*nacosServerConfig})
		if err != nil {
			spew.Dump(nacosClientConfig)
			spew.Dump(nacosServerConfig)
			return err
		}
		m.RegisterElement(&NacosClientElementKey, m.nacosClient)
		config, err := m.nacosClient.Get(PlatformConfigDataID, PlatformConfigGroup)
		if err != nil {
			return err
		}
		err = m.setDynamicConfig(config)
		if err != nil {
			return err
		}
		err = m.nacosClient.Listen(PlatformConfigDataID, PlatformConfigGroup, func(namespace, group, dataId, config string) {
			if group != PlatformConfigGroup || dataId != PlatformConfigDataID || namespace != "" {
				fmt.Printf("!!! %s, %s, %s\n", namespace, group, dataId)
				return
			}
			err0 := m.onConfigChanged(config)
			if err0 != nil {
				m.loggerConfigFunc().Error(err0)
				if err0 == ErrNeedRestart {
					m.stopSignal <- syscall.SIGINT
				}
			}
		})
		if err != nil {
			return err
		}
	}

	//setup network
	ips, err := utils.GetAllIPv4Address()
	if err != nil {
		return err
	}
	if len(ips) == 0 {
		return fmt.Errorf("no ip address found")
	} else if !basicConf.InSwarm {
		ipstrs := strings.Split(ips[0].String(), ".")
		cluster := fmt.Sprintf("%s-%s", ipstrs[0], ipstrs[1]) //cluster name can only have these characters: 0-9a-zA-Z-
		m.GlobalCluster = cluster
		m.GlobalIP = ips[0]
		m.PrivateCluster = "127-0"
		m.PrivateIP = net.ParseIP("127.0.0.1")
	} else {
		for _, v := range ips {
			ip := v
			ipstrs := strings.Split(ip.String(), ".")
			cluster := fmt.Sprintf("%s-%s", ipstrs[0], ipstrs[1]) //cluster name can only have these characters: 0-9a-zA-Z-
			if !utils.IsKigaOverlayIP(ip) {
				m.PrivateCluster = cluster
				m.PrivateIP = ip
			} else {
				m.GlobalCluster = cluster
				m.GlobalIP = ip
			}
		}
	}
	if m.GlobalIP == nil || m.PrivateIP == nil {
		spew.Dump(basicConf)
		spew.Dump(ips)
		spew.Dump(m.GlobalIP)
		spew.Dump(m.PrivateIP)
		return fmt.Errorf("cannot find private or global ip")
	}
	fmt.Println("GlobalIP: " + m.GlobalIP.String())
	fmt.Println("PrivateIP: " + m.PrivateIP.String())

	// fixed
	viper.Set(conf.BasicMachine, m.PrivateCluster)
	viper.Set(conf.BasicAppName, m.AppName)
	viper.Set(conf.BasicAppVersion, m.AppVersion)
	name, err := os.Hostname()
	if err != nil {
		return err
	}
	viper.Set(conf.BasicInstance, name)

	//dump
	if basicConf.IsDevMode {
		basicConf = conf.GetBasicConfig()
		spew.Dump(basicConf)
	}

	return nil
}

func (m *Server) setDynamicConfig(data string) error {
	basicConf := conf.GetBasicConfig()
	pc := &platformConf.TopologyConfig{}
	err := json.Unmarshal([]byte(data), pc)
	if err != nil {
		fmt.Println("setDynamicConfig hit error")
		spew.Dump(data)
		return err
	}
	zone, ok := pc.Zones[basicConf.Zone]
	if !ok {
		if basicConf.IsDevMode {
			fmt.Printf("zone %s not found\n", basicConf.Zone)
		}
		return nil
	}
	node, ok := zone.Nodes[basicConf.Node]
	if !ok {
		if basicConf.IsDevMode {
			fmt.Printf("node %s not found\n", basicConf.Node)
		}
		return nil
	}
	for _, v := range m.components {
		err = v.SetDynamicConfig(&node)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Server) onConfigChanged(data string) error {
	basicConf := conf.GetBasicConfig()
	pc := &platformConf.TopologyConfig{}
	err := json.Unmarshal([]byte(data), pc)
	if err != nil {
		fmt.Println("onConfigChanged hit error")
		spew.Dump(data)
		return err
	}
	zone, ok := pc.Zones[basicConf.Zone]
	if !ok {
		return nil
	}
	node, ok := zone.Nodes[basicConf.Node]
	if !ok {
		return nil
	}
	for _, v := range m.components {
		err := v.OnConfigChanged(&node)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Server) subscribeService() error {
	for _, v := range m.components {
		list := v.GetSubscribeServiceList()
		for _, service := range list {
			err := m.nacosClient.Subscribe([]string{m.GlobalCluster, m.PrivateCluster}, PlatformConfigGroup, service, v.OnServiceChanged)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// GetElement return an element(not component) from micro server
func (m *Server) GetElement(key interface{}) interface{} {
	//fmt.Println(key)
	d := m.ctx.Value(&regkey).(map[interface{}]interface{})
	/*
		for k, v := range d {
			fmt.Println(k, v)
		}
	*/
	return d[key]
}

// RegisterElement register an element(not component) of micro server
func (m *Server) RegisterElement(key, val interface{}) {
	//fmt.Println(key)
	//fmt.Println(val)
	d := m.ctx.Value(&regkey).(map[interface{}]interface{})
	d[key] = val
	/*
		for k, v := range d {
			fmt.Println(k, v)
		}
	*/

}

// Init the micro server
func (m *Server) Init() error {
	err := m.preInit()
	if err != nil {
		return err
	}
	for _, v := range m.components {
		err = v.Init(m)
		if err != nil {
			return err
		}
	}
	return m.postInit()
}

func (m *Server) postInit() error {
	var err error
	for _, v := range m.components {
		err = v.PostInit(m.ctx)
		if err != nil {
			return err
		}
	}
	// subscribe services
	basicConf := conf.GetBasicConfig()
	if basicConf.IsDynamicConfig {
		err = m.subscribeService()
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Server) preStart() error {
	for _, v := range m.components {
		err := v.PreStart(m.ctx)
		if err != nil {
			return err
		}
	}

	//loggers
	m.loggerHTTPFunc = GenerateLoggerForModule(m, "http")
	m.loggerConfigFunc = GenerateLoggerForModule(m, "config")
	m.loggerMicroFunc = GenerateLoggerForModule(m, "micro")
	m.loggerMonitorFunc = GenerateLoggerForModule(m, "monitor")

	basicConf := conf.GetBasicConfig()
	swaggerURI := basicConf.APIRoot + "/swagger"
	swaggerAssetsPath := basicConf.APIRoot + "/static/swagger"

	// start http server
	m.e = echo.New()
	m.e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			uri := c.Request().RequestURI
			if !basicConf.IsDevMode {
				if uri == urlHealth || uri == constant.URLMetrics || uri == swaggerURI {
					return zapLoggerEchoMiddleware(m.loggerHTTPFunc, true)(next)(c)
				}
				if strings.Contains(uri, swaggerAssetsPath) {
					return next(c)
				}
			}
			return zapLoggerEchoMiddleware(m.loggerHTTPFunc, false)(next)(c)
		}
	})

	// 启用请求Body限制中间件
	if basicConf.IsAPIBody {
		m.e.Use(middleware.BodyLimitWithConfig(middleware.BodyLimitConfig{
			Limit: basicConf.APIBodyLimit,
			Skipper: func(ctx echo.Context) bool {
				if m.APIBodySkipper != nil {
					uri := ctx.Request().RequestURI
					return m.APIBodySkipper(uri)
				}
				return false
			},
		}))
	}

	// 启用超时
	if basicConf.IsAPITimeout {
		// 超时信息
		msgTimeOut, err := json.Marshal(utils.ResponseV2{
			Code: http.StatusRequestTimeout,
			Msg:  http.StatusText(http.StatusRequestTimeout),
		})
		if err != nil {
			return err
		}
		// 注册超时中间件
		m.e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
			Skipper: func(ctx echo.Context) bool {
				// 忽略对websocket校验
				if ctx.IsWebSocket() {
					return true
				}
				if m.APITimeOutSkipper != nil {
					uri := ctx.Request().RequestURI
					return m.APITimeOutSkipper(uri)
				}
				return false
			},
			ErrorMessage:               string(msgTimeOut),
			OnTimeoutRouteErrorHandler: nil,
			Timeout:                    time.Millisecond * time.Duration(basicConf.APITimeout),
		}))
	}

	// 启用限流器核心配置
	if basicConf.IsAPIRate {
		var identifier string
		rateLimit := middleware.RateLimiterConfig{
			Skipper: func(ctx echo.Context) bool {
				if m.APIRateSkipper != nil {
					uri := ctx.Request().RequestURI
					return m.APIRateSkipper(uri)
				}
				return false
			},
			Store: middleware.NewRateLimiterMemoryStoreWithConfig(middleware.RateLimiterMemoryStoreConfig{
				Rate:      rate.Limit(basicConf.APIRate), //请求数/秒
				Burst:     basicConf.APIBurst,
				ExpiresIn: time.Duration(basicConf.APIExpiresIn) * time.Second,
			}),
			// 使用请求参数sensorID,限制读取访问频率
			IdentifierExtractor: func(ctx echo.Context) (string, error) {
				if identifier = ctx.QueryParam("sensorid"); identifier != "" {
					return identifier, nil
				}
				// sensorID为空，则使用客户端的IP
				identifier = ctx.RealIP()
				return identifier, nil
			},
			ErrorHandler: func(context echo.Context, err error) error {
				// 访问标识符返回值为定义nil,此处直接返回nil即可
				return nil
			},
			DenyHandler: func(context echo.Context, identifier string, err error) error {
				return context.JSON(http.StatusTooManyRequests, utils.ResponseV2{
					Code: http.StatusTooManyRequests,
					Msg:  http.StatusText(http.StatusTooManyRequests) + ":" + identifier,
				})
			},
		}
		// 注册请求限速中间件
		m.e.Use(middleware.RateLimiterWithConfig(rateLimit))
	}

	m.e.Use(middleware.Recover())
	m.e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Skipper: func(c echo.Context) bool {
			// proxy 模式(防止多次压缩损坏数据)
			xf := c.Request().Header.Get("X-Forwarded-For")
			// 转发者不压缩，接收请求节点压缩
			if xf != "" && !strings.Contains(xf, ",") {
				return true
			}

			uri := c.Request().RequestURI
			if uri == constant.URLMetrics || strings.Contains(uri, "debug/pprof") {
				return true
			}

			if m.GzipSkipper != nil {
				return m.GzipSkipper(uri)
			}

			return false
		},
	}))
	//metric
	p := prometheus.NewPrometheus("echo", func(c echo.Context) bool {
		uri := c.Request().RequestURI
		return uri == urlHealth || uri == constant.URLMetrics
	})
	p.Use(m.e)

	//health
	hf := func(c echo.Context) error {
		status := m.GetStatus()
		health := Health{
			IsHealth: status.IsOK,
		}
		return c.JSON(http.StatusOK, health)
	}
	m.e.GET(urlHealth, hf)

	//status
	sf := func(c echo.Context) error {
		return c.JSON(http.StatusOK, m.GetStatus())
	}
	m.e.GET(urlStatus, sf)

	// pprof - 存储性能分析
	if basicConf.IsProf {
		pprof.Register(m.e, basicConf.APIRoot+"/debug/pprof")
	}

	// api
	m.e.Static(swaggerAssetsPath, "./swagger")
	// swagger spec path is `/swagger/swagger.json`
	m.apiRoot = echoswagger.New(m.e, swaggerURI, &echoswagger.Info{
		Title:   basicConf.AppName,
		Version: basicConf.AppVersion,
	}).SetRequestContentType("application/json").
		SetResponseContentType("application/json").
		SetUI(echoswagger.UISetting{
			CDN: swaggerAssetsPath,
		})
	m.httpServer = http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", basicConf.APIPort),
		Handler: m.e, // set Echo as handler
	}

	g := m.apiRoot.Group("Micro", basicConf.APIRoot)

	g.GET("/status", sf).
		AddResponse(http.StatusOK, "", Status{}, nil).
		SetOperationId("getStatus").
		SetSummary("获取服务状态")

	g.GET("/health", hf).
		AddResponse(http.StatusOK, "", Health{}, nil).
		SetOperationId("getHealth").
		SetSummary("获取服务健康")

	for _, v := range m.components {
		err := v.SetupHandler(m.apiRoot, basicConf.APIRoot)
		if err != nil {
			return err
		}
	}

	if basicConf.IsDevMode {
		go DoMonitor(15, func(stat *MonitorStat) {
			m.loggerMonitorFunc().Debugw("DoMonitor",
				"alloc", stat.Alloc,
				"total", stat.TotalAlloc,
				"sys", stat.Sys,
				"tmallocs", stat.Mallocs,
				"frees", stat.Frees,
				"pause", stat.PauseTotalNs,
				"GC", stat.NumGC,
				"goroutine", stat.NumGoroutine,
				"heap_idle", stat.HeapIdle,
				"heap_inuse", stat.HeapInuse,
				"heap_released", stat.HeapReleased,
				"stack", stat.StackInuse,
			)
		})
	}
	return nil
}

// StartServe start the micro server
func (m *Server) StartServe() error {
	err := m.preStart()
	if err != nil {
		return err
	}
	for _, v := range m.components {
		err = v.Start(m.ctx)
		if err != nil {
			return err
		}
	}
	l, err := net.Listen("tcp4", m.httpServer.Addr)
	if err != nil {
		return err
	}
	ln := netutil.LimitListener(l, 128)
	go func() {
		m.loggerHTTPFunc().Info("start http server")
		err = m.httpServer.Serve(ln)
		if err != nil {
			m.loggerHTTPFunc().Error(err)
		}
		err = ln.Close()
		if err != nil {
			m.loggerHTTPFunc().Error(err)
		}
		err = l.Close()
		if err != nil {
			m.loggerHTTPFunc().Error(err)
		}
		m.loggerHTTPFunc().Info("stop http server")
	}()
	return m.postStart()
}

func (m *Server) postStart() error {
	for _, v := range m.components {
		err := v.PostStart(m.ctx)
		if err != nil {
			return err
		}
	}
	basicConf := conf.GetBasicConfig()
	if !basicConf.IsDynamicConfig {
		return nil
	}

	roP := vo.RegisterInstanceParam{
		Ip:          m.PrivateIP.String(), //required
		Port:        80,                   //required
		Weight:      100,                  //required,it must be lager than 0
		Enable:      true,                 //required,the instance can be access or not
		Healthy:     true,                 //required,the instance is health or not
		Ephemeral:   true,
		ServiceName: basicConf.Service,   //required
		GroupName:   PlatformConfigGroup, //optional,default:DEFAULT_GROUP
		ClusterName: m.PrivateCluster,    //optional,default:DEFAULT
	}
	err := m.nacosClient.Register(roP)
	if err != nil {
		spew.Dump(roP)
		return err
	}

	roG := vo.RegisterInstanceParam{
		Ip:          m.GlobalIP.String(), //required
		Port:        80,                  //required
		Weight:      10,                  //required,it must be lager than 0
		Enable:      true,                //required,the instance can be access or not
		Healthy:     true,                //required,the instance is health or not
		Ephemeral:   true,
		ServiceName: basicConf.Service,   //required
		GroupName:   PlatformConfigGroup, //optional,default:DEFAULT_GROUP
		ClusterName: m.GlobalCluster,     //optional,default:DEFAULT
	}
	err = m.nacosClient.Register(roG)
	if err != nil {
		spew.Dump(roG)
		return err
	}

	return nil
}

func (m *Server) preStop() error {
	basicConf := conf.GetBasicConfig()
	if basicConf.IsDynamicConfig {
		doP := vo.DeregisterInstanceParam{
			Ip:          m.PrivateIP.String(), //required
			Port:        80,                   //required
			Ephemeral:   true,
			ServiceName: basicConf.Service,   //required
			GroupName:   PlatformConfigGroup, //optional,default:DEFAULT_GROUP
			Cluster:     m.PrivateCluster,    //optional,default:DEFAULT
		}
		err := m.nacosClient.Deregister(doP)
		if err != nil {
			spew.Dump(doP)
			return err
		}

		doG := vo.DeregisterInstanceParam{
			Ip:          m.GlobalIP.String(), //required
			Port:        80,                  //required
			Ephemeral:   true,
			ServiceName: basicConf.Service,   //required
			GroupName:   PlatformConfigGroup, //optional,default:DEFAULT_GROUP
			Cluster:     m.GlobalCluster,     //optional,default:DEFAULT
		}
		err = m.nacosClient.Deregister(doG)
		if err != nil {
			spew.Dump(doG)
			return err
		}
	}

	// Stop the service gracefully.
	err := m.httpServer.Shutdown(context.TODO())
	if err != nil {
		return err
	}

	for _, v := range m.components {
		err := v.PreStop(m.ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

// StopServe stop the micro server
func (m *Server) StopServe() error {
	err := m.preStop()
	if err != nil {
		return err
	}

	for _, v := range m.components {
		err = v.Stop(m.ctx)
		if err != nil {
			return err
		}
	}
	return m.postStop()
}

func (m *Server) postStop() error {
	for _, v := range m.components {
		err := v.PostStop(m.ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

// Run the micro server until fetch SIGINT or SIGTERM signal
func (m *Server) Run() error {
	defer func() {
		if r := recover(); r != nil {
			m.loggerMicroFunc().Errorw("Recovered", "recover", r)
			fmt.Println("Recovered", "recover", r)
			debug.PrintStack()
		}
	}()

	err := m.StartServe()
	if err != nil {
		return err
	}

	// Handle SIGINT and SIGTERM.
	signal.Notify(m.stopSignal, syscall.SIGINT, syscall.SIGTERM)
	<-m.stopSignal

	return m.StopServe()
}
