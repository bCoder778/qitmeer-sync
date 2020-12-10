package main

import (
	"flag"
	"fmt"
	"github.com/bCoder778/log"
	"github.com/bCoder778/qitmeer-sync/config"
	"github.com/bCoder778/qitmeer-sync/db"
	"github.com/bCoder778/qitmeer-sync/sync"
	"github.com/bCoder778/qitmeer-sync/version"
	"os"
	"runtime"
	"runtime/debug"
)

func main() {
	setSystemResource()
	dealCommand()
	runSync()
}

func dealCommand() {
	v := flag.Bool("version", false, "show bin info")
	c := flag.Bool("clear", false, "clear data")
	conf := flag.String("config", "config.toml", "set config file")
	flag.Parse()

	if *v {
		_, _ = fmt.Fprint(os.Stderr, version.StringifyMultiLine())
		os.Exit(1)
	}
	if *c {
		clearData()
		os.Exit(1)
	}
	config.ConfigFile = *conf
}

func setSystemResource() {
	cpuNumber := runtime.NumCPU()
	gcPercent := 20
	if config.Setting != nil {
		if config.Setting.Resources.CPUNumber < cpuNumber {
			cpuNumber = config.Setting.Resources.CPUNumber
		}
		if config.Setting.Resources.GCPercent > 0 && config.Setting.Resources.GCPercent < 100 {
			gcPercent = config.Setting.Resources.GCPercent
		}

	}
	runtime.GOMAXPROCS(runtime.NumCPU())
	debug.SetGCPercent(gcPercent)
}

func runSync() {
	if err := config.LoadConfig(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	log.SetOption(&log.Option{
		LogLevel: config.Setting.Log.Level,
		Mode:     config.Setting.Log.Mode,
		Email: &log.EMailOption{
			User:   config.Setting.Email.User,
			Pass:   config.Setting.Email.Pass,
			Host:   config.Setting.Email.Host,
			Port:   config.Setting.Email.Port,
			Target: config.Setting.Email.To,
		},
	})

	sync, err := sync.NewQitmeerSync()
	if err != nil {
		log.Errorf("Create qitmeer sync failed! %v", err)
	}
	sync.Run()
}

func clearData() {
	fmt.Println("Are you sure you want to clear all data?(y/n)")
	readBytes := [10]byte{}
	_, err := os.Stdin.Read(readBytes[:])
	if err != nil {
		fmt.Println("Failed to read input, ", err.Error())
		os.Exit(1)
	}
	rs := string(readBytes[:1])
	switch rs {
	case "y":
		fallthrough
	case "Y":
		fmt.Println("Start to clear db data...")
		db, err := db.ConnectDB(config.Setting)
		if err != nil {
			fmt.Printf("Connect db filed! %s\n", err)
		}
		if err := db.Clear(); err != nil {
			fmt.Printf("Clear db failed! %s\n", err)
		}
		fmt.Println("Clear db success!")
		db.Close()
	}
}
