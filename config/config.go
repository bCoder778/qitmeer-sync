package config

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/bCoder778/log"
	"os"
	"sync"
)

var (
	ConfigFile = "config.toml"
)

var Setting *Config
var once sync.Once

func LoadConfig() error {
	if Exist(ConfigFile) {
		if _, err := toml.DecodeFile(ConfigFile, &Setting); err != nil {
			return err
		}
		return nil
	} else {
		return fmt.Errorf("%s not is exist", ConfigFile)
	}
}

type Config struct {
	Rpc       *Rpc       `toml:"rpc"`
	DB        *DB        `toml:"db"`
	Log       *Log       `toml:"log"`
	Email     *EMail     `toml:"email"`
	Verify    *Verify    `toml:"verify"`
	Resources *Resources `toml:"resources"`
}

type Rpc struct {
	Host     string `toml:"host"`
	Admin    string `toml:"admin"`
	Password string `toml:"password"`
}

type DB struct {
	DBType   string `toml:"dbtype"`
	User     string `toml:"user"`
	Password string `toml:"password"`
	Address  string `toml:"address"`
	DBName   string `toml:"dbname"`
}

type Resources struct {
	CPUNumber int `toml:"cpu-number"`
	GCPercent int `toml:"gc-percent"`
}

type Log struct {
	Mode  log.Mode  `toml:"mode"`
	Level log.Level `toml:"level"`
	Path  string    `toml:"path"`
}

type EMail struct {
	Title string   `toml:"title"`
	User  string   `toml:"user"`
	Pass  string   `toml:"pass"`
	Host  string   `toml:"host"`
	Port  string   `toml:"port"`
	To    []string `toml:"to"`
}

type Verify struct {
	UTXO     bool   `toml:"utxo"`
	Fees     bool   `toml:"fees"`
	Interval uint64 `toml:"interval"`
	Version  string `toml:"version"`
}

func Exist(fileName string) bool {
	_, err := os.Stat(fileName)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
