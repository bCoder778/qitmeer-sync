package rpc

import (
	"github.com/Qitmeer/qng/common/hash"
	"github.com/Qitmeer/qng/rpc/client"
	"github.com/bCoder778/qitmeer-sync/config"
)

func NewNotificationRpc(cfg *config.Ws, handle func(hash *hash.Hash, order int64, olds []*hash.Hash)) (*client.Client, error) {
	ntfnHandlers := &client.NotificationHandlers{
		OnReorganization: handle,
	}

	connCfg := &client.ConnConfig{
		Host:       cfg.Host,
		Endpoint:   "ws",
		User:       cfg.User,
		Pass:       cfg.Pass,
		DisableTLS: true,
	}

	client, err := client.New(connCfg, ntfnHandlers)
	if err != nil {
		return nil, err
	}

	// Register for block connect and disconnect notifications.
	if err := client.NotifyBlocks(); err != nil {
		return nil, err
	}

	return client, nil
}
