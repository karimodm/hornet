package gossip

import "github.com/gohornet/hornet/pkg/protocol"

func Fuzz(data []byte) int {
	p := protocol.New(nil)
	p.Receive(data)
	return 0
}
