package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"

	"github.com/kiga-hub/common/configuration"
)

func main() {
	//初始化nacos http://xxxxxxx:8848/nacos/
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
		log.Println("初始化nacos失败: ", newNacosErr)
		return
	}

	//2.服务注册 将需要接入的服务注册到nacos
	err := nacosClient.Register(vo.RegisterInstanceParam{
		Ip:          "127.0.0.1",                          //required
		Port:        10081,                                //required
		Weight:      10,                                   //required,it must be lager than 0
		Enable:      true,                                 //required,the instance can be access or not
		Healthy:     true,                                 //required,the instance is health or not
		Ephemeral:   true,                                 //optional
		Metadata:    map[string]string{"idc": "shanghai"}, //optional
		ClusterName: "cluster-a",                          //optional,default:DEFAULT
		ServiceName: "demo.go",                            //required
		GroupName:   "group-a",                            //optional,default:DEFAULT_GROUP
	})
	//判断是否成功
	if err != nil {
		log.Println("注册服务失败:", err)
		return
	}

	//获取初始配置 示例配置信息为`{{"AppName":"nacos-demo","IP":"192.168.8.230","Port":"8099"}}`
	config, err := nacosClient.Get("test-dataID", "test-group")
	if err != nil {
		log.Println("获取配置失败:", err)
		return
	}
	log.Println("configuration: ", config)
	//解析配置
	var i interface{}
	var configMap map[string]interface{}
	jsonErr := json.Unmarshal([]byte(config), &i)
	if jsonErr != nil {
		log.Println("解析配置失败:", jsonErr)
		return
	}
	configMap = i.(map[string]interface{})
	log.Println("configMap: ", configMap)

	//监听配置 可在匿名函数中处理监听到变化之后的业务逻辑
	listenErr := nacosClient.Listen("test-dataID", "test-group", func(namespace, group, dataId, data string) {
		log.Println("监听到修改配置:", data)
		//发现配置更改后解析
		var i interface{}
		var configMapNew map[string]interface{}
		jsonErr := json.Unmarshal([]byte(config), &i)
		if jsonErr != nil {
			log.Println("解析配置失败:", jsonErr)
			return
		}
		configMapNew = i.(map[string]interface{})
		log.Println("new configMap: ", configMapNew)
	})
	if listenErr != nil {
		log.Println("监听配置失败:", err)
		return
	}

	//3.启动一个http服务 (可忽略)
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
		fmt.Println("io.WriteString错误:", err)
	}
	fmt.Println(str)
}
