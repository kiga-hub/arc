package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"

	"github.com/kiga-hub/arc/configuration"
)

func main() {
	// init nacos http://xxxxxxx:8848/nacos/
	clientConfig := constant.ClientConfig{
		NamespaceId:         "e525eafa-f7d7-4029-83d9-008937f9d468", //the namespaceId of Nacos.When namespace is public, fill in the blank string here.
		TimeoutMs:           3000,                                   //timeout for requesting Nacos server, default value is 10000ms
		NotLoadCacheAtStart: true,                                   //not to load persistent nacos service info in CacheDir at start time
		LogDir:              "/tmp/nacos/log",                       //the directory for log, default is current path
		CacheDir:            "/tmp/nacos/cache",                     //the directory for persist nacos service info,default value is current path
		LogLevel:            "debug",                                //the level of log, it's must be debug,info,warn,error, default value is info
	}
	serviceConfigs := []constant.ServerConfig{
		{
			IpAddr:      "192.168.8.234", //the nacos server address
			ContextPath: "/nacos",        //the nacos server contextPath
			Port:        8848,            //the nacos server port
		},
	}

	nacosClient, newNacosErr := configuration.NewNacos(clientConfig, serviceConfigs)
	if newNacosErr != nil {
		log.Println("nacos initialization failed: ", newNacosErr)
		return
	}

	//2.service register. register the serivice that needs to be accessed to nacos.
	err := nacosClient.Register(vo.RegisterInstanceParam{
		Ip:          "127.0.0.1",                          //required
		Port:        10081,                                //required
		Weight:      10,                                   //required,it must be larger than 0
		Enable:      true,                                 //required,the instance can be access or not
		Healthy:     true,                                 //required,the instance is health or not
		Ephemeral:   true,                                 //optional
		Metadata:    map[string]string{"idc": "shanghai"}, //optional
		ClusterName: "cluster-a",                          //optional,default:DEFAULT
		ServiceName: "demo.go",                            //required
		GroupName:   "group-a",                            //optional,default:DEFAULT_GROUP
	})
	// determine if it is successful.
	if err != nil {
		log.Println("service registered failed:", err)
		return
	}

	// get initialconfiguration. The example configuration information is {{"AppName":"nacos-demo","IP":"192.168.8.230","Port":"8099"}}`
	config, err := nacosClient.Get("test-dataID", "test-group")
	if err != nil {
		log.Println("get configuration failed:", err)
		return
	}
	log.Println("configuration: ", config)
	// decode configuiration
	var i interface{}
	var configMap map[string]interface{}
	jsonErr := json.Unmarshal([]byte(config), &i)
	if jsonErr != nil {
		log.Println("Parse failed:", jsonErr)
		return
	}
	configMap = i.(map[string]interface{})
	log.Println("configMap: ", configMap)

	// Listen to the configuration. you can handle the business logic after listening to the change in the anonymous function.
	listenErr := nacosClient.Listen("test-dataID", "test-group", func(namespace, group, dataId, data string) {
		log.Println("listen:", data)
		// Parse after discovering configuration changes
		var i interface{}
		var configMapNew map[string]interface{}
		jsonErr := json.Unmarshal([]byte(config), &i)
		if jsonErr != nil {
			log.Println("Parse failed:", jsonErr)
			return
		}
		configMapNew = i.(map[string]interface{})
		log.Println("new configMap: ", configMapNew)
	})
	if listenErr != nil {
		log.Println("Failed to listenf to the configuration:", err)
		return
	}

	//3.start a http service(can be ignored)
	ht := http.HandlerFunc(helloHandler)
	if ht != nil {
		http.Handle("/", ht)
	}
	err = http.ListenAndServe(":8099", nil)
	if err != nil {
		fmt.Println(" http.ListenAndServe失败:", err)
		return
	}
}

func helloHandler(w http.ResponseWriter, _ *http.Request) {
	str := "Hello world ! "
	_, err := io.WriteString(w, str)
	if err != nil {
		fmt.Println("io.WriteString failed:", err)
	}
	fmt.Println(str)
}
