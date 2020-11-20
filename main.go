package main

import (
	"flag"
	"fmt"
	"github.com/bCoder778/log"
	"github.com/bCoder778/qitmeer-sync/config"
	"github.com/bCoder778/qitmeer-sync/sync"
	"github.com/bCoder778/qitmeer-sync/version"
	_ "net/http/pprof"
	"os"
	"runtime"
	"runtime/debug"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	debug.SetGCPercent(20)

	v := flag.Bool("v", false, "show bin info")
	flag.Parse()
	if *v {
		_, _ = fmt.Fprint(os.Stderr, version.StringifyMultiLine())
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
