package component

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/kiga-hub/arc/configuration"
	"github.com/kiga-hub/arc/leadelection"
	"github.com/kiga-hub/arc/logging"
	"github.com/kiga-hub/arc/micro"
	microConf "github.com/kiga-hub/arc/micro/conf"
	"github.com/kiga-hub/arc/utils"

	"github.com/go-resty/resty/v2"
	"github.com/labstack/echo/v4"
	"github.com/lni/dragonboat/v3"
	"github.com/lni/dragonboat/v3/config"
	"github.com/lni/dragonboat/v3/statemachine"
	"github.com/pangpanglabs/echoswagger/v2"
)

// TODO This component is not working yet, 3 node in 1 instance is NOT acceptable in prod.

/*
usage:
	&RaftClusterComponent{
		ClusterID: 10000,
		CreateFun: statemachine.NewExampleStateMachine,
	},
*/

// RaftClusterComponent is Component for logging
type RaftClusterComponent struct {
	micro.EmptyComponent
	logger        logging.ILogger
	raftNode      map[uint64]*dragonboat.NodeHost
	nacosClient   *configuration.NacosClient
	ip            net.IP
	ctx           context.Context
	ctxCancel     context.CancelFunc
	CreateFun     statemachine.CreateStateMachineFunc
	raftAddress   map[uint64]string
	raftNodeID    []uint64
	ClusterID     uint64
	leaderElector *leadelection.LeaderElector
	opsLock       sync.Mutex
}

// Name of the component
func (c *RaftClusterComponent) Name() string {
	return "LeaderRaftClusterComponent"
}

// PreInit called before Init()
func (c *RaftClusterComponent) PreInit(ctx context.Context) error {
	// load config
	//conf.SetDefaultLogConfig()
	return nil
}

// Init the component
func (c *RaftClusterComponent) Init(server *micro.Server) error {
	// init
	c.logger = server.GetElement(&micro.LoggingElementKey).(logging.ILogger)
	c.nacosClient = server.GetElement(&micro.NacosClientElementKey).(*configuration.NacosClient)
	c.ip = server.GlobalIP
	c.ctx, c.ctxCancel = context.WithCancel(context.Background())
	id := inetToUint64(c.ip)
	c.raftNodeID = []uint64{
		id, id << 1, id << 2,
	}
	c.raftAddress = map[uint64]string{
		id:      server.GlobalIP.String() + ":12345",
		id << 1: server.GlobalIP.String() + ":12346",
		id << 2: server.GlobalIP.String() + ":12347",
	}
	c.raftNode = map[uint64]*dragonboat.NodeHost{}
	c.opsLock = sync.Mutex{}
	return os.MkdirAll("/raft", os.ModePerm)
}

// PostStart called after Start()
func (c *RaftClusterComponent) PostStart(ctx context.Context) error {
	var err error
	basicConf := microConf.GetBasicConfig()
	//leadship test
	c.leaderElector, err = leadelection.NewLeaderElector(leadelection.LeaderElectionConfig{
		Lock: leadelection.NewNacosLock("DEFAULT_GROUP", fmt.Sprintf("lead-%d", c.ClusterID), c.ip.String(), c.nacosClient),
		Callbacks: leadelection.LeaderCallbacks{
			OnStartedLeading: func(context.Context) {
				fmt.Println("OnStartedLeading")
				for len(c.raftNode) == 0 {
					time.Sleep(time.Second)
					c.startRaftCluster()
				}
			},
			OnStoppedLeading: func() {
				fmt.Println("OnStoppedLeading")
			},
			OnNewLeader: func(identity string) {
				if identity == c.ip.String() {
					return
				}
				fmt.Println("OnNewLeader " + identity)
				time.Sleep(time.Second)
				for len(c.raftNode) == 0 {
					c.joinRaftCluster(identity)
				}
			},
		},
		LeaseDuration:   time.Second * 5,
		RenewDeadline:   time.Second * 3,
		RetryPeriod:     time.Second * 2,
		ReleaseOnCancel: true,
		Name:            basicConf.Instance,
	}, c.logger)
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
				if len(c.raftNode) > 0 {
					return
				}
			case <-stop.C:
				fmt.Println("超时")
				panic("cannot build cluster")
			}
		}
	}()
	return nil
}

