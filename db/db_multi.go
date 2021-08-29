package db

import (
	"github.com/rakutentech/go-echo-kit/logger"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"gorm.io/plugin/dbresolver"

	"log"
	"os"
	"sync"
	"time"
)

var onceGormDB sync.Once
var gormDB *gorm.DB

type ConnType int64

// MultiDbConf represents for config for 1 master DB and several slave DB
type MultiDbConf struct {
	Master  string // master db dsn
	Slaves  []string // slave db dsn array
	DbName  string // db name
}

// ConnTypeMaster ...
const ConnTypeMaster ConnType = 1

// ConnTypeSlave ...
const ConnTypeSlave ConnType = 0

// OpenDBConn connect to multiple DB sources (mysql only)
func OpenDBConn(conf[] MultiDbConf) *gorm.DB {
	if len(conf) == 0 {
		logger.LogCritf("[Fatal Error]can not connect to DB: empty dsn given")
	}

	onceGormDB.Do(func() {
		enableSqlLog := os.Getenv("SQL_LOGGER_ENABLED")

		gormConfig := &gorm.Config{}

		// print Slow SQL and happening errors
		if enableSqlLog == "true" {
			sqlLogger := gormlogger.New(
				log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
				gormlogger.Config{
					SlowThreshold: time.Second,   	  // Slow SQL threshold
					LogLevel:      gormlogger.Error, // Log level
					Colorful:      false,           // Disable color
				},
			)
			gormConfig = &gorm.Config{
				Logger: sqlLogger,
			}
		}

		/** default DB connection **/
		defaultMaster := conf[0].Master
		defaultDbName := conf[0].DbName
		DB, err := gorm.Open(mysql.Open(defaultMaster), gormConfig)

		if err != nil {
			logger.LogCritf("[Fatal Error]can not init default master DB: %v;dsn: %v", err, defaultMaster)
		}

		var defaultDialector []gorm.Dialector
		for _, defaultSlave := range conf[0].Slaves {
			defaultDialector = append(defaultDialector, mysql.Open(defaultSlave))
		}
		dbResolver := dbresolver.Register(dbresolver.Config{
			Replicas: defaultDialector,
		}, defaultDbName)

		/** default DB connection end **/

		/** connect from other sources **/
		for idx, c := range conf {
			if idx == 0 {
				continue
			}

			var dialector []gorm.Dialector
			for _, slave := range c.Slaves {
				dialector = append(dialector, mysql.Open(slave))
			}

			dbResolver.Register(dbresolver.Config{
				Sources:  []gorm.Dialector{mysql.Open(c.Master)},
				Replicas: dialector,
			}, c.DbName)
		}

		err = DB.Use(dbResolver)

		if err != nil {
			logger.LogCritf("[Fatal Error]can not connect to DB: %v", err)
		}
		gormDB = DB
	})

	return gormDB
}

// CloseDBConn close database connection
func CloseDBConn(dbConn *gorm.DB) {
	gormDB, err := dbConn.DB()
	if err != nil {
		logger.LogErrorf("[Error]can not get gormDB: %v", err)
	}

	err = gormDB.Close()
	if err != nil {
		logger.LogErrorf("[Error]can not close gormDB: %v", err)
	}
}

// GetConn get master or slave connection from DB i
func GetConn(DBName string, connType ConnType) *gorm.DB {
	appDebug := os.Getenv("APP_DEBUG")
	operation := dbresolver.Read; if connType == ConnTypeMaster {
		operation = dbresolver.Write
	}

	if appDebug =="true" {
		return gormDB.Clauses(dbresolver.Use(DBName), operation).Debug()
	}
	return gormDB.Clauses(dbresolver.Use(DBName), operation)
}