package conf

import (
	"reflect"
	"time"

	"github.com/bxcodec/faker/v3"
)

// TopologyConfig defines a topology
type TopologyConfig struct {
	// Version of the config
	Version string `json:"version,omitempty"`
	// CreateTime of the config
	CreateTime time.Time `json:"create_time,omitempty"`
	// ProjectName of the topology
	ProjectName string `json:"project_name,omitempty"`
	// Zones of the project
	Zones map[string]ZoneConfig `json:"zones"`
}

// ZoneConfig defines a zone
type ZoneConfig struct {
	// Name of the zone
	Name string `json:"name,omitempty"`
	// Nodes definition of the topology
	Nodes map[string]NodeConfig `json:"nodes"`
	// 网络设备
	// 链路
	// 采集设备
}

const (
	// SoftwareSystem is system type and version
	SoftwareSystem = "system"
	// SoftwareCaas is docker and containerd version
	SoftwareCaas = "caas"
	// SoftwarePlatform is platform version
	SoftwarePlatform = "platform"
	// SoftwareBasicBuzz is basical buzz version
	SoftwareBasicBuzz = "basic-buzz"
	// SoftwareVisualData is visual data version
	SoftwareVisualData = "visual-data"
	// SoftwareProject is project-related objects version
	SoftwareProject = "project"
)

// NodeConfig defines a node
type NodeConfig struct {
	// Name of the node
	Name string `json:"name,omitempty"`
	// Location of the node
	Location string `json:"location,omitempty"`
	// Hardware configuration of the node
	Hardware *HardwareConfig `json:"hardware,omitempty"`
	// PublicIPs of the node
	PublicIPs []string `json:"ips,omitempty"`
	// OverlayNetwork for the kiga overlay network
	OverlayNetwork string `json:"overlay_network,omitempty"`
	// GlobalCluster is a-b from OverlayNetwork a.b.c.0/24
	GlobalCluster string `json:"global_cluster,omitempty"`
	// Software version of every software part
	Software map[string]string `json:"software,omitempty"`
	// DataTransfer configuration
	DataTransfer *DataTransferConfig `json:"data_transfer,omitempty"`
	// APM configuration
	APM *APMConfig `json:"apm,omitempty"`
}

// HardwareConfig defines the hardware configuration of the node
type HardwareConfig struct {
	// Model of the hardware
	Model string `json:"model,omitempty"` //机型
	// CPU type
	CPU string `json:"cpu,omitempty"`
	// CPUCount of the node
	CPUCount int `json:"cpu_count,omitempty"`
	// MemoryGB of the node
	MemoryGB int `json:"memory_gb,omitempty"`
	// SystemGB is the space of the system disk
	SystemGB int `json:"system_gb,omitempty"`
	// DataGB is the space of the data space
	DataGB int `json:"data_gb,omitempty"`
	// DataLvmSetting is lvm settng of the data space
	DataLvmSetting []VolumeConfig `json:"data_lvm_setting,omitempty"`
	// NICCount is the count of the NIC
	NICCount int `json:"nic_count,omitempty"`
}

// VolumeConfig defines of volume in lvm
type VolumeConfig struct {
	// TotalUsefulSizeGB is the total size of volume
	TotalUsefulSizeGB int `json:"total_useful_size_gb,omitempty"`
	// DiskCount is number of disk in the volume
	DiskCount int `json:"disk_count,omitempty"`
	// PerDiskRealSizeGB is the space for each of the data disk
	PerDiskRealSizeGB int `json:"per_disk_size_gb,omitempty"`
	// RaidType of the data disks, -1 for none, 1 for raid1, 5 for raid5
	RaidType int `json:"raid_type,omitempty"` // -1为非raid
}

// APMConfig is the APM(Application Performance Managment) setting of a node
type APMConfig struct {
	// EnableLog or not
	EnableLog bool `json:"enable_log,omitempty"`
	// EnableTrace or not
	EnableTrace bool `json:"enable_trace,omitempty"`
	// EnableMetrics or not
	EnableMetrics bool `json:"enable_metrics,omitempty"`
}

// DataTransferConfig is the data transfer config of a node
type DataTransferConfig struct {
	// EnableDeviceControl or not
	EnableDeviceControl bool `json:"enable_device_control,omitempty"` //是否控制设备
	// EnableReceive or not
	EnableReceive bool `json:"enable_receive,omitempty"` //是否接收数据
	// ReceiveConfig for receive
	ReceiveConfig *DataTransferReceiveConfig `json:"receive_config,omitempty"`
	// EnableReadWrite or not
	EnableReadWrite bool `json:"enable_read_write,omitempty"` //是否存取数据
	// ReadWriteConfig for store
	ReadWriteConfig *DataTransferReadWriteConfig `json:"read_write_config,omitempty"`
	// EnableSend or not
	EnableSend bool `json:"enable_send,omitempty"` //是否外发数据
	// SendConfig for send
	SendConfig *DataTransferSendConfig `json:"send_config,omitempty"`
}

// SourceType stands for receive Source
type SourceType string

const (
	// SourceTypeFrame means frames from sensors
	SourceTypeFrame = SourceType("frame")
)

// DataTransferReceiveConfig is the receive part of data transfer config
type DataTransferReceiveConfig struct {
	// EnableCalculateTimestamp or not
	EnableCalculateTimestamp bool `json:"enable_calculate_timestamp,omitempty"` // 是否重新计算时间戳
	// EnableCRCCheck or not
	EnableCRCCheck bool `json:"enable_crc_check,omitempty"` // 是否校验CRC
	// Sources types and count
	//Sources map[SourceType]int64 `json:"sources,omitempty" faker:"sourcetype_int64_map"` //收什么，收多少(测点数)
}

// DataTransferReadWriteConfig is the store part of data transfer config
type DataTransferReadWriteConfig struct {
	// KeepDays keep data for N days, then may delete it when where is no space
	//KeepDays map[SourceType]int64 `json:"keep_days,omitempty" faker:"sourcetype_int64_map"` //存多久
}

// DataTransferSendConfig is the send part of data transfer config
type DataTransferSendConfig struct {
	// RemoteDeviceMgmtAddr -
	RemoteDeviceMgmtAddr string `json:"remote_device_mgmt_addr"` // 目标device地址
	// RemoteReceiverAddr -
	RemoteReceiverAddr string `json:"remote_receiver_addr"` // 目标receiver地址
	// Targets the address and send the type of data
	//Targets map[string]SourceType `json:"targets,omitempty" faker:"string_sourcetype_map"` //发给谁，发什么
}

// GenerateFakeTopology generate a fake topology
func GenerateFakeTopology() (*TopologyConfig, error) {
	err := faker.SetRandomMapAndSliceSize(5)
	if err != nil {
		return nil, err
	}
	err = faker.AddProvider("sourcetype_int64_map", func(v reflect.Value) (interface{}, error) {
		return map[SourceType]int64{
			"foo": 123,
			"bar": 345,
		}, nil
	})
	if err != nil {
		return nil, err
	}
	err = faker.AddProvider("string_sourcetype_map", func(v reflect.Value) (interface{}, error) {
		return map[string]SourceType{
			"ip1": SourceType("frame"),
			"ip2": SourceType("data"),
		}, nil
	})
	if err != nil {
		return nil, err
	}
	var topology TopologyConfig
	err = faker.FakeData(&topology)
	if err != nil {
		return nil, err
	}
	return &topology, err
}
