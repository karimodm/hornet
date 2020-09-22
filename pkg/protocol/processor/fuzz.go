package processor

import (
	"github.com/gohornet/hornet/pkg/compressed"
	"github.com/gohornet/hornet/pkg/profile"
	"github.com/iotaledger/iota.go/transaction"
)

func Fuzz(data []byte) int {

	opts := &Options{
		ValidMWM:          8,
		WorkUnitCacheOpts: profile.Profile8GB.Caches.IncomingTransactionFilter,
	}

	proc := New(nil, nil, opts)

	_wu, _, _ := WorkUnitFactory(data[0 : len(data)-1])
	wu := _wu.(*WorkUnit)
	// Random state, last byte of input
	wu.UpdateState(WorkUnitState(data[len(data)-1]))

	proc.ProcessWorkUnit(wu, nil)

	tx, err := compressed.TransactionFromCompressedBytes(wu.receivedTxBytes)
	if err != nil {
		if wu.Is(Hashed) {
			if !transaction.HasValidNonce(tx, opts.ValidMWM) {
				panic("WTF?")
			}
		}
	}

	if err != nil {
		return 1
	}

	return 0

}
