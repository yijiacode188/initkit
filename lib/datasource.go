package lib

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	// GORM 驱动
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/utils"

	// GORM Dialectors
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlserver"
)

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
	gormMapPool = map[string]*gorm.DB{}

	for _, item := range conf {
		// 根据不同的数据库驱动，采用不同的初始化方式
		var dbGorm *gorm.DB
		switch item.DriverName {
		case "mysql":
			dbGorm, err = initMySQL(item)

		case "sqlserver", "mssql":
			dbGorm, err = initSQLServer(item)

		default:
			err = fmt.Errorf("不支持的数据库驱动: %s", item.DriverName)
		}
		if err != nil {
			return fmt.Errorf("初始化数据库 %s 失败: %v", item.Name, err)
		}

		gormMapPool[item.Name] = dbGorm
	}
	return nil
}

// initMySQL 初始化 MySQL 连接
func initMySQL(conf datasourceConf) (*gorm.DB, error) {
	// 方式1：使用连接池
	dbPool, err := sql.Open("mysql", conf.DataSourceName)
	if err != nil {
		return nil, err
	}

	setConnectionPool(dbPool, conf)

	err = dbPool.Ping()
	if err != nil {
		dbPool.Close()
		return nil, err
	}

	dbGorm, err := gorm.Open(mysql.New(mysql.Config{Conn: dbPool}), &gorm.Config{
		Logger: &DefaultSqlGormLogger,
	})
	if err != nil {
		dbPool.Close()
		return nil, err
	}

	return dbGorm, nil
}

// initSQLServer 初始化 SQL Server 连接
func initSQLServer(conf datasourceConf) (*gorm.DB, error) {
	dbPool, err := sql.Open("sqlserver", conf.DataSourceName)
	if err != nil {
		return nil, err
	}

	setConnectionPool(dbPool, conf)

	err = dbPool.Ping()
	if err != nil {
		dbPool.Close()
		return nil, err
	}

	dbGorm, err := gorm.Open(sqlserver.New(sqlserver.Config{Conn: dbPool}), &gorm.Config{
		Logger: &DefaultSqlGormLogger,
	})
	if err != nil {
		dbPool.Close()
		return nil, err
	}

	return dbGorm, nil
}

// setConnectionPool 设置连接池参数
func setConnectionPool(dbPool *sql.DB, conf datasourceConf) {
	if conf.MaxOpenConn > 0 {
		dbPool.SetMaxOpenConns(conf.MaxOpenConn)
	}
	if conf.MaxIdleConn > 0 {
		dbPool.SetMaxIdleConns(conf.MaxIdleConn)
	}
	if conf.MaxConnLifeTime > 0 {
		dbPool.SetConnMaxLifetime(time.Duration(conf.MaxConnLifeTime) * time.Second)
	}
}

func GetGormPool(name string) (*gorm.DB, error) {
	if dbPool, ok := gormMapPool[name]; ok {
		return dbPool, nil
	}
	return nil, errors.New("get pool error")
}

// 获取底层的 sql.DB 用于手动管理连接
func GetSqlDB(name string) (*sql.DB, error) {
	if dbPool, ok := gormMapPool[name]; ok {
		sqlDB, err := dbPool.DB()
		if err != nil {
			return nil, err
		}
		return sqlDB, nil
	}
	return nil, errors.New("get pool error")
}

func closeDB() error {
	for _, dbGorm := range gormMapPool {
		if sqlDB, err := dbGorm.DB(); err == nil {
			sqlDB.Close()
		}
	}
	gormMapPool = make(map[string]*gorm.DB)
	return nil
}

// 日志部分保持不变
var DefaultSqlGormLogger = SqlGormLogger{
	LogLevel:      logger.Info,
	SlowThreshold: 200 * time.Millisecond,
}

type SqlGormLogger struct {
	LogLevel      logger.LogLevel
	SlowThreshold time.Duration
}

func (mgl *SqlGormLogger) LogMode(logLevel logger.LogLevel) logger.Interface {
	newLogger := *mgl
	newLogger.LogLevel = logLevel
	return &newLogger
}

func (mgl *SqlGormLogger) Info(ctx context.Context, message string, values ...interface{}) {
	if mgl.LogLevel < logger.Info {
		return
	}
	params := make(map[string]interface{})
	params["message"] = message
	params["values"] = fmt.Sprint(values)
	//Log.TagInfo(trace, "_com_mysql_Info", params)
}

func (mgl *SqlGormLogger) Warn(ctx context.Context, message string, values ...interface{}) {
	if mgl.LogLevel < logger.Warn {
		return
	}
	params := make(map[string]interface{})
	params["message"] = message
	params["values"] = fmt.Sprint(values)
	//Log.TagInfo(trace, "_com_mysql_Warn", params)
}

func (mgl *SqlGormLogger) Error(ctx context.Context, message string, values ...interface{}) {
	if mgl.LogLevel < logger.Error {
		return
	}
	params := make(map[string]interface{})
	params["message"] = message
	params["values"] = fmt.Sprint(values)
	//Log.TagInfo(trace, "_com_mysql_Error", params)
}

func (mgl *SqlGormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
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
	case mgl.LogLevel >= logger.Info:
		if rows == -1 {
			//Log.TagInfo(trace, "_com_mysql_success", msg)
		} else {
			msg["rows"] = rows
			//Log.TagInfo(trace, "_com_mysql_success", msg)
		}
	}
}
