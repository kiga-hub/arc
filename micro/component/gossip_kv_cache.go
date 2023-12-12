package component

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/hashicorp/memberlist"
	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
	"github.com/vulcand/oxy/forward"

	"github.com/kiga-hub/arc/configuration"
	"github.com/kiga-hub/arc/leadelection"
	"github.com/kiga-hub/arc/logging"
	"github.com/kiga-hub/arc/micro"
	microConf "github.com/kiga-hub/arc/micro/conf"
	"github.com/kiga-hub/arc/utils"

	_ "go.uber.org/automaxprocs" // 用于获取容器内实际cpu数量
)

// GossipKVCacheNodeMeta is meta for node, MetaMaxSize of memberlist is 512
type GossipKVCacheNodeMeta struct {
	Name           string `json:"n,omitempty"`    //50
	InClusterIP    string `json:"inip,omitempty"` //15
	PrivateIP      string `json:"pip,omitempty"`  //15
	PrivateCluster string `json:"pc,omitempty"`   //15
	GlobalIP       string `json:"gip,omitempty"`  //15
	GlobalCluster  string `json:"gc,omitempty"`   //15
	ServiceName    string `json:"sn,omitempty"`   //50
	Port           int    `json:"p,omitempty"`    //5
	WorkLoad       int    `json:"wl,omitempty"`   // 5byte 工作负载数量
}

// GossipKVCacheComponent is Component for logging
type GossipKVCacheComponent struct {
	micro.EmptyComponent
	l             func() logging.ILogger
	nacosClient   *configuration.NacosClient
	ctx           context.Context
	ctxCancel     context.CancelFunc
	leaderElector *leadelection.LeaderElector
	isSetuped     bool
	list          *memberlist.Memberlist
	opsLock       sync.Mutex
	broadcasts    *memberlist.TransmitLimitedQueue

	// OnJoinCluster will be called when find a new leader or/and join a cluster, it might called multiple times.
	OnJoinCluster func()
	// OnNodeJoin will be called when a node join the cluster
	OnNodeJoin func(*GossipKVCacheNodeMeta)
	// OnNodeLeave will be called when a node leave the cluster
	OnNodeLeave func(*GossipKVCacheNodeMeta)

	InMachineMode bool
	ClusterName   string
	inClusterIP   string
	Port          int

	metadata []byte
	meta     *GossipKVCacheNodeMeta
	metas    map[string]*GossipKVCacheNodeMeta
	items    map[string]string //key-value e.g. sensorid-cluster
}

// Name of the component
func (c *GossipKVCacheComponent) Name() string {
	return "KVCache"
}

// Init the component
func (c *GossipKVCacheComponent) Init(server *micro.Server) error {
	basicConf := microConf.GetBasicConfig()
	if !basicConf.InSwarm {
		return nil
	}

	var err error

	// init
	c.l = micro.GenerateLoggerForModule(server, c.Name())
	c.nacosClient = server.GetElement(&micro.NacosClientElementKey).(*configuration.NacosClient)
	c.ctx, c.ctxCancel = context.WithCancel(context.Background())
	c.opsLock = sync.Mutex{}
	c.items = map[string]string{}
	c.metas = map[string]*GossipKVCacheNodeMeta{}

	var workLoad int
	if basicConf.WorkLoad > 0 {
		// 配置文件对外公布算力
		workLoad = basicConf.WorkLoad
	} else {
		// 获取当前容器cpu数量
		workLoad = runtime.GOMAXPROCS(-1)
	}
	name := basicConf.Service + "." + server.PrivateCluster
	c.meta = &GossipKVCacheNodeMeta{
		Name:           name,
		ServiceName:    basicConf.Service,
		Port:           c.Port,
		PrivateIP:      server.PrivateIP.String(),
		PrivateCluster: server.PrivateCluster,
		GlobalIP:       server.GlobalIP.String(),
		GlobalCluster:  server.GlobalCluster,
		WorkLoad:       workLoad,
	}
	bs, err := json.Marshal(c.meta)
	if err != nil {
		return err
	}
	c.metadata = bs
	c.metas[name] = c.meta

	if c.InMachineMode {
		c.inClusterIP = server.PrivateIP.String()
	} else {
		c.inClusterIP = server.GlobalIP.String()
	}

	config := memberlist.DefaultLANConfig()
	config.Name = name
	config.BindPort = c.Port
	config.AdvertiseAddr = c.inClusterIP
	config.AdvertisePort = c.Port
	config.Events = c
	config.Delegate = c
	config.PushPullInterval = 10 * time.Second

	if !basicConf.IsDevMode {
		config.LogOutput = ioutil.Discard // Drop all memberlist output in non-dev mode
	}

	c.list, err = memberlist.Create(config)
	if err != nil {
		return err
	}

	c.broadcasts = &memberlist.TransmitLimitedQueue{
		NumNodes: func() int {
			return c.list.NumMembers()
		},
		RetransmitMult: 3,
	}
	// spew.Dump(basicConf)
	server.RegisterElement(&micro.GossipKVCacheElementKey, c)
	return nil
}

