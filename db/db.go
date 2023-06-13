package db

import (
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"  // gorm driver
	_ "github.com/jinzhu/gorm/dialects/sqlite" // gorm driver
)

var (
	manager         *Manager
	once            sync.Once
	connStrings     []string
	adminConnString string
)

const (
	defaultDriver          = "mysql"
	defaultConnMaxLifetime = 0 // max connection life time in seconds
	defaultMaxIdleConns    = 2
	defaultMaxOpenConns    = 0
)

// Manager ...
type Manager struct {
	Driver          string
	ConnMaxLifetime time.Duration
	MaxIdleConns    int
	MaxOpenConns    int
	Admin           *gorm.DB
	Master          *gorm.DB
	Slaves          []*gorm.DB
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

	driver := os.Getenv("DB_DRIVER")
	if len(driver) == 0 {
		driver = defaultDriver
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

// AddConnStrings can add mulitple connect strings. First string will be the master
func (m *Manager) AddConnStrings(connString []string) {
	connStrings = append(connStrings, connString...)
}

// Open will create DB instances
func (m *Manager) Open() *Manager {
	if len(connStrings) == 0 {
		panic("There is no connect string for DB. Please add them using AddConnString method")
	}
	m.Master = m.open(connStrings[0])
	m.Slaves = make([]*gorm.DB, len(connStrings[1:]))
	for i, connString := range connStrings[1:] {
		m.Slaves[i] = m.open(connString)
	}
	return m
}

// OpenAdmin will create admin instance
func (m *Manager) OpenAdmin() *Manager {
	if adminConnString == "" {
		panic("There is no admin connect string for DB. Please add them using AddAdminConnString method")
	}
	m.Admin = m.open(adminConnString)
	return m
}

// OpenMaster will create master instance
func (m *Manager) OpenMaster() *Manager {
	if len(connStrings) == 0 {
		panic("There is no connect string for DB. Please add them using AddConnString method")
	}
	m.Master = m.open(connStrings[0])
	return m
}

// OpenSlaves will create slave instances
func (m *Manager) OpenSlaves() *Manager {
	if len(connStrings) == 0 {
		panic("There is no connect string for DB. Please add them using AddConnString method")
	}
	m.Slaves = make([]*gorm.DB, len(connStrings))
	for i, connString := range connStrings {
		m.Slaves[i] = m.open(connString)
	}
	return m
}

func (m *Manager) open(connectString string) *gorm.DB {
	instance, err := gorm.Open(m.Driver, connectString)
	if err != nil {
		panic(err)
	}
	instance.DB().SetConnMaxLifetime(m.ConnMaxLifetime)
	instance.DB().SetMaxIdleConns(m.MaxIdleConns)
	instance.DB().SetMaxOpenConns(m.MaxOpenConns)
	return instance
}

// SetLogMode will change SQL log mode (default: false)
func (m *Manager) SetLogMode(logMode bool) {
	m.Master.LogMode(logMode)
	for _, slv := range m.Slaves {
		slv.LogMode(logMode)
	}
}

// Close will release all DB instances
func (m *Manager) Close() {
	m.Master.Close()
	for _, slv := range m.Slaves {
		slv.Close()
	}
}

// CloseMaster will release all master instance
func (m *Manager) CloseMaster() {
	m.Master.Close()
}

// CloseSlaves will release all slave instances
func (m *Manager) CloseSlaves() {
	for _, slv := range m.Slaves {
		slv.Close()
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
	if l := len(m.Slaves); l > 0 {
		return m.Slaves[rand.Intn(l)]
	}

	return m.MasterConn()
}
