// +build gofuzz

package gossip

import (
	"fmt"

	"github.com/gohornet/hornet/pkg/config"
	"github.com/iotaledger/hive.go/logger"
)

func Fuzz(data []byte) int {
	/*
				var bogusCfg *viper.Viper
				bogusCfg = viper.New()
				bogusCfg.Set(config.CfgCoordinatorAddress, "UDYXTZBE9GZGPM9SSQV9LTZNDLJIZMPUVVXYXFYVBLIEUHLSEWFTKZZLXYRHHWVQV9MNNX9KZC9D9UZWZ")
				bogusCfg.Set(config.CfgCoordinatorMWM, 14)
			  bogusCfg.Set(config.CfgNetGossipBindAddress, "0.0.0.0:15600")

		//cli.ParseConfig()
		//logger.InitGlobalLogger(bogusCfg)
	*/
	fmt.Println(data)
	config.FetchConfig()
	logger.InitGlobalLogger(config.NodeConfig)
	plugin := PLUGIN // My own gossip plugin
	PLUGIN.Events.Configure.Trigger(plugin)
	PLUGIN.Events.Run.Trigger(plugin)
	/*
		  	proc.wp = workerpool.New(func(task workerpool.Task) {
				p := task.Param(0).(*peer.Peer)
				data := task.Param(2).([]byte)

				switch task.Param(1).(message.Type) {
				case sting.MessageTypeTransaction:
					proc.processTransaction(p, data)
				case sting.MessageTypeTransactionRequest:
					proc.processTransactionRequest(p, data)
				case sting.MessageTypeMilestoneRequest:
					proc.processMilestoneRequest(p, data)
				}
	*/
	return 0
}
