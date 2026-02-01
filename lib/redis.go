package lib

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"reflect"
	"strconv"
	"time"
)

type redisConf struct {
	Name     string `mapstructure:"name"`
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	Db       int    `mapstructure:"db"`
}

// RedisManager Redis连接池管理器
type redisManager struct {
	client *redis.Client
	config *redisConf
	ctx    context.Context
	cancel context.CancelFunc
}

var redisMapPool map[string]*redisManager

func initRedis() error {
	var redisConf []redisConf
	if !IsSetConf("redis") {
		return errors.New("未配置redis")
	}
	err := viperConf.UnmarshalKey("redis", &redisConf)
	if err != nil {
		return err
	}
	redisMapPool = map[string]*redisManager{}
	for _, item := range redisConf {
		ctx, cancel := context.WithCancel(context.Background())

		manager := &redisManager{
			config: &item,
			ctx:    ctx,
			cancel: cancel,
		}
		err = manager.link()
		if err != nil {
			return err
		}
		redisMapPool[item.Name] = manager
	}
	return nil
}

func (m *redisManager) link() error {
	client := redis.NewClient(&redis.Options{
		Addr:         m.config.Addr,
		Password:     m.config.Password,
		DB:           m.config.Db,
		PoolSize:     100,
		MinIdleConns: 10,
		MaxConnAge:   time.Hour,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		IdleTimeout:  5 * time.Minute,
	})
	// 测试连接
	if err := client.Ping(context.Background()).Err(); err != nil {
		return fmt.Errorf("redis连接测试失败: %v", err)
	}
	m.client = client
	return nil
}

// IsHealthy 检查连接是否健康
func (m *redisManager) IsHealthy(ctx context.Context) bool {
	client := m.client
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(m.ctx, 2*time.Second)
		defer cancel()
	}
	return client.Ping(ctx).Err() == nil
}
func getRedisClient(name string) (*redisManager, error) {
	if redisClient, ok := redisMapPool[name]; ok {
		return redisClient, nil
	}
	return nil, errors.New("get redis error")
}
func closeRedis() error {
	for _, manager := range redisMapPool {
		manager.cancel()
		manager.client.Close()
	}
	redisMapPool = make(map[string]*redisManager)
	return nil
}

// SetStore 存储任意类型到Redis
func SetStore[T any](ctx context.Context, name, key string, data T, exp ...time.Duration) error {

	tx, err := getRedisClient(name)
	if err != nil {
		return err
	}

	// 根据类型决定如何存储
	var storeValue interface{}

	// 获取类型信息
	tType := reflect.TypeOf(data)

	switch tType.Kind() {
	case reflect.String:
		// 字符串直接存储
		storeValue = data
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// 整数转换为字符串
		storeValue = fmt.Sprintf("%d", reflect.ValueOf(data).Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		// 无符号整数转换为字符串
		storeValue = fmt.Sprintf("%d", reflect.ValueOf(data).Uint())
	case reflect.Float32, reflect.Float64:
		// 浮点数转换为字符串
		storeValue = fmt.Sprintf("%f", reflect.ValueOf(data).Float())
	case reflect.Bool:
		// 布尔值转换为字符串
		if reflect.ValueOf(data).Bool() {
			storeValue = "1"
		} else {
			storeValue = "0"
		}
	default:
		// 复杂类型使用JSON序列化
		jsonData, err := json.Marshal(data)
		if err != nil {
			return err
		}
		storeValue = string(jsonData)
	}

	// 设置过期时间（可选）
	if len(exp) > 0 && exp[0] > 0 {
		return tx.client.Set(ctx, key, storeValue, exp[0]).Err()
	}
	// 永久存储
	return tx.client.Set(ctx, key, storeValue, 0).Err()
}

// GetStore 从Redis获取任意类型数据
func GetStore[T any](ctx context.Context, name, key string) (T, error) {
	var zeroValue T
	tx, err := getRedisClient(name)
	if err != nil {
		return zeroValue, err
	}
	result, err := tx.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return zeroValue, nil
	} else if err != nil {
		return zeroValue, err
	}

	// 获取目标类型信息
	var target T
	targetType := reflect.TypeOf(target)

	// 根据类型进行解析
	return parseValueFromString[T](result, targetType)
}

// DefaultSetStore 设置默认
func DefaultSetStore[T any](ctx context.Context, key string, data T, exp ...time.Duration) error {
	return SetStore(ctx, "default", key, data, exp...)
}

// DefaultGetStore 获取值
func DefaultGetStore[T any](ctx context.Context, key string) (T, error) {
	return GetStore[T](ctx, "default", key)
}

// parseValueFromString 根据目标类型解析字符串
func parseValueFromString[T any](str string, targetType reflect.Type) (T, error) {
	var result T
	valuePtr := reflect.ValueOf(&result).Elem()

	switch targetType.Kind() {
	case reflect.String:
		// 字符串直接返回
		valuePtr.SetString(str)
		return result, nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// 解析整数
		intVal, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return result, fmt.Errorf("cannot parse '%s' as int: %v", str, err)
		}
		valuePtr.SetInt(intVal)
		return result, nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		// 解析无符号整数
		uintVal, err := strconv.ParseUint(str, 10, 64)
		if err != nil {
			return result, fmt.Errorf("cannot parse '%s' as uint: %v", str, err)
		}
		valuePtr.SetUint(uintVal)
		return result, nil

	case reflect.Float32, reflect.Float64:
		// 解析浮点数
		floatVal, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return result, fmt.Errorf("cannot parse '%s' as float: %v", str, err)
		}
		valuePtr.SetFloat(floatVal)
		return result, nil

	case reflect.Bool:
		// 解析布尔值
		boolVal := false
		if str == "1" || str == "true" || str == "TRUE" || str == "True" {
			boolVal = true
		} else if str != "0" && str != "false" && str != "FALSE" && str != "False" {
			// 尝试解析数字
			if intVal, err := strconv.Atoi(str); err == nil && intVal != 0 {
				boolVal = true
			}
		}
		valuePtr.SetBool(boolVal)
		return result, nil

	default:
		// 复杂类型使用JSON反序列化
		// 首先尝试直接JSON解析
		if err := json.Unmarshal([]byte(str), &result); err == nil {
			return result, nil
		}

		// 如果不是有效的JSON，尝试包装为JSON字符串再解析
		// 这适用于之前存储的纯字符串
		if isSimpleString(str) {
			// 对于简单字符串，尝试直接作为字符串字段
			return result, fmt.Errorf("stored value is plain string, but target type is %v", targetType)
		}

		return result, fmt.Errorf("cannot unmarshal '%s' to type %v", str, targetType)
	}
}

// isSimpleString 检查是否是简单字符串（不包含JSON结构）
func isSimpleString(str string) bool {
	// 检查是否以引号开头和结尾（可能是JSON字符串）
	if len(str) >= 2 && str[0] == '"' && str[len(str)-1] == '"' {
		return false
	}

	// 检查是否是JSON数组或对象
	if len(str) >= 2 && (str[0] == '[' && str[len(str)-1] == ']' ||
		str[0] == '{' && str[len(str)-1] == '}') {
		return false
	}

	return true
}
