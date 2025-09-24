package utils

import (
	"gitlab.mvk.com/go/vkgo/pkg/paas/rpcf"
	"gitlab.mvk.com/go/vkgo/pkg/paas/vklog"
)

func InitRpcManager(rpcPath string) rpcf.RpcManager {
	rpcManager, err := rpcf.NewManager(
		&rpcf.ManagerConfig{
			Connections: []string{rpcPath},
		},
		vklog.NewNoop(),
	)

	if err != nil {
		panic(err)
	}
	return rpcManager
}