// PostStart called after Start()
func (c *GossipKVCacheComponent) PostStart(ctx context.Context) error {
	basicConf := microConf.GetBasicConfig()
	if !basicConf.InSwarm {
		return nil
	}
	var err error
	//leadship test
	c.leaderElector, err = leadelection.NewLeaderElector(leadelection.LeaderElectionConfig{
		Lock: leadelection.NewNacosLock(micro.PlatformConfigGroup, fmt.Sprintf("lead-%s", c.ClusterName), c.inClusterIP, c.nacosClient),
		Callbacks: leadelection.LeaderCallbacks{
			OnStartedLeading: func(context.Context) {
				c.l().Debug("OnStartedLeading")
			},
			OnStoppedLeading: func() {
				c.l().Debug("OnStoppedLeading")
			},
			OnNewLeader: func(identity string) {
				if identity == "" {
					c.l().Debug("OnNewLeader empty leader")
					return
				}
				c.l().Debug("OnNewLeader " + identity)
				c.isSetuped = true

				if identity == c.inClusterIP {
					c.l().Debug("OnNewLeader it's me")
					return
				}
				_, err = c.list.Join([]string{fmt.Sprintf("%s:%d", identity, c.Port)})
				if err != nil {
					panic("Failed to join cluster: " + err.Error())
				}
				if c.OnJoinCluster != nil {
					c.OnJoinCluster()
				}
				c.l().Debug("OnNewLeader done")
			},
			OnStopRunning: func() {
				go c.leaderElector.Run(c.ctx)
			},
		},
		LeaseDuration:   time.Second * 35,
		RenewDeadline:   time.Second * 30, // nacos will keep configuration for 30 days, total = 30*24*60*60/30 = 86k
		RetryPeriod:     time.Second * 2,
		ReleaseOnCancel: true,
		Name:            basicConf.Instance,
	}, c.l())
	if err != nil {
		return err
	}
	go c.leaderElector.Run(c.ctx)

	// wait until join cluster or timeout
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		stop := time.NewTimer(time.Second * 30)
		defer stop.Stop()
		for {
			select {
			case <-ticker.C:
				if c.isSetuped {
					return
				}
			case <-stop.C:
				c.l().Debug("join cluster timeout")
				panic("cannot build cluster")
			}
		}
	}()
	return nil
}

// PreStop called before Stop()
func (c *GossipKVCacheComponent) PreStop(ctx context.Context) error {
	// post stop
	basicConf := microConf.GetBasicConfig()
	if !basicConf.InSwarm {
		return nil
	}

	c.ctxCancel()
	return c.list.Leave(time.Second * 3)
}

type broadcast struct {
	msg    []byte
	notify chan<- struct{}
}

func (b *broadcast) Invalidates(other memberlist.Broadcast) bool {
	return false
}

func (b *broadcast) Message() []byte {
	return b.msg
}

func (b *broadcast) Finished() {
	if b.notify != nil {
		close(b.notify)
	}
}

const (
	updateActionAdd = "add"
	updateActionDel = "del"
)

type update struct {
	Action string // add, del
	Data   map[string]string
}

const (
	urlGroupGossip = "Gossip Cluster"
	urlKVCache     = "/kv"
	urlNode        = "/node"
	urlGossip      = "/gossip"
)

