package main

import (
	"gitlab.mvk.com/go/vkgo/projects/donut/shooter/internal"
	"gitlab.mvk.com/go/vkgo/projects/donut/shooter/internal/action"
)

// --target-rps=1500 --shooting-duration=30m --avg-timing=40ms --stats-period=15s --rpc-path=https://localhost:8082
func main() {
	params := internal.LoadParameters()
	shooter := internal.Shooter{
		ShooterName: "donut-subscription get-content-access",
		Parameters:  params,
		Action:      action.NewGetContentAccess(params.ActorID, params.RpcPath, params.ErrorProbability),
	}

	shooter.Launch()
}
