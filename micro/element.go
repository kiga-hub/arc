package micro

// ElementKey is type of key of element (not component)
type ElementKey string

// LoggingElementKey is ElementKey for logging
var LoggingElementKey = ElementKey("LoggingComponent")

// LoggerGroupElementKey is ElementKey for LoggerGroup
var LoggerGroupElementKey = ElementKey("LoggerGroupComponent")

// NacosClientElementKey is ElementKey for nacos client
var NacosClientElementKey = ElementKey("NacosClient")

// GossipKVCacheElementKey is ElementKey for GossipKVCache
var GossipKVCacheElementKey = ElementKey("GossipKVCacheComponent")