// SetupHandler of echo if the component need
func (c *GossipKVCacheComponent) SetupHandler(root echoswagger.ApiRoot, base string) error {
	basicConf := microConf.GetBasicConfig()

	g := root.Group(urlGroupGossip, urlKVCache)
	g.POST("", c.addHandler).
		AddParamQuery("", "key", "key", true).
		AddParamQuery("", "value", "value", false).
		AddResponse(http.StatusOK, "successful operation", "", nil).
		SetOperationId("add").
		SetSummary("add a key-value pair, if value is not set, kv will use consider is a sensorid-cluster item and use its cluster name as value")
	g.GET("", c.getHandler).
		AddParamQuery("", "key", "key", false).
		AddResponse(http.StatusOK, "successful operation", "", nil).
		AddResponse(http.StatusNotFound, "key is set but does not exist", "", nil).
		SetOperationId("get").
		SetSummary("get value, only one item if key is set")
	g.DELETE("", c.deleteHandler).
		AddParamQuery("", "key", "key", true).
		AddResponse(http.StatusOK, "successful operation", "", nil).
		SetOperationId("delete").
		SetSummary("delte a key")

	g = root.Group(urlGroupGossip, urlNode)
	g.GET("", c.getMembersHandler).
		AddResponse(http.StatusOK, "successful operation", []*memberlist.Node{}, nil).
		SetOperationId("getAll").
		SetSummary("get all cluster members")

	g = root.Group(urlGroupGossip, urlGossip)
	g.GET("/demo1", c.SensorIDHandlerWrapper(basicConf.Service, c.demoHandler, false)).
		AddParamQuery("", "sensorid", "sensorid", true).
		AddParamQuery(true, "inside", "inside swarm or not", false).
		SetOperationId("sensorid-cluster-redirect").
		SetSummary("测试sensorid-cluster重定向能力")
	g.GET("/demon", c.SensorIDsHandlerWrapper(basicConf.Service, c.demoHandler, nil)).
		AddParamQuery("", "sensorids", "sensorids", true).
		SetOperationId("sensorids-cluster").
		SetSummary("测试sensorids-cluster重定向或聚合能力")
	return nil
}

func (c *GossipKVCacheComponent) demoHandler(ctx echo.Context) error {
	sensorid := ctx.QueryParam("sensorid")
	if sensorid != "" {
		c.l().Infow("the real response", "sensorid", sensorid, "privateip", c.meta.PrivateIP, "globalip", c.meta.GlobalIP)
		return ctx.JSON(http.StatusOK, utils.Response{
			Status: utils.ResponseStatusOK,
			Data: []string{
				fmt.Sprintf("data for %s, from %s / %s ", sensorid, c.meta.PrivateIP, c.meta.GlobalIP),
			},
		})
	}
	sensorids := ctx.QueryParam("sensorids")
	c.l().Infof("the real response", "sensorid", sensorid, "privateip", c.meta.PrivateIP, "globalip", c.meta.GlobalIP)
	data := []string{}
	for _, v := range strings.Split(sensorids, idSeparater) {
		data = append(data, fmt.Sprintf("data for %s, from %s / %s ", v, c.meta.PrivateIP, c.meta.GlobalIP))
	}
	return ctx.JSON(http.StatusOK, utils.Response{
		Status: utils.ResponseStatusOK,
		Data:   data,
	})
}

func (c *GossipKVCacheComponent) addHandler(ctx echo.Context) error {
	var key, value string
	err := echo.QueryParamsBinder(ctx).
		String("key", &key).
		String("value", &value).
		BindError()
	if err != nil {
		return utils.GetJSONResponse(ctx, err, nil)
	}
	c.l().Debugf("%s => %s-%s", ctx.Request().RequestURI, key, value)
	if value == "" {
		err = c.HaveSensorID(key)
	} else {
		err = c.Add(key, value)
	}
	if err != nil {
		return utils.GetJSONResponse(ctx, err, nil)
	}
	return utils.GetJSONResponse(ctx, nil, "OK")
}

func (c *GossipKVCacheComponent) getHandler(ctx echo.Context) error {
	key := ctx.QueryParam("key")
	if key == "" {
		all, err := c.GetAll()
		if err != nil {
			return utils.GetJSONResponse(ctx, err, nil)
		}
		return utils.GetJSONResponse(ctx, nil, all)
	}

	value, ok := c.Get(key)
	if !ok {
		return ctx.NoContent(http.StatusNotFound)
	}
	return utils.GetJSONResponse(ctx, nil, value)
}

func (c *GossipKVCacheComponent) deleteHandler(ctx echo.Context) error {
	key := ctx.QueryParam("key")
	err := c.Delete(key)
	if err != nil {
		return utils.GetJSONResponse(ctx, err, nil)
	}
	return utils.GetJSONResponse(ctx, nil, "OK")
}

func (c *GossipKVCacheComponent) getMembersHandler(ctx echo.Context) error {
	members := c.GetMembers()
	return utils.GetJSONResponse(ctx, nil, members)
}

