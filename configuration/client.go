package configuration

import (
	"log"

	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/model"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

//NacosClient NacosClient
type NacosClient struct {
	Client       config_client.IConfigClient
	NamingClient naming_client.INamingClient
}

//NewNacos New Nacos Client
func NewNacos(clientConfig constant.ClientConfig, serverConfigs []constant.ServerConfig) (*NacosClient, error) {
	nacosClient := map[string]interface{}{
		"serverConfigs": serverConfigs,
		"clientConfig":  clientConfig,
	}
	// Create Config Center
	client, err := clients.CreateConfigClient(nacosClient)
	if err != nil {
		log.Println("CreateConfigClient: ", err)
		return nil, err
	}
	// Create Register Center
	namingClient, err := clients.CreateNamingClient(nacosClient)
	if err != nil {
		log.Println("CreateNamingClient: ", err)
		return nil, err
	}

	return &NacosClient{Client: client, NamingClient: namingClient}, nil
}

// Publish Config
func (c *NacosClient) Publish(dataID, group, content string) (bool, error) {
	ok, err := c.Client.PublishConfig(vo.ConfigParam{
		DataId:  dataID,
		Group:   group,
		Content: content,
	})
	if err != nil {
		return false, err
	}
	return ok, err
}

// Get Config
func (c *NacosClient) Get(dataID, group string) (string, error) {
	return c.Client.GetConfig(vo.ConfigParam{
		DataId: dataID,
		Group:  group,
	})
}

// Delete Config
func (c *NacosClient) Delete(dataID, group string) error {
	if _, err := c.Client.DeleteConfig(vo.ConfigParam{
		DataId: dataID,
		Group:  group,
	}); err != nil {
		return err
	}
	return nil
}

// Listen Config
func (c *NacosClient) Listen(dataID, group string, OnChange func(namespace, group, dataId, data string)) error {
	return c.Client.ListenConfig(vo.ConfigParam{
		DataId:   dataID,
		Group:    group,
		OnChange: OnChange,
	})
}

// Register Service
func (c *NacosClient) Register(param vo.RegisterInstanceParam) error {
	_, err := c.NamingClient.RegisterInstance(param)
	return err
}

// Deregister Service
func (c *NacosClient) Deregister(param vo.DeregisterInstanceParam) error {
	_, err := c.NamingClient.DeregisterInstance(param)
	return err
}

//SearchConfig SearchConfig
func (c *NacosClient) SearchConfig(search, dataID, group string, pageNo, pageSize int) (*model.ConfigPage, error) {
	configPage, err := c.Client.SearchConfig(vo.SearchConfigParam{
		Search:   search,
		DataId:   dataID,
		Group:    group,
		PageNo:   pageNo,
		PageSize: pageSize,
	})
	if err != nil {
		return nil, err
	}

	return configPage, nil
}

// GetService Service
func (c *NacosClient) GetService(param vo.GetServiceParam) (model.Service, error) {
	return c.NamingClient.GetService(param)
}

// SelectOneHealthyInstance Health Service
func (c *NacosClient) SelectOneHealthyInstance(serviceName string, clusters []string) (*model.Instance, error) {
	return c.NamingClient.SelectOneHealthyInstance(vo.SelectOneHealthInstanceParam{
		ServiceName: serviceName,
		Clusters:    clusters,
	})
}

// Subscribe the service change event
func (c *NacosClient) Subscribe(clusters []string, group, service string, cb func(services []model.SubscribeService, err error)) error {
	return c.NamingClient.Subscribe(&vo.SubscribeParam{
		GroupName:         group,
		Clusters:          clusters,
		ServiceName:       service,
		SubscribeCallback: cb,
	})
}
