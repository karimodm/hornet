package gossip

import (
	"github.com/gohornet/hornet/pkg/protocol/sting"
)

func Fuzz(data []byte) int {
	_ = sting.ParseHeartbeat(data)
	return 0
}