// Add a key-value to store
func (c *GossipKVCacheComponent) Add(key, val string) error {
	c.opsLock.Lock()
	c.items[key] = val
	c.opsLock.Unlock()

	b, err := json.Marshal([]*update{
		{
			Action: updateActionAdd,
			Data: map[string]string{
				key: val,
			},
		},
	})

	if err != nil {
		return err
	}

	c.broadcasts.QueueBroadcast(&broadcast{
		msg:    append([]byte("d"), b...),
		notify: nil,
	})
	return nil
}

// AddKeys add a lot of keys with same value to store
func (c *GossipKVCacheComponent) AddKeys(keys []string, sameValue string) error {
	c.opsLock.Lock()
	for _, key := range keys {
		c.items[key] = sameValue
	}
	c.opsLock.Unlock()

	for _, key := range keys {
		b, err := json.Marshal([]*update{
			{
				Action: updateActionAdd,
				Data: map[string]string{
					key: sameValue,
				},
			},
		})

		if err != nil {
			return err
		}

		c.broadcasts.QueueBroadcast(&broadcast{
			msg:    append([]byte("d"), b...),
			notify: nil,
		})
	}
	return nil
}

// Delete a key-value from store
func (c *GossipKVCacheComponent) Delete(key string) error {
	c.opsLock.Lock()
	delete(c.items, key)
	c.opsLock.Unlock()

	b, err := json.Marshal([]*update{
		{
			Action: updateActionDel,
			Data: map[string]string{
				key: "",
			},
		},
	})

	if err != nil {
		return err
	}

	c.broadcasts.QueueBroadcast(&broadcast{
		msg:    append([]byte("d"), b...),
		notify: nil,
	})
	return nil
}

// Get a value of key from store
func (c *GossipKVCacheComponent) Get(key string) (string, bool) {
	c.opsLock.Lock()
	defer c.opsLock.Unlock()
	val, ok := c.items[key]
	return val, ok
}

// GetAll key-value pairs from store
func (c *GossipKVCacheComponent) GetAll() (map[string]string, error) {
	c.opsLock.Lock()
	defer c.opsLock.Unlock()
	result := map[string]string{}
	for k, v := range c.items {
		result[k] = v
	}
	return result, nil
}

// GetMembers return all members in cluster
func (c *GossipKVCacheComponent) GetMembers() []*memberlist.Node {
	if c.list == nil {
		return []*memberlist.Node{}
	}
	return c.list.Members()
}

// memberlist.Delegate interface

// NodeMeta is used to retrieve meta-data about the current node
// when broadcasting an alive message. It's length is limited to
// the given byte size. This metadata is available in the Node structure.
func (c *GossipKVCacheComponent) NodeMeta(limit int) []byte {
	if len(c.metadata) > limit {
		c.l().Error(fmt.Errorf("marshaled meta is %d bytes, which is larger than limit %d bytes", len(c.metadata), limit))
		return []byte{}
	}
	return c.metadata
}

// NotifyMsg is called when a user-data message is received.
// Care should be taken that this method does not block, since doing
// so would block the entire UDP packet receive loop. Additionally, the byte
// slice may be modified after the call returns, so it should be copied if needed
func (c *GossipKVCacheComponent) NotifyMsg(b []byte) {
	if len(b) == 0 {
		return
	}

	switch b[0] {
	case 'd': // data
		var updates []*update
		if err := json.Unmarshal(b[1:], &updates); err != nil {
			return
		}
		c.opsLock.Lock()
		for _, u := range updates {
			for k, v := range u.Data {
				switch u.Action {
				case updateActionAdd:
					c.items[k] = v
				case updateActionDel:
					delete(c.items, k)
				}
			}
		}
		c.opsLock.Unlock()
	}
}

// GetBroadcasts is called when user data messages can be broadcast.
// It can return a list of buffers to send. Each buffer should assume an
// overhead as provided with a limit on the total byte size allowed.
// The total byte size of the resulting data to send must not exceed
// the limit. Care should be taken that this method does not block,
// since doing so would block the entire UDP packet receive loop.
func (c *GossipKVCacheComponent) GetBroadcasts(overhead, limit int) [][]byte {
	return c.broadcasts.GetBroadcasts(overhead, limit)
}

