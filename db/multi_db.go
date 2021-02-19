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

var onceMultiDb sync.Once
var Connection *gorm.DB

type MultiDbConf struct {
	Master  string // master db dsn
	Slaves  []string // slave db dsn array
	DbName  string // db name
}

// connect to multiple DB sources (mysql only)
func ConnMultiDB(conf[] MultiDbConf) *gorm.DB {
	if len(conf) == 0 {
		logger.LogCritf("[Fatal Error]can not connect to DB: empty dsn given")
	}

	onceMultiDb.Do(func() {
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
		Connection = DB
	})

	return Connection
}

// close database connection
func CloseMultiDB(dbConn *gorm.DB) {
	gormDB, err := dbConn.DB()
	if err != nil {
		logger.LogErrorf("[Error]can not get DB: %v", err)
	}

	err = gormDB.Close()
	if err != nil {
		logger.LogErrorf("[Error]can not close DB: %v", err)
	}
}

func GetInstance(connectionName string, isMaster bool) *gorm.DB {
	if isMaster {
		return Connection.Clauses(dbresolver.Use(connectionName), dbresolver.Write)
	} else {
		return Connection.Clauses(dbresolver.Use(connectionName), dbresolver.Read)
	}
}