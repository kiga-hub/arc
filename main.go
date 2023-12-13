package main

import (
	"fmt"

	"github.com/kiga-hub/arc/micro"
	"github.com/kiga-hub/arc/micro/component"
	"github.com/kiga-hub/arc/tracing"

	kafkaComponent "github.com/kiga-hub/arc/kafka"
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
			&kafkaComponent.Component{},
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