// LocalState is used for a TCP Push/Pull. This is sent to
// the remote side in addition to the membership information. Any
// data can be sent here. See MergeRemoteState as well. The `join`
// boolean indicates this is for a join instead of a push/pull.
func (c *GossipKVCacheComponent) LocalState(join bool) []byte {
	c.opsLock.Lock()
	b, _ := json.Marshal(c.items)
	c.opsLock.Unlock()
	return b
}

// MergeRemoteState is invoked after a TCP Push/Pull. This is the
// state received from the remote side and is the result of the
// remote side's LocalState call. The 'join'
// boolean indicates this is for a join instead of a push/pull.
func (c *GossipKVCacheComponent) MergeRemoteState(buf []byte, join bool) {
	if len(buf) == 0 {
		return
	}

	if !join {
		return
	}

	var m map[string]string
	if err := json.Unmarshal(buf, &m); err != nil {
		c.l().Error(err)
		return
	}
	c.opsLock.Lock()
	for k, v := range m {
		c.items[k] = v
	}
	c.opsLock.Unlock()
}

// memeberlist.EventDelegate interface

func (c *GossipKVCacheComponent) getNodeMetaFromNode(node *memberlist.Node) *GossipKVCacheNodeMeta {
	if len(node.Meta) == 0 {
		c.l().Errorf("no meta found for node %s", node.Name)
		return nil
	}
	var meta GossipKVCacheNodeMeta
	err := json.Unmarshal(node.Meta, &meta)
	if err != nil {
		c.l().Error(err)
		return nil
	}
	return &meta
}

func (c *GossipKVCacheComponent) addNodeMeta(node *memberlist.Node) {
	meta := c.getNodeMetaFromNode(node)
	if meta == nil {
		return
	}
	c.opsLock.Lock()
	c.metas[meta.Name] = meta
	c.opsLock.Unlock()
	if c.OnNodeJoin != nil {
		c.OnNodeJoin(meta)
	}
}
func (c *GossipKVCacheComponent) removeNodeMeta(node *memberlist.Node) {
	meta := c.getNodeMetaFromNode(node)
	if meta == nil {
		return
	}
	c.opsLock.Lock()
	delete(c.metas, meta.Name)
	c.opsLock.Unlock()
	if c.OnNodeLeave != nil {
		c.OnNodeLeave(meta)
	}
}

// NotifyJoin is invoked when a node is detected to have joined.
// The Node argument must not be modified.
func (c *GossipKVCacheComponent) NotifyJoin(node *memberlist.Node) {
	c.l().Debug("A node has joined: " + node.String())
	c.addNodeMeta(node)
}

// NotifyLeave is invoked when a node is detected to have left.
// The Node argument must not be modified.
func (c *GossipKVCacheComponent) NotifyLeave(node *memberlist.Node) {
	c.l().Debugf("A node has left: " + node.String())
	c.removeNodeMeta(node)
}

// NotifyUpdate is invoked when a node is detected to have
// updated, usually involving the meta data. The Node argument
// must not be modified.
func (c *GossipKVCacheComponent) NotifyUpdate(node *memberlist.Node) {
	c.l().Debugf("A node was updated: " + node.String())
	c.addNodeMeta(node)
}

// for sensorid-cluster usage

// FindMemberIPs return global ip and private ip for node of given service in given cluster
func (c *GossipKVCacheComponent) FindMemberIPs(cluster, service string) (string, string, error) { //globalip, privateip, error
	c.opsLock.Lock()
	defer c.opsLock.Unlock()
	meta, ok := c.metas[fmt.Sprintf("%s.%s", service, cluster)]
	if !ok {
		return "", "", errors.New("member not found")
	}
	return meta.GlobalIP, meta.PrivateIP, nil
}

// FindNodeMetasByPrefix 查询符合前缀的节点列表
func (c *GossipKVCacheComponent) FindNodeMetasByPrefix(cluster, servicePrefix string) []*GossipKVCacheNodeMeta {
	c.opsLock.Lock()
	defer c.opsLock.Unlock()
	metasLen := len(c.metas)
	if metasLen == 0 {
		return nil
	}
	// 默认切片容量为map容量一半,奇数情况+1后/2
	nodes := make([]*GossipKVCacheNodeMeta, 0, (metasLen%2+metasLen)/2)
	for k := range c.metas {
		// eg key: kiga-buzz-xxx.172-217
		sp := strings.Split(k, ".")
		if len(sp) < 2 {
			continue
		}
		if strings.Index(sp[0], servicePrefix) != 0 {
			continue
		}
		if cluster != sp[1] {
			continue
		}
		nodes = append(nodes, c.metas[k])
	}
	if len(nodes) == 0 {
		return nil
	}
	return nodes
}

