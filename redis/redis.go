package redis

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gomodule/redigo/redis"

	error2 "github.com/kiga-hub/arc/error"
)

// RXRedisCache  implement RXCache
type RXRedisCache struct {
	sync.Mutex
	pool *redis.Pool
}

//RedisCache  redis cache object
// var RedisCache RXRedisCache

// InitRedisPool  init redis connection pool
//
//goland:noinspection GoUnusedExportedFunction
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

// IncrBy redis set key storage limit and timeout.
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

// DecrBy  redis  key The stored value minus the specified decrement
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

// Keys Redis Keys The command is used to find all keys that match the given pattern
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

// Get Redis Get The command is used to get the value of the specified key. If the key does not exist, return nil.
// If the value stored in the key is not a string type, return an error.
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

// GetString  Get The command is used to get the value of the specified key and convert it to a string
func (rc *RXRedisCache) GetString(key string) (string, error) {
	v, err := rc.Get(key)
	if v != nil && err == nil {
		return string(v.([]byte)), nil
	}
	return "", err

}

// GetObject redis get object
func (rc *RXRedisCache) GetObject(key string, ref interface{}) error {
	data, err := rc.Get(key)
	if data != nil && err == nil {
		err := json.Unmarshal(data.([]byte), ref)
		return err
	}
	return err

}

// Put redis insert data
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

// PutWithExpire  Redis Setex The command sets the value and expiration time for the specified key.
// If the key already exists, the SETEX command will replace the old value
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

// PutNX Redis Setnx（SET if Not exist）
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

// LPush  Redis Lpush Command to insert one or more values at the beginning of the list.
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

// LPop Redis Lpop Command to remove and return the first element of the list.
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

// Delete Redis DEL  Command to delete an existing key. Non-existent keys will be ignored.
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

// IsExist Redis EXISTS Command to check if the given key exists.
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

// LOCK redis Lock reply. Unlock."
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

// UNLOCK redis Unlock
func (rc *RXRedisCache) UNLOCK(key, requestID string, timeOut time.Duration) error {
	_ = timeOut
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
