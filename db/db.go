package db

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"

	"gorm.io/driver/mysql"     // gorm driver
	"gorm.io/driver/postgres"  // gorm driver
	"gorm.io/driver/sqlite"    // gorm driver
	"gorm.io/driver/sqlserver" // gorm driver

	// gorm driver
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	manager         *Manager
	once            sync.Once
	connStrings     []string
	adminConnString string
)

type Driver string

const (
	driverMySQL     Driver = "mysql"
	driverSQLServer Driver = "mssql"
	driverPostgres  Driver = "postgres"
	driverSQLLite   Driver = "sqllite"
)

const (
	defaultDriver          Driver = driverMySQL
	defaultConnMaxLifetime        = 0 // max connection life time in seconds
	defaultMaxIdleConns           = 2
	defaultMaxOpenConns           = 0
)

// Manager ...
type Manager struct {
	Driver          Driver
	ConnMaxLifetime time.Duration
	MaxIdleConns    int
	MaxOpenConns    int
	Admin           *gorm.DB
	Master          *gorm.DB
	Slaves          []*gorm.DB
	LogLevel        logger.LogLevel
}

// New returns singleton instance of DB manager
func New() *Manager {
	once.Do(func() {
		manager = setUpManager()
	})
	return manager
}

func setUpManager() *Manager {
	var maxLife int64
	var maxIdleConn int
	var maxOpenConn int
	var err error

	driverStr := os.Getenv("DB_DRIVER")
	var driver Driver
	if len(driverStr) == 0 {
		driver = defaultDriver
	} else {
		driver = Driver(driverStr)
	}

	envMaxLife := os.Getenv("DB_CONN_MAX_LIFETIME")
	if len(envMaxLife) == 0 {
		maxLife = defaultConnMaxLifetime
	} else {
		maxLife, err = strconv.ParseInt(envMaxLife, 10, 64)
		if err != nil {
			panic(err)
		}
	}

	envMaxIdleConn := os.Getenv("DB_MAX_IDLE_CONNS")
	if len(envMaxIdleConn) == 0 {
		maxIdleConn = defaultMaxIdleConns
	} else {
		maxIdleConn, err = strconv.Atoi(envMaxIdleConn)
		if err != nil {
			panic(err)
		}
	}

	envMaxOpenConn := os.Getenv("DB_MAX_OPEN_CONNS")
	if len(envMaxOpenConn) == 0 {
		maxOpenConn = defaultMaxOpenConns
	} else {
		maxOpenConn, err = strconv.Atoi(envMaxOpenConn)
		if err != nil {
			panic(err)
		}
	}

	return &Manager{
		Driver:          driver,
		ConnMaxLifetime: time.Second * time.Duration(maxLife),
		MaxIdleConns:    maxIdleConn,
		MaxOpenConns:    maxOpenConn,
	}
}

// AddAdminConnString can add admin connect string
func (m *Manager) AddAdminConnString(connString string) {
	adminConnString = connString
}

// AddConnString can add connect string. First string will be the master
func (m *Manager) AddConnString(connString string) {
	connStrings = append(connStrings, connString)
}

// AddConnStrings can add multiple connect strings. First string will be the master
func (m *Manager) AddConnStrings(connString []string) {
	connStrings = append(connStrings, connString...)
}

// Open will create DB instances
func (m *Manager) Open(opt ...gorm.Option) *Manager {
	if len(connStrings) == 0 {
		panic("There is no connect string for DB. Please add them using AddConnString method")
	}
	m.Master = m.open(connStrings[0], opt...)
	m.Slaves = make([]*gorm.DB, len(connStrings[1:]))
	for i, connString := range connStrings[1:] {
		m.Slaves[i] = m.open(connString, opt...)
	}
	return m
}

// OpenAdmin will create admin instance
func (m *Manager) OpenAdmin(opt ...gorm.Option) *Manager {
	if adminConnString == "" {
		panic("There is no admin connect string for DB. Please add them using AddAdminConnString method")
	}
	m.Admin = m.open(adminConnString, opt...)
	return m
}

// OpenMaster will create master instance
func (m *Manager) OpenMaster(opt ...gorm.Option) *Manager {
	if len(connStrings) == 0 {
		panic("There is no connect string for DB. Please add them using AddConnString method")
	}
	m.Master = m.open(connStrings[0], opt...)
	return m
}

// OpenSlaves will create slave instances
func (m *Manager) OpenSlaves(opt ...gorm.Option) *Manager {
	if len(connStrings) == 0 {
		panic("There is no connect string for DB. Please add them using AddConnString method")
	}
	m.Slaves = make([]*gorm.DB, len(connStrings))
	for i, connString := range connStrings {
		m.Slaves[i] = m.open(connString, opt...)
	}
	return m
}

func (m *Manager) open(connectString string, opt ...gorm.Option) *gorm.DB {
	var dialector gorm.Dialector

	switch m.Driver {
	case driverMySQL:
		dialector = mysql.Open(connectString)
	case driverSQLLite:
		dialector = sqlite.Open(connectString)
	case driverPostgres:
		dialector = postgres.Open(connectString)
	case driverSQLServer:
		dialector = sqlserver.Open(connectString)
	default:
		panic(fmt.Sprintf("invalid driver detected %s", m.Driver))
	}
	instance, err := gorm.Open(dialector, opt...)
	if err != nil {
		panic(err)
	}
	db, err := instance.DB()
	if err != nil {
		panic(err)
	}
	db.SetConnMaxLifetime(m.ConnMaxLifetime)
	db.SetMaxIdleConns(m.MaxIdleConns)
	db.SetMaxOpenConns(m.MaxOpenConns)
	return instance
}

// Close will release all DB instances
func (m *Manager) Close() {
	m.CloseMaster()
	m.CloseSlaves()
}

// CloseMaster will release all master instance
func (m *Manager) CloseAdmin() {
	db, _ := m.Admin.DB()
	db.Close()
}

// CloseMaster will release all master instance
func (m *Manager) CloseMaster() {
	db, _ := m.Master.DB()
	db.Close()
}

// CloseSlaves will release all slave instances
func (m *Manager) CloseSlaves() {
	for _, slv := range m.Slaves {
		db, _ := slv.DB()
		db.Close()
	}
}

// AdminConn will return admin connection
func (m *Manager) AdminConn() *gorm.DB {
	return m.Admin
}

// MasterConn will return master connection
func (m *Manager) MasterConn() *gorm.DB {
	return m.Master
}

// SlaveConn will return one of slave connection or master if all slave failed
func (m *Manager) SlaveConn() *gorm.DB {
	rand.Shuffle(len(m.Slaves), func(i, j int) { m.Slaves[i], m.Slaves[j] = m.Slaves[j], m.Slaves[i] })

	for _, slv := range m.Slaves {
		if _, err := slv.DB(); err == nil {
			return slv
		}
	}

	return m.MasterConn()
}