// HaveSensorID add a item with sensorid as key and c.ClusterName as value
func (c *GossipKVCacheComponent) HaveSensorID(sensorid string) error {
	return c.Add(sensorid, c.meta.PrivateCluster)
}

// HaveSensorIDs add a lot of items with sensorid as key and c.ClusterName as value
func (c *GossipKVCacheComponent) HaveSensorIDs(sensorids []string) error {
	return c.AddKeys(sensorids, c.meta.PrivateCluster)
}

// SensorIDHandlerWrapper return a echo.HandlerFunc that will forward/redirect request to the instance which serve with the sensorid if need
// v2 version, if permitSelfHandlerFunc is true, use self handler func when no sensorID
func (c *GossipKVCacheComponent) SensorIDHandlerWrapper(selfServiceName string, f echo.HandlerFunc, permitSelfHandlerFunc bool) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		sensorid := ctx.QueryParam("sensorid")
		collectorid := ctx.QueryParam("collectorid")
		inside := ctx.QueryParam("inside")

		if sensorid == "" && collectorid != "" {
			c.l().Warn("should use sensorid instead of collectorid")
			sensorid = collectorid
		}

		// not a request with sensorid
		if sensorid == "" {
			c.l().Debug("not a request with sensorid")
			return f(ctx)
		}
		c.l().Debugf("handle with sensorid %s", sensorid)

		// sensorid is unknown
		cluster, ok := c.Get(sensorid)
		if !ok {
			c.l().Debugf("response when unknown sensorid")
			if permitSelfHandlerFunc {
				return f(ctx)
			}
			return utils.GetJSONResponse(ctx, nil, "unknown sensorid")
		}

		// find instance for request
		gip, _, err := c.FindMemberIPs(cluster, selfServiceName)

		// hit error
		if err != nil {
			c.l().Error(err)
			return utils.GetJSONResponse(ctx, err, "")
		}
		// self serve
		if gip == c.inClusterIP {
			c.l().Debugf("self serve")
			return f(ctx)
		}

		u := ctx.Request().URL
		u.Host = gip
		u.Scheme = "http"
		addr := u.String()
		ctx.Request().URL = u

		// inside swarm, then redirect
		if strings.ToLower(inside) == "true" {
			c.l().Debugf("inside swarm, need to redirect to %s", addr)
			return ctx.Redirect(http.StatusTemporaryRedirect, addr)
		}

		// outside swarm, then forward, should support websocket
		fwd, err := forward.New()
		// hit error
		if err != nil {
			c.l().Error(err)
			return utils.GetJSONResponse(ctx, err, "")
		}

		c.l().Debugf("outside swarm, need to forward to %s", addr)
		fwd.ServeHTTP(ctx.Response().Writer, ctx.Request())
		return nil
	}
}

const idSeparater = ","

type wsmsg struct {
	Type int
	Data []byte
}

func (c *GossipKVCacheComponent) buildTargetsFromSensorIDs(selfServiceName, sensorids string) (map[string][]string, error) { // ip-[]id
	// group by cluster id, and send with one request
	targets := map[string][]string{} // ip-[]id
	for _, sensorid := range strings.Split(sensorids, idSeparater) {
		if sensorid == "" {
			continue
		}
		// sensorid is unknown
		cluster, ok := c.Get(sensorid)
		if !ok {
			c.l().Debugf("response when unknown sensorid %s", sensorid)
			// return nil, fmt.Errorf("unknown sensorid " + sensorid)
			continue
		}

		// find instance for request
		gip, _, err := c.FindMemberIPs(cluster, selfServiceName)

		// hit error
		if err != nil {
			c.l().Error(err)
			return nil, err
		}
		_, ok = targets[gip]
		if !ok {
			targets[gip] = []string{sensorid}
		} else {
			targets[gip] = append(targets[gip], sensorid)
		}
	}
	return targets, nil
}

// SensorIDsHandlerWrapper return a echo.HandlerFunc that will forward/redirect request to the instances which serve with the sensorids
func (c *GossipKVCacheComponent) SensorIDsHandlerWrapper(selfServiceName string, fhttp echo.HandlerFunc, fws func(*websocket.Conn, int, []byte) error) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		// handle websocket
		if forward.IsWebsocketRequest(ctx.Request()) {
			if fws == nil {
				return utils.GetJSONResponse(ctx, fmt.Errorf("method is not supported websocket"), "")
			}
			return c.wrapperWebsocket(selfServiceName, ctx, fws)
		}
		if fhttp == nil {
			return utils.GetJSONResponse(ctx, fmt.Errorf("method is not supported http"), "")
		}
		return c.wrapperHTTP(selfServiceName, ctx, fhttp)
	}
}

