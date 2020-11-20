package main

import (
	"github.com/bCoder778/log"
	"net/http"
	_ "net/http/pprof"
	"qitmeer-sync/config"
	"qitmeer-sync/sync"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
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

	go http.ListenAndServe("0.0.0.0:8000", nil)
	sync, err := sync.NewQitmeerSync()
	if err != nil {
		log.Errorf("Create qitmeer sync failed! %v", err)
	}
	sync.Run()

}
