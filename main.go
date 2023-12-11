package main

import (
	"fmt"

	"github.com/kiga-hub/common/kafka"
	"github.com/kiga-hub/common/micro"
	"github.com/kiga-hub/common/micro/component"
	"github.com/kiga-hub/common/taos"
	"github.com/kiga-hub/common/tracing"
)

func main() {
	server, err := micro.NewServer(
		"demo",
		"v100",
		[]micro.IComponent{
			&component.LoggingComponent{},
			&tracing.Component{},
			&component.GossipKVCacheComponent{
				ClusterName:   "platform-global",
				Port:          6666,
				InMachineMode: false,
			},
			&kafka.Component{},
			&taos.Component{},
		},
	)
	if err != nil {
		panic(err)
	}
	err = server.Init()
	if err != nil {
		panic(err)
	}

	err = server.Run()
	if err != nil {
		fmt.Println(err)
	}
}
