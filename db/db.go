package db

import (
	"fmt"
	"qitmeer-sync/config"
	"qitmeer-sync/db/sqldb"
	"qitmeer-sync/storage"
)

func ConnectDB(setting *config.Config) (storage.IDB, error) {
	var (
		db  storage.IDB
		err error
	)
	switch setting.DB.DBType {
	case "mysql":
		if db, err = sqldb.ConnectMysql(setting.DB); err != nil {
			return nil, fmt.Errorf("failed to connect mysql, error:%v", err)
		}
	case "sqlserver":
		if db, err = sqldb.ConnectSqlServer(setting.DB); err != nil {
			return nil, fmt.Errorf("failed to connect mysql, error:%v", err)
		}
	default:
		return nil, fmt.Errorf("unsupported database %s", setting.DB.DBType)
	}
	return db, nil
}
