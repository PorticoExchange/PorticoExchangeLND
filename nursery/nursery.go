package nursery

import (
	"github.com/BoltzExchange/boltz-lnd/boltz"
	"github.com/BoltzExchange/boltz-lnd/database"
	"github.com/BoltzExchange/boltz-lnd/lnd"
	"github.com/BoltzExchange/boltz-lnd/logger"
	"github.com/BoltzExchange/boltz-lnd/scrooge"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/lightningnetwork/lnd/lnrpc/chainrpc"
	"sync"
)

type Nursery struct {
	symbol      string
	boltzPubKey string

	chainParams *chaincfg.Params

	lnd      *lnd.LND
	boltz    *boltz.Boltz
	database *database.Database

	scrooge *scrooge.Scrooge
}

// Map between Swap ids and a channel that tells its SSE event listeners to stop
var eventListeners = make(map[string]chan bool)
var eventListenersLock sync.RWMutex

func (nursery *Nursery) Init(
	symbol string,
	boltzPubKey string,
	chainParams *chaincfg.Params,
	lnd *lnd.LND,
	boltz *boltz.Boltz,
	database *database.Database,
	scrooge *scrooge.Scrooge,
) error {
	nursery.symbol = symbol
	nursery.boltzPubKey = boltzPubKey

	nursery.chainParams = chainParams

	nursery.lnd = lnd
	nursery.boltz = boltz
	nursery.database = database

	nursery.scrooge = scrooge

	logger.Info("Starting nursery")

	// TODO: use channel acceptor to prevent invalid channel openings from happening

	blockNotifier := make(chan *chainrpc.BlockEpoch)
	err := nursery.lnd.RegisterBlockListener(blockNotifier)

	if err != nil {
		return err
	}

	err = nursery.recoverSwaps(blockNotifier)

	if err != nil {
		return err
	}

	err = nursery.recoverReverseSwaps()

	return err
}
