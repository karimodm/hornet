package processor

import (
	"github.com/gohornet/hornet/pkg/compressed"
	"github.com/gohornet/hornet/pkg/profile"
	peeringplugin "github.com/gohornet/hornet/plugins/peering"
	"github.com/iotaledger/iota.go/transaction"
)

func Fuzz(data []byte) int {

	opts := &Options{
		ValidMWM:          8,
		WorkUnitCacheOpts: profile.LoadProfile().Caches.IncomingTransactionFilter,
	}

	proc := New(nil, peeringplugin.Manager(), opts)

	_wu, _, _ := WorkUnitFactory(data[1:])
	wu := _wu.(*WorkUnit)
	wu.UpdateState(WorkUnitState(data[0]))

	proc.ProcessWorkUnit(wu, nil)

	tx, _ := compressed.TransactionFromCompressedBytes(wu.receivedTxBytes)
	if wu.Is(Hashed) {
		if !transaction.HasValidNonce(tx, opts.ValidMWM) {
			panic("WTF?")
		}
	}

	return 0

}
