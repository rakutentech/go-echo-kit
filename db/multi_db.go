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
var SingletonDB *gorm.DB

type multiDbConf struct {
	Master string // master db dsn
	Slaves  []string // slave db dsn
	Name   string // db name
}

// return singleton DB
func ConnDB(conf[] multiDbConf) *gorm.DB {
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
		defaultDbName := conf[0].Name
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
			}, c.Name)
		}

		err = DB.Use(dbResolver)

		if err != nil {
			logger.LogCritf("[Fatal Error]can not connect to DB: %v", err)
		}
		SingletonDB = DB
	})

	return SingletonDB
}
