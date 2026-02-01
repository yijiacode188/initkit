package lib

import (
	"errors"
	"fmt"
	"github.com/sony/sonyflake/v2"
)

var sf *sonyflake.Sonyflake

func initSnowflake() error {

	settings := sonyflake.Settings{
		MachineID: func() (int, error) {
			return int(ipv4ToUint16(localIP)), nil
		},
	}
	var err error
	sf, err = sonyflake.New(settings)
	if err != nil {
		return errors.New(fmt.Sprintf("sonyflake not created:error %s", err.Error()))

	}
	return nil
}

// GetSnowflakeID 获取雪花id
func GetSnowflakeID() int64 {
	id, err := sf.NextID()
	if err != nil {
		panic(fmt.Sprintf("generate id failed: %v", err))
	}
	return id
}