func (c *RaftClusterComponent) startRaftCluster() {
	c.opsLock.Lock()
	defer c.opsLock.Unlock()
	for _, nodeID := range c.raftNodeID {
		time.Sleep(time.Second)
		c.startOneRaftCluster(nodeID)
	}
}

func (c *RaftClusterComponent) startOneRaftCluster(nodeID uint64) {
	_, ok := c.raftNode[nodeID]
	if ok {
		return
	}

	fmt.Printf("start raft node %d\n", nodeID)
	var err error
	rc := config.Config{
		// ClusterID and NodeID of the raft node
		NodeID:             nodeID,
		ClusterID:          c.ClusterID,
		ElectionRTT:        10,
		HeartbeatRTT:       1,
		CheckQuorum:        true,
		SnapshotEntries:    10,
		CompactionOverhead: 5,
	}
	datadir := filepath.Join("/raft", fmt.Sprintf("node%d", nodeID))
	// raft node
	nhc := config.NodeHostConfig{
		WALDir:         datadir,
		NodeHostDir:    datadir,
		RTTMillisecond: 200,
		RaftAddress:    c.raftAddress[nodeID],
	}
	c.raftNode[nodeID], err = dragonboat.NewNodeHost(nhc)
	if err != nil {
		panic(err)
	}
	err = c.raftNode[nodeID].StartCluster(c.raftAddress, false, c.CreateFun, rc)
	if err != nil {
		c.logger.Error(err)
		delete(c.raftNode, nodeID)
	}
}

/*
func inet_ntoa(ipnr uint64) net.IP {
	var bytes [4]byte
	bytes[0] = byte(ipnr & 0xFF)
	bytes[1] = byte((ipnr >> 8) & 0xFF)
	bytes[2] = byte((ipnr >> 16) & 0xFF)
	bytes[3] = byte((ipnr >> 24) & 0xFF)

	return net.IPv4(bytes[3], bytes[2], bytes[1], bytes[0])
}
*/

func inetToUint64(ipnr net.IP) uint64 {
	bits := strings.Split(ipnr.String(), ".")

	b0, _ := strconv.Atoi(bits[0])
	b1, _ := strconv.Atoi(bits[1])
	b2, _ := strconv.Atoi(bits[2])
	b3, _ := strconv.Atoi(bits[3])

	var sum uint64

	sum += uint64(b0) << 24
	sum += uint64(b1) << 16
	sum += uint64(b2) << 8
	sum += uint64(b3)

	return sum
}

func (c *RaftClusterComponent) submitRequest(nodeID uint64, leadip string, add bool) error {
	client := resty.New()
	req := joinRequest{
		NodeID:   nodeID,
		NodeAddr: c.raftAddress[nodeID],
	}
	if add {
		req.CMD = joinRequestCmdAdd
	} else {
		req.CMD = joinRequestCmdRemove
	}

	resp, err := client.SetTimeout(time.Second * 3).R().
		SetBody(req).
		Post(fmt.Sprintf("http://%s%s", leadip, urlRequest))
	if err != nil {
		return err
	}
	code := resp.StatusCode()
	if code != http.StatusOK {
		return errors.New(string(resp.Body()))
	}
	return nil
}

func (c *RaftClusterComponent) joinRaftCluster(leadip string) {
	c.opsLock.Lock()
	defer c.opsLock.Unlock()
	for _, nodeID := range c.raftNodeID {
		time.Sleep(time.Second)
		c.joinOneRaftCluster(nodeID, leadip)
	}
}

