# Go-Echo-Kit

![Build](https://github.com/rakutentech/go-echo-kit/workflows/go-echo-kit/badge.svg)

## Overview
Go Echo Kit provides useful tools for go-echo development.

1. Configuration with [Viper](https://github.com/spf13/viper)
2. DB connector with [Gorm](https://github.com/jinzhu/gorm)
3. Logger
4. Messages with [i18n](https://github.com/nicksnyder/go-i18n)

## Installation
### go get
```bash
go get -u github.com/rakutentech/go-echo-kit
```

## Database
### How to use it
#### DB manager
DB can be started like below
```go
import "github.com/rakutentech/go-echo-kit/db"

m := db.New()
m.AddConnString(connectStringMaster) // Add single connect string
m.AddConnStrings(connectStringSlaves) // Add mutliple connect string
m.Open()
defer m.Close()

m.MasterConn().Create(&user) // Connect to master
m.SlaveConn().Find(&user) // Connect to slave
```
You can add mulitple connect strings, and the first one will be the master.
SlaveConn will return one of slave randomly, and it can fail over to other slaves and master.

Regarding DB queries and how to generate connect string, please refer to [GORM guide](http://gorm.io/docs/index.html)

#### ConnectStringBuilder
This package also provides useful tool to build connection strings
```go
import "github.com/rakutentech/go-echo-kit/db"

builder := db.ConnStringBuilder{}
connStr := builder.
  SetFormat()
  SetHost("YourHost").
  SetUsername("YourUsername").
  SetPort("YourPort").
  SetPassword("YourPwd").
  SetOptions(optionMap)
  Build()
```
There is one more option to build string from config

```go
import "github.com/rakutentech/go-echo-kit/config"

cfg := config.New()
connStr := builder.SetWithConfig(cfg.Sub("databases.master")).Build())
```
Check testdata/config.yaml to get more details


### Enviroment variables
| Variable name        | Description                                | Default     |
| -----------          | -----------                                | ----------- |
| DB_DRIVER            | Db drivers which Gorm supports             | mysql       |
| DB_CONN_MAX_LIFETIME | Maximum connection life time(seconds)      | unlimited   |
| DB_MAX_IDLE_CONNS    | Maximum number of idle connection          | 2           |
| DB_MAX_OPEN_CONNS    | Maximum number of open connection          | unlimited   |

## Multiple DB connections (mysql only)
### How to use it
```go
import "github.com/rakutentech/go-echo-kit/db"
var dbConfig[] db.MultiDbConf

db1Config := db.MultiDbConf {
    Master: db1MasterDsn, // master db1 connection string
    Slaves: db1SlaveDsns, // salve db1 connection strings
    DbName: "db1",
}
db2Config := db.MultiDbConf {
    Master: db2MasterDsn, // master db2 connection string
    Slaves: db2SlaveDsns, // slave db2 connection strings
    DbName: "db2",
}

dbConfig = append(dbConfig, db1Config, db2Config)
dbConn = db.OpenDBConn(dbConfig)

defer db.CloseDBConn(dbConn)
```
```go
import "github.com/rakutentech/go-echo-kit/db"

query := db.GetInstance("db1", db.ConnTypeSlave).
		Select("user_id, name, email, created_at").
		Table("user").
		Limit(2)
```

## Configuration
### How to use it
Configuration can be started like below
```go
import "github.com/rakutentech/go-echo-kit/config"

cfg := config.New()
secret := cfg.GetString("app.secret")
```
Default config file type is `yaml`, but can be changed whatever viper supports.

#### Sample config file
```yaml
app:
  env: stg
  secret: ${APP_SECRET}
```
For ${APP_SECRET}, go-echo-kit will automatically find `.env`(dotenv) file from `CONFIG_PATH` or find it from environment variable.

#### Sample .env file
```
APP_SECRET=secret
DB_DRIVER=sqlite3
DB_PORT=4306
```

## Logger
### How to use it
Logger can be started like below
```go
import "github.com/rakutentech/go-echo-kit/logger"

logger.SetLogFile(filepath) // for simple use
logger.Notice("Your notice")
logger.Error("Your error")
logger.Crit("Your critical error that app should stop")

logger.SetRotatingLogFile(filepathPattern, options...) // For log rotation
```
Check [file-rotatelogs](https://github.com/lestrrat-go/file-rotatelogs) to get information about log rotation

#### Sample commands
```sh
# For test
cp config/.env.testing config/.env
go test ./...

# For build/run
cp config/.env.stg config/.env
go run main.go
```
You can create seprate dotenv files and replace them when you want.

### Enviroment variables
| Variable name      | Description                       | Default                     |
| -----------        | -----------                       | -----------                 |
| CONFIG_TYPE        | Config file type viper supports   | Yaml                        |
| CONFIG_PATH        | Where config file exists          | ./config(for local)         |

For more information about configuration, please refer to [Viper](https://github.com/spf13/viper)
