package redis

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gomodule/redigo/redis"

	error2 "common/error"
)

// RXRedisCache  实现了RXCache
type RXRedisCache struct {
	sync.Mutex
	pool *redis.Pool
}

//RedisCache  redis缓存对象
// var RedisCache RXRedisCache

// InitRedisPool  初始化redis链接池
func InitRedisPool(config Config) (*RXRedisCache, error) {
	var err error
	redisPool := &redis.Pool{
		MaxIdle:     config.MaxIdle,
		MaxActive:   config.MaxActive,
		IdleTimeout: time.Second * time.Duration(config.IdleTimeout),
		Wait:        config.Wait,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", config.Address, redis.DialDatabase(config.DB))
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			if config.Password != "" {
				if _, err := c.Do("AUTH", config.Password); err != nil {
					if err := c.Close(); err != nil {
						fmt.Println("c.Close() err:", err)
					}
					fmt.Println(err)
					return nil, err
				}
			}
			return c, err
		},
	}
	redisCache := RXRedisCache{pool: redisPool}
	// just check redis connection
	_, err = redisCache.Get("")
	return &redisCache, err
}

// IncrBy reids 设置key 存储上线  超时时间
func (rc *RXRedisCache) IncrBy(key string, num int64, timeOut int) (interface{}, error) {
	conn := rc.pool.Get()
	defer func() {
		err := conn.Close()
		if err != nil {
			fmt.Println("IncrBy conn.Close() err:", err)
		}
	}()

	_, err := conn.Do("INCRBY", key, num)
	if err == nil {
		v, err := conn.Do("EXPIRE", key, timeOut)
		if err == nil {
			return v, nil
		}

		return nil, err

	}

	return nil, err
}

// DecrBy  redis  key 所储存的值减去指定的减量值
func (rc *RXRedisCache) DecrBy(key string, num int64) (interface{}, error) {
	conn := rc.pool.Get()
	defer func() {
		err := conn.Close()
		if err != nil {
			fmt.Println("DecrBy conn.Close() err:", err)
		}
	}()
	v, err := conn.Do("DECRBY", key, num)
	if err == nil {
		return v, nil
	}

	return nil, err

}

// Keys Redis Keys 命令用于查找所有符合给定模式 pattern 的 key 。。
func (rc *RXRedisCache) Keys(expression string) (interface{}, error) {
	conn := rc.pool.Get()
	defer func() {
		err := conn.Close()
		if err != nil {
			fmt.Println("Keys conn.Close() err:", err)
		}
	}()
	v, err := conn.Do("KEYS", expression)
	if err == nil {
		return v, nil
	}
	return nil, err

}

// Get Redis Get 命令用于获取指定 key 的值。如果 key 不存在，返回 nil 。如果key 储存的值不是字符串类型，返回一个错误。
func (rc *RXRedisCache) Get(key string) (interface{}, error) {
	conn := rc.pool.Get()

	defer func() {
		err := conn.Close()
		if err != nil {
			fmt.Println("Get conn.Close() err:", err)
		}
	}()
	v, err := conn.Do("GET", key)
	if err == nil {
		return v, nil
	}
	return nil, err

}

// GetString  Get 命令用于获取指定 key 的值 并转化为字符串
func (rc *RXRedisCache) GetString(key string) (string, error) {
	v, err := rc.Get(key)
	if v != nil && err == nil {
		return string(v.([]byte)), nil
	}
	return "", err

}

// GetObject redis获取对象
func (rc *RXRedisCache) GetObject(key string, ref interface{}) error {
	data, err := rc.Get(key)
	if data != nil && err == nil {
		err := json.Unmarshal(data.([]byte), ref)
		return err
	}
	return err

}

// Put redis 放入数据
func (rc *RXRedisCache) Put(key string, value interface{}) error {
	conn := rc.pool.Get()

	defer func() {
		err := conn.Close()
		if err != nil {
			fmt.Println("Put conn.Close() err:", err)
		}
	}()
	bytes, err := json.Marshal(value)
	if err == nil {
		_, err := conn.Do("SET", key, bytes)
		return err
	}
	return err
}

