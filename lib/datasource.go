package lib

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
	"time"
)

var dbMapPool map[string]*sql.DB
var gormMapPool map[string]*gorm.DB

type datasourceConf struct {
	Name            string `mapstructure:"name"`
	DriverName      string `mapstructure:"driver_name"`
	DataSourceName  string `mapstructure:"data_source_name"`
	MaxOpenConn     int    `mapstructure:"max_open_conn"`
	MaxIdleConn     int    `mapstructure:"max_idle_conn"`
	MaxConnLifeTime int    `mapstructure:"max_conn_life_time"`
}

// InitDatasource 初始化数据库
func initDatasource() error {
	var conf []datasourceConf

	if !IsSetConf("datasource") {
		return errors.New("未配置数据库")
	}
	err := viperConf.UnmarshalKey("datasource", &conf)
	if err != nil {
		return err
	}
	dbMapPool = map[string]*sql.DB{}
	gormMapPool = map[string]*gorm.DB{}
	for _, item := range conf {

		dbPool, err := sql.Open(item.DriverName, item.DataSourceName)
		if err != nil {
			return err
		}
		dbPool.SetMaxOpenConns(item.MaxOpenConn)
		dbPool.SetMaxIdleConns(item.MaxIdleConn)
		dbPool.SetConnMaxLifetime(time.Duration(item.MaxConnLifeTime) * time.Second)
		err = dbPool.Ping()
		if err != nil {
			return err
		}

		if item.DriverName == "mysql" {

			//gorm连接方式
			dbGorm, err := gorm.Open(mysql.New(mysql.Config{Conn: dbPool}), &gorm.Config{
				Logger: &DefaultSqlGormLogger,
			})
			if err != nil {
				return err
			}
			dbMapPool[item.Name] = dbPool
			gormMapPool[item.Name] = dbGorm
		}
	}
	return nil
}
func GetDBPool(name string) (*sql.DB, error) {
	if dbPool, ok := dbMapPool[name]; ok {
		return dbPool, nil
	}
	return nil, errors.New("get pool error")
}

func GetGormPool(name string) (*gorm.DB, error) {
	if dbPool, ok := gormMapPool[name]; ok {
		return dbPool, nil
	}
	return nil, errors.New("get pool error")
}

func closeDB() error {
	for _, dbPool := range dbMapPool {
		dbPool.Close()
	}
	dbMapPool = make(map[string]*sql.DB)
	gormMapPool = make(map[string]*gorm.DB)
	return nil
}

// mysql 日志打印类型

var DefaultSqlGormLogger = SqlGormLogger{
	LogLevel:      logger.Info,
	SlowThreshold: 200 * time.Millisecond,
}

type SqlGormLogger struct {
	LogLevel      logger.LogLevel
	SlowThreshold time.Duration
}

func (mgl *SqlGormLogger) LogMode(logLevel logger.LogLevel) logger.Interface {
	mgl.LogLevel = logLevel
	return mgl
}

func (mgl *SqlGormLogger) Info(ctx context.Context, message string, values ...interface{}) {
	//trace := GetTraceContext(ctx)
	params := make(map[string]interface{})
	params["message"] = message
	params["values"] = fmt.Sprint(values)
	//Log.TagInfo(trace, "_com_mysql_Info", params)
}
func (mgl *SqlGormLogger) Warn(ctx context.Context, message string, values ...interface{}) {
	//trace := GetTraceContext(ctx)
	params := make(map[string]interface{})
	params["message"] = message
	params["values"] = fmt.Sprint(values)
	//Log.TagInfo(trace, "_com_mysql_Warn", params)
}
func (mgl *SqlGormLogger) Error(ctx context.Context, message string, values ...interface{}) {
	//trace := GetTraceContext(ctx)
	params := make(map[string]interface{})
	params["message"] = message
	params["values"] = fmt.Sprint(values)
	//Log.TagInfo(trace, "_com_mysql_Error", params)
}
func (mgl *SqlGormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	//trace := GetTraceContext(ctx)

	if mgl.LogLevel <= logger.Silent {
		return
	}

	sqlStr, rows := fc()
	currentTime := begin.Format(timeFormat)
	elapsed := time.Since(begin)
	msg := map[string]interface{}{
		"FileWithLineNum": utils.FileWithLineNum(),
		"sql":             sqlStr,
		"rows":            "-",
		"proc_time":       float64(elapsed.Milliseconds()),
		"current_time":    currentTime,
	}
	switch {
	case err != nil && mgl.LogLevel >= logger.Error && (!errors.Is(err, logger.ErrRecordNotFound)):
		msg["err"] = err
		if rows == -1 {
			//Log.TagInfo(trace, "_com_mysql_failure", msg)
		} else {
			msg["rows"] = rows
			//Log.TagInfo(trace, "_com_mysql_failure", msg)
		}
	case elapsed > mgl.SlowThreshold && mgl.SlowThreshold != 0 && mgl.LogLevel >= logger.Warn:
		slowLog := fmt.Sprintf("SLOW SQL >= %v", mgl.SlowThreshold)
		msg["slowLog"] = slowLog
		if rows == -1 {
			//Log.TagInfo(trace, "_com_mysql_success", msg)
		} else {
			msg["rows"] = rows
			//Log.TagInfo(trace, "_com_mysql_success", msg)
		}
	case mgl.LogLevel == logger.Info:
		if rows == -1 {
			//Log.TagInfo(trace, "_com_mysql_success", msg)
		} else {
			msg["rows"] = rows
			//Log.TagInfo(trace, "_com_mysql_success", msg)
		}
	}
}
