package tangle

import (
	"github.com/gohornet/hornet/pkg/utils"
)

// measures the MPS values
func (t *Tangle) measureMPS() {
	incomingMsgCnt := t.serverMetrics.Messages.Load()
	incomingNewMsgCnt := t.serverMetrics.NewMessages.Load()
	outgoingMsgCnt := t.serverMetrics.SentMessages.Load()

	mpsMetrics := &MPSMetrics{
		Incoming: utils.GetUint32Diff(incomingMsgCnt, t.lastIncomingMsgCnt),
		New:      utils.GetUint32Diff(incomingNewMsgCnt, t.lastIncomingNewMsgCnt),
		Outgoing: utils.GetUint32Diff(outgoingMsgCnt, t.lastOutgoingMsgCnt),
	}

	// store the new counters
	t.lastIncomingMsgCnt = incomingMsgCnt
	t.lastIncomingNewMsgCnt = incomingNewMsgCnt
	t.lastOutgoingMsgCnt = outgoingMsgCnt

	// trigger events for outside listeners
	t.Events.MPSMetricsUpdated.Trigger(mpsMetrics)
}