// PutWithExpire  Redis Setex 命令为指定的 key 设置值及其过期时间。如果 key 已经存在， SETEX 命令将会替换旧的值。
func (rc *RXRedisCache) PutWithExpire(key string, value interface{}, expire interface{}) error {
	bytes, err := json.Marshal(value)
	if err == nil {
		conn := rc.pool.Get()
		defer func() {
			err := conn.Close()
			if err != nil {
				fmt.Println("PutWithExpire conn.Close() err:", err)
			}
		}()
		_, err := conn.Do("SETEX", key, expire, bytes)
		return err
	}
	return err

}

// PutNX Redis Setnx（SET if Not eXists） 命令在指定的 key 不存在时，为 key 设置指定的值。
func (rc *RXRedisCache) PutNX(key string, value interface{}) (bool, error) {
	bytes, err := json.Marshal(value)

	if err == nil {
		conn := rc.pool.Get()
		defer func() {
			err := conn.Close()
			if err != nil {
				fmt.Println("PutNX conn.Close() err:", err)
			}
		}()
		reply, err := conn.Do("SETNX", key, bytes)
		fmt.Println(reply)
		return reply.(int64) == 1, err
	}

	return false, err
}

// LPush  Redis Lpush 命令将一个或多个值插入到列表头部。
func (rc *RXRedisCache) LPush(key string, value interface{}) error {
	conn := rc.pool.Get()
	defer func() {
		err := conn.Close()
		if err != nil {
			fmt.Println("LPush conn.Close() err:", err)
		}
	}()
	bytes, err := json.Marshal(value)
	if err == nil {
		_, err = conn.Do("LPUSH", key, bytes)
		return err
	}
	return err

}

// LPop Redis Lpop 命令用于移除并返回列表的第一个元素。
func (rc *RXRedisCache) LPop(key string) (interface{}, error) {
	conn := rc.pool.Get()
	defer func() {
		err := conn.Close()
		if err != nil {
			fmt.Println("LPop conn.Close() err:", err)
		}
	}()
	v, err := conn.Do("LPOP", key)
	if err == nil {
		return v, nil
	}
	return nil, err

}

// Delete Redis DEL 命令用于删除已存在的键。不存在的 key 会被忽略。
func (rc *RXRedisCache) Delete(key string) error {
	conn := rc.pool.Get()
	defer func() {
		err := conn.Close()
		if err != nil {
			fmt.Println("Delete conn.Close() err:", err)
		}
	}()
	_, err := conn.Do("DEL", key)
	return err
}

// IsExist Redis EXISTS 命令用于检查给定 key 是否存在。
func (rc *RXRedisCache) IsExist(key string) bool {
	conn := rc.pool.Get()
	defer func() {
		err := conn.Close()
		if err != nil {
			fmt.Println("IsExist conn.Close() err:", err)
		}
	}()
	v, err := redis.Bool(conn.Do("EXISTS", key))
	if err != nil {
		return false
	}
	return v
}

// LOCK reids 锁的回复
func (rc *RXRedisCache) LOCK(key, requestID string, timeOut time.Duration) error {
	conn := rc.pool.Get()
	defer func() {
		err := conn.Close()
		if err != nil {
			fmt.Println("LOCK conn.Close() err:", err)
		}
	}()
	lockReply, err := conn.Do("SET", key, requestID, "EX", timeOut, "NX")
	if err != nil {
		return error2.ErrRedisFail
	}

	if lockReply == "OK" {
		return nil
	}
	return error2.ErrRedisFail

}

// UNLOCK redis 取消锁
func (rc *RXRedisCache) UNLOCK(key, requestID string, timeOut time.Duration) error {
	var delScript = redis.NewScript(1, `if redis.call('get', KEYS[1]) == ARGV[1] then return redis.call('del', KEYS[1]) else return 0 end`)
	conn := rc.pool.Get()

	defer func() {
		err := conn.Close()
		if err != nil {
			fmt.Println("UNLOCK conn.Close() err:", err)
		}
	}()
	_, err := delScript.Do(conn, key, requestID)
	return err
}
