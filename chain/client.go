package chain

import (
	"github.com/btcsuite/btcd/wire"
)

type Client interface {
	GetRawTransaction(transactionId string) (*wire.MsgTx, error)
	TransactionIsConfirmed(transactionId string) (bool, error)

	BroadcastTransaction(transaction *wire.MsgTx) error
}