// return data structure is []interface{}
func (c *GossipKVCacheComponent) wrapperHTTP(selfServiceName string, ctx echo.Context, fhttp echo.HandlerFunc) error {
	if ctx.Request().Method != http.MethodGet {
		return utils.GetJSONResponse(ctx, fmt.Errorf("method is not supported by wrapper"), "")
	}

	sensorids := ctx.QueryParam("sensorids")
	collectorids := ctx.QueryParam("collectorids")

	if sensorids == "" && collectorids != "" {
		c.l().Warn("should use sensorids instead of collectorids")
		sensorids = collectorids
	}
	// not a request with sensorid
	if sensorids == "" {
		c.l().Debug("not a request with sensorids")
		return fhttp(ctx)
	}
	c.l().Debugf("handle with sensorids %s", sensorids)

	targets, err := c.buildTargetsFromSensorIDs(selfServiceName, sensorids)
	if err != nil {
		return utils.GetJSONResponse(ctx, err, "")
	}

	c.l().Debug(targets)
	// self serve
	if _, ok := targets[c.inClusterIP]; ok && len(targets) == 1 {
		c.l().Debug("self serve all sensorids")
		return fhttp(ctx)
	}

	// build urls
	addresses := []*url.URL{}
	for ip, ids := range targets {
		sensorids := strings.Join(ids, idSeparater)
		u, _ := url.Parse(ctx.Request().URL.String())
		u.Host = ip
		q := u.Query()
		q.Set("sensorids", sensorids)
		q.Set("collectorids", collectorids)
		u.RawQuery = q.Encode()
		addresses = append(addresses, u)
	}

	data := []interface{}{}
	for _, u := range addresses {
		u.Scheme = "http"
		addr := u.String()
		c.l().Debugf("request %s", addr)
		resp, err := http.Get(addr)
		if err != nil {
			return utils.GetJSONResponse(ctx, err, "")
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return utils.GetJSONResponse(ctx, err, "")
		}
		var marshaled utils.Response
		err = json.Unmarshal(body, &marshaled)
		if err != nil {
			return utils.GetJSONResponse(ctx, err, "")
		}
		if arr, ok := marshaled.Data.([]interface{}); ok {
			data = append(data, arr...)
		}
	}
	return utils.GetJSONResponse(ctx, nil, data)
}

