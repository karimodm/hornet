package gossip

import (
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/pkg/errors"

	"github.com/iotaledger/hive.go/objectstorage"
	"github.com/iotaledger/hive.go/syncutils"

	"github.com/gohornet/hornet/pkg/model/storage"
	"github.com/gohornet/hornet/pkg/metrics"
	"github.com/gohornet/hornet/pkg/model/hornet"
	"github.com/gohornet/hornet/pkg/p2p"
)

// WorkUnitState defines the state which a WorkUnit is in.
type WorkUnitState byte

const (
	Hashing WorkUnitState = 1 << 0
	Invalid WorkUnitState = 1 << 1
	Hashed  WorkUnitState = 1 << 2
)

// newWorkUnit creates a new WorkUnit and initializes values by unmarshalling key.
func newWorkUnit(key []byte, serverMetrics *metrics.ServerMetrics) *WorkUnit {
	wu := &WorkUnit{
		receivedMsgBytes: make([]byte, len(key)),
		receivedFrom:     make([]*Protocol, 0),
		serverMetrics:    serverMetrics,
	}
	copy(wu.receivedMsgBytes, key)
	return wu
}

// CachedWorkUnit represents a cached WorkUnit.
type CachedWorkUnit struct {
	objectstorage.CachedObject
}

// WorkUnit gets the underlying WorkUnit.
func (c *CachedWorkUnit) WorkUnit() *WorkUnit {
	return c.Get().(*WorkUnit)
}

// WorkUnit defines the work around processing a received message and its
// associated requests from peers. There is at most one WorkUnit active per same
// message bytes.
type WorkUnit struct {
	objectstorage.StorableObjectFlags
	processingLock syncutils.Mutex

	serverMetrics *metrics.ServerMetrics

	// data
	dataLock         syncutils.RWMutex
	receivedMsgBytes []byte
	msg              *storage.Message

	// status
	stateLock syncutils.RWMutex
	state     WorkUnitState

	// received from
	receivedFromLock syncutils.RWMutex
	receivedFrom     []*Protocol
}

func (wu *WorkUnit) Update(_ objectstorage.StorableObject) {
	panic("WorkUnit should never be updated")
}

func (wu *WorkUnit) ObjectStorageKey() []byte {
	return wu.receivedMsgBytes
}

func (wu *WorkUnit) ObjectStorageValue() []byte {
	return nil
}

// UpdateState updates the WorkUnit's state.
func (wu *WorkUnit) UpdateState(state WorkUnitState) {
	wu.stateLock.Lock()
	wu.state = 0
	wu.state |= state
	wu.stateLock.Unlock()
}

// Is tells whether the WorkUnit has the given state.
func (wu *WorkUnit) Is(state WorkUnitState) bool {
	wu.stateLock.Lock()
	defer wu.stateLock.Unlock()
	return wu.state&state > 0
}

// adds a Request for the given peer to this WorkUnit.
// requestedMessageID can be nil to flag that this request just reflects a receive from the given
// peer and has no associated request.
func (wu *WorkUnit) addReceivedFrom(p *Protocol, requestedMessageID *hornet.MessageID) {
	wu.receivedFromLock.Lock()
	defer wu.receivedFromLock.Unlock()
	wu.receivedFrom = append(wu.receivedFrom, p)
}

// punishes, respectively increases the invalid message metric of all peers
// which sent the given underlying message of this WorkUnit.
// it also closes the connection to these peers.
func (wu *WorkUnit) punish(mng *p2p.Manager, reason error) {
	wu.receivedFromLock.Lock()
	defer wu.receivedFromLock.Unlock()
	for _, p := range wu.receivedFrom {
		wu.serverMetrics.InvalidMessages.Inc()

		// drop the connection to the peer
		_ = mng.DisconnectPeer(p.PeerID, errors.WithMessagef(reason, "peer was punished"))
	}
}

// builds a Broadcast where all peers which are associated with this WorkUnit are excluded from.
func (wu *WorkUnit) broadcast() *Broadcast {
	wu.receivedFromLock.Lock()
	defer wu.receivedFromLock.Unlock()
	exclude := map[peer.ID]struct{}{}
	for _, p := range wu.receivedFrom {
		exclude[p.PeerID] = struct{}{}
	}
	return &Broadcast{
		MsgData:      wu.receivedMsgBytes,
		ExcludePeers: exclude,
	}
}

// increases the known message metric of all peers
// except the given peer
func (wu *WorkUnit) increaseKnownTxCount(excludedPeer *Protocol) {
	wu.receivedFromLock.Lock()
	defer wu.receivedFromLock.Unlock()

	for _, p := range wu.receivedFrom {
		if p.PeerID == excludedPeer.PeerID {
			continue
		}
		wu.serverMetrics.KnownMessages.Inc()
		p.Metrics.KnownMessages.Inc()
	}
}