func (c *RaftClusterComponent) joinOneRaftCluster(nodeID uint64, leadip string) {
	_, ok := c.raftNode[nodeID]
	if ok {
		return
	}

	fmt.Printf("join raft node %d\n", nodeID)

	var err error
	datadir := filepath.Join("/raft", fmt.Sprintf("node%d", nodeID))
	rc := config.Config{
		// ClusterID and NodeID of the raft node
		NodeID:             nodeID,
		ClusterID:          c.ClusterID,
		ElectionRTT:        10,
		HeartbeatRTT:       1,
		CheckQuorum:        true,
		SnapshotEntries:    10,
		CompactionOverhead: 5,
	}
	// raft node
	nhc := config.NodeHostConfig{
		WALDir:         datadir,
		NodeHostDir:    datadir,
		RTTMillisecond: 200,
		RaftAddress:    c.raftAddress[nodeID],
	}
	c.raftNode[nodeID], err = dragonboat.NewNodeHost(nhc)
	if err != nil {
		panic(err)
	}
	err = c.raftNode[nodeID].StartCluster(map[uint64]string{}, true, c.CreateFun, rc)
	if err != nil {
		c.logger.Error(err)
		delete(c.raftNode, nodeID)
	}

	// ask to join if fall means lead down?
	err = c.submitRequest(nodeID, leadip, true)
	if err != nil {
		c.logger.Error(err)
		c.raftNode[nodeID].Stop()
		delete(c.raftNode, nodeID)
	}
}

const (
	urlGroupCluster = "Cluster"
	urlRequest      = "/raft/request"
	//urlNodes        = "/raft/nodes"
)

// SetupHandler of echo if the component need
func (c *RaftClusterComponent) SetupHandler(root echoswagger.ApiRoot, base string) error {
	g := root.Group(urlGroupCluster, urlRequest)
	g.POST("", c.handleRequest).
		AddParamBody(joinRequest{}, "request", "request to fulfil", true).
		AddResponse(http.StatusOK, "successful operation", "", nil).
		SetSummary("submit a request to add/remove a node")

	g.GET("", func(ctx echo.Context) error {
		id := c.raftNodeID[0]
		if c.raftNode[id] == nil {
			return utils.GetJSONResponse(ctx, errors.New("server is not in a cluster yet"), nil)
		}
		cs := c.raftNode[id].GetNoOPSession(c.ClusterID)
		// input is a regular message need to be proposed
		cc, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		// make a proposal to update the IStateMachine instance
		result, err := c.raftNode[id].SyncPropose(cc, cs, []byte("abc"))
		cancel()
		if err != nil {
			return utils.GetJSONResponse(ctx, err, nil)
		}
		return utils.GetJSONResponse(ctx, nil, result)
	})

	return nil
}

type joinRequest struct {
	CMD      string `json:"cmd,omitempty"`
	NodeID   uint64 `json:"node_id,omitempty"`
	NodeAddr string `json:"node_addr,omitempty"`
}

const (
	joinRequestCmdAdd    = "ADD"
	joinRequestCmdRemove = "REMOVE"
)

func (c *RaftClusterComponent) handleRequest(ctx echo.Context) error {
	req := &joinRequest{}
	err := ctx.Bind(req)
	if err != nil {
		return utils.GetJSONResponse(ctx, err, nil)
	}
	id := c.raftNodeID[0]
	if c.raftNode[id] == nil {
		return utils.GetJSONResponse(ctx, errors.New("server is not in a cluster yet"), nil)
	}
	var rs *dragonboat.RequestState
	if req.CMD == joinRequestCmdAdd {
		rs, err = c.raftNode[id].RequestAddNode(c.ClusterID, req.NodeID, req.NodeAddr, 0, 3*time.Second)
	} else if req.CMD == joinRequestCmdRemove {
		rs, err = c.raftNode[id].RequestDeleteNode(c.ClusterID, req.NodeID, 0, 3*time.Second)
	} else {
		return utils.GetJSONResponse(ctx, errors.New("unknown cmd "+req.CMD), nil)
	}
	if err != nil {
		return utils.GetJSONResponse(ctx, err, nil)
	}
	r := <-rs.AppliedC()
	if r.Completed() {
		return utils.GetJSONResponse(ctx, nil, "OK")
	}
	return utils.GetJSONResponse(ctx, errors.New("membership change failed"), nil)
}

// PreStop called before Stop()
func (c *RaftClusterComponent) PreStop(ctx context.Context) error {
	c.ctxCancel()
	for _, nodeID := range c.raftNodeID {
		if c.raftNode[nodeID] == nil {
			continue
		}
		rs, err := c.raftNode[nodeID].RequestDeleteNode(c.ClusterID, nodeID, 0, 3*time.Second)
		<-rs.AppliedC()
		if err != nil {
			c.logger.Error(err)
		}
		c.raftNode[nodeID].Stop()

	}
	return nil
}