// just forward result from every sub websockets w/o modification
func (c *GossipKVCacheComponent) wrapperWebsocket(selfServiceName string, ctx echo.Context, fws func(*websocket.Conn, int, []byte) error) error {
	msgInChan := make(chan wsmsg, 1024)
	defer close(msgInChan)
	msgOutChan := make(chan wsmsg, 1024)
	defer close(msgOutChan)

	// serve websocket
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	cancelCtx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	wg := sync.WaitGroup{}
	mainConn, err := upgrader.Upgrade(ctx.Response().Writer, ctx.Request(), nil)
	if err != nil {
		c.l().Error(err)
		return utils.GetJSONResponse(ctx, err, "")
	}
	defer mainConn.Close()
	conns := map[string]chan wsmsg{} // ip-conn

	// c.l().Debugf("websocket %s", mainConn.RemoteAddr().String())

	//read loop
	wg.Add(1)
	go func(done context.Context, conn *websocket.Conn, wg *sync.WaitGroup) {
		defer wg.Done()
		for {
			select {
			case <-done.Done():
				return
			default:
			}
			mt, msg, err := conn.ReadMessage()
			if err != nil {
				c.l().Error(err)
				cancelFunc()
				return
			}
			//parse sensorids
			var req map[string]interface{}
			err = json.Unmarshal(msg, &req)
			if err != nil {
				msgInChan <- wsmsg{Type: mt, Data: msg}
				continue
			}
			sensorids, ok := req["sensorids"]
			if !ok {
				sensorids, ok = req["collectorids"]
			}
			// sensorids not found
			if !ok {
				c.l().Debug("not a request with sensorids")
				if len(conns) == 0 { // first message
					c.l().Debug("self serve all sensorids")
					// handle self
					err := fws(conn, mt, msg)
					if err != nil {
						c.l().Error(err)
					}
					cancelFunc()
					return
				}
				msgInChan <- wsmsg{Type: mt, Data: msg}
				continue
			}

			if sensorids.(string) == "" {
				cancelFunc()
				return
			}

			// find sensorids
			targets, err := c.buildTargetsFromSensorIDs(selfServiceName, sensorids.(string))
			if err != nil {
				c.l().Error(err)
				cancelFunc()
				return
			}
			if len(targets) == 0 {
				c.l().Info(fmt.Errorf("cannot find targets, serviceName:%s, sensorID:%s", selfServiceName, sensorids.(string)))
				cancelFunc()
				return
			}
			if len(conns) == 0 { // first message
				c.l().Debug("no connection set yet")
				if _, ok := targets[c.inClusterIP]; ok && len(targets) == 1 {
					c.l().Debug("self serve all sensorids")
					// handle self
					err := fws(conn, mt, msg)
					if err != nil {
						c.l().Error(err)
					}
					cancelFunc()
					return
				}
				c.l().Debug("start read loop")
				// out loop
				wg.Add(1)
				go func(done context.Context, cc *websocket.Conn, wg *sync.WaitGroup) {
					defer wg.Done()
					for {
						select {
						case <-done.Done():
							return
						case wsmsg := <-msgOutChan:
							// c.l().Debug("write message to client")
							err := cc.WriteMessage(wsmsg.Type, wsmsg.Data)
							if err != nil {
								c.l().Error(err)
								cancelFunc()
								return
							}
						}
					}
				}(cancelCtx, conn, wg)
				// in loop
				c.l().Debug("start write loop")
				wg.Add(1)
				go func(done context.Context, wg *sync.WaitGroup) {
					defer wg.Done()
					for {
						select {
						case <-done.Done():
							return
						case wsmsg := <-msgInChan:
							for ip, subconn := range conns {
								c.l().Debugf("write message to %s", ip)
								subconn <- wsmsg
							}
						}
					}
				}(cancelCtx, wg)
			}

			// create ws client to each host
			for ip, ids := range targets {
				_, ok := conns[ip]
				if !ok {
					u, _ := url.Parse(ctx.Request().URL.String())
					u.Scheme = "ws"
					u.Host = ip
					subconn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
					if err != nil {
						c.l().Error(err)
						cancelFunc()
						return
					}
					defer subconn.Close()
					subInChan := make(chan wsmsg, 1024)
					defer close(subInChan)
					// write loop
					wg.Add(1)
					go func(done context.Context, cc *websocket.Conn, wg *sync.WaitGroup) {
						defer wg.Done()
						for {
							select {
							case <-done.Done():
								return
							case wsmsg := <-subInChan:
								// c.l().Debugf("write message to %s", ip)
								err := cc.WriteMessage(wsmsg.Type, wsmsg.Data)
								if err != nil {
									c.l().Error(err)
									cancelFunc()
									return
								}
							}
						}
					}(cancelCtx, subconn, wg)

					// read loop
					wg.Add(1)
					go func(done context.Context, cc *websocket.Conn, wg *sync.WaitGroup) {
						defer wg.Done()
						for {
							select {
							case <-done.Done():
								return
							default:
							}
							mt, msg, err := cc.ReadMessage()
							if err != nil {
								c.l().Error(err)
								cancelFunc()
								return
							}
							// c.l().Debugf("read message from %s", ip)
							msgOutChan <- wsmsg{Type: mt, Data: msg}
						}
					}(cancelCtx, subconn, wg)

					conns[ip] = subInChan
					c.l().Debugf("start websocket proxy for %v", ip)
				}
				subconn := conns[ip]
				// write to each with ids
				reqClone := map[string]interface{}{}
				for k, v := range req {
					reqClone[k] = v
				}
				reqClone["sensorids"] = strings.Join(ids, idSeparater)
				reqClone["collectorids"] = strings.Join(ids, idSeparater)
				bs, err := json.Marshal(reqClone)
				if err != nil {
					c.l().Error(err)
					cancelFunc()
					return
				}
				c.l().Debugf("write filtered message to %s, %s, %v, %s", ip, string(msg), req, string(bs))
				subconn <- wsmsg{Type: mt, Data: bs}
			}
		}
	}(cancelCtx, mainConn, &wg)
	// wait until all goroutines stop
	wg.Wait()
	c.l().Debug("serve websocket proxy done")
	return nil
}
