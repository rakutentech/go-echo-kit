package config

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

var (
	once       sync.Once
	instance   *viper.Viper
	subs       map[string]*viper.Viper
	configPath string
	configType string
	appEnv     string
)

const (
	dafultConfigPath = "./config"
	dafultConfigType = "yaml"
	dotEnvFileName   = ".env"
)

// New returns singleton instance of config
func New() *viper.Viper {
	once.Do(func() {
		instance = new()
	})

	return instance
}

func new() *viper.Viper {
	viper := viper.New()

	// Config file merge phase
	setUpViper(viper)

	for _, f := range getFileList() {
		fileName := f.Name()
		if isTargetConfigFile(fileName) {
			mergeConfigFile(viper, fileName)
		}
	}

	// Dotenv and os env merge phase
	loadEnv()

	for _, key := range viper.AllKeys() {
		mergeEnv(key, viper)
	}

	return viper
}

func setUpViper(viper *viper.Viper) {
	configPath = os.Getenv("CONFIG_PATH")
	if len(configPath) == 0 {
		configPath = dafultConfigPath
	}

	configType = os.Getenv("CONFIG_TYPE")
	if len(configType) == 0 {
		configType = dafultConfigType
	}

	viper.SetConfigType(configType)
	viper.AddConfigPath(configPath)
}

func isTargetConfigFile(fileName string) bool {
	// ignore dotenv and find files by extension
	return !strings.HasPrefix(fileName, ".") && filepath.Ext(fileName) == "."+configType
}

func mergeConfigFile(viper *viper.Viper, fileName string) {
	viper.SetConfigName(strings.TrimSuffix(fileName, filepath.Ext(fileName)))
	if err := viper.MergeInConfig(); err != nil {
		panic(err)
	}
}

func loadEnv() {
	var err error
	filePath := filepath.Join(configPath, dotEnvFileName)
	err = godotenv.Load(filePath)
	if err != nil {
		log.Print("dotenv file not found, skipping dotenv setting")
		return
	}
}

func mergeEnv(key string, viper *viper.Viper) {
	value := viper.GetString(key)
	if needEnvVariable(value) {
		viper.Set(key, getEnv(extractValue(value)))
	}
}

func needEnvVariable(value string) bool {
	return strings.HasPrefix(value, "${") && strings.HasSuffix(value, "}")
}

func getEnv(env string) string {
	return os.Getenv(env)
}

func extractValue(value string) string {
	return strings.TrimSuffix(strings.TrimPrefix(value, "${"), "}")
}

func getFileList() []os.FileInfo {
	files, err := ioutil.ReadDir(configPath)
	if err != nil {
		panic(err)
	}
	return files
}

// CreateAlias creates alias config
func CreateAlias(name string, prefix string) {
	cfg := New()
	if subs == nil {
		subs = make(map[string]*viper.Viper)
	}
	subs[name] = cfg.Sub(prefix)
}

// Alias returns singleton instance of aliased config
func Alias(name string) *viper.Viper {
	return subs[name]
}

// Refresh recreate config instance
func Refresh() {
	instance = new()
	subs = nil
}
