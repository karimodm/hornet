package spammer

import (
	"fmt"
	"time"

	iotago "github.com/iotaledger/iota.go"

	"github.com/gohornet/hornet/pkg/metrics"
	"github.com/gohornet/hornet/pkg/model/hornet"
	"github.com/gohornet/hornet/pkg/model/storage"
	"github.com/gohornet/hornet/pkg/pow"
)

// SendMessageFunc is a function which sends a message to the network.
type SendMessageFunc = func(msg *storage.Message) error

// SpammerTipselFunc selects tips for the spammer.
type SpammerTipselFunc = func() (isSemiLazy bool, tips hornet.MessageIDs, err error)

// Spammer is used to issue messages to the IOTA network to create load on the tangle.
type Spammer struct {

	// config options
	message         string
	index           string
	indexSemiLazy   string
	tipselFunc      SpammerTipselFunc
	powHandler      *pow.Handler
	sendMessageFunc SendMessageFunc
	serverMetrics   *metrics.ServerMetrics
}

// New creates a new spammer instance.
func New(message string, index string, indexSemiLazy string, tipselFunc SpammerTipselFunc, powHandler *pow.Handler, sendMessageFunc SendMessageFunc, serverMetrics *metrics.ServerMetrics) *Spammer {

	return &Spammer{
		message:         message,
		index:           index,
		indexSemiLazy:   indexSemiLazy,
		tipselFunc:      tipselFunc,
		powHandler:      powHandler,
		sendMessageFunc: sendMessageFunc,
		serverMetrics:   serverMetrics,
	}
}

func (s *Spammer) DoSpam(shutdownSignal <-chan struct{}) (time.Duration, time.Duration, error) {

	timeStart := time.Now()
	isSemiLazy, tips, err := s.tipselFunc()
	if err != nil {
		return time.Duration(0), time.Duration(0), err
	}
	durationGTTA := time.Since(timeStart)

	indexation := s.index
	if isSemiLazy {
		indexation = s.indexSemiLazy
	}

	txCount := int(s.serverMetrics.SentSpamMessages.Load()) + 1

	now := time.Now()
	messageString := s.message
	messageString += fmt.Sprintf("\nCount: %06d", txCount)
	messageString += fmt.Sprintf("\nTimestamp: %s", now.Format(time.RFC3339))
	messageString += fmt.Sprintf("\nTipselection: %v", durationGTTA.Truncate(time.Microsecond))

	iotaMsg := &iotago.Message{Version: 1, Parent1: *tips[0], Parent2: *tips[1], Payload: &iotago.Indexation{Index: indexation, Data: []byte(messageString)}}

	timeStart = time.Now()
	if err := s.powHandler.DoPoW(iotaMsg, shutdownSignal, 1); err != nil {
		return time.Duration(0), time.Duration(0), err
	}
	durationPOW := time.Since(timeStart)

	msg, err := storage.NewMessage(iotaMsg, iotago.DeSeriModePerformValidation)
	if err != nil {
		return time.Duration(0), time.Duration(0), err
	}

	if err := s.sendMessageFunc(msg); err != nil {
		return time.Duration(0), time.Duration(0), err
	}

	return durationGTTA, durationPOW, nil
}
