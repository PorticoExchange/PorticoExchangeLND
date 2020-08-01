package chain

import (
	"encoding/hex"
	"errors"
	"github.com/BoltzExchange/boltz-lnd/boltz"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
)

type BoltzChainClient struct {
	Boltz *boltz.Boltz
}

func (client BoltzChainClient) GetRawTransaction(transactionId string) (*wire.MsgTx, error) {
	transactionResponse, err := client.Boltz.GetTransaction(transactionId)

	if err != nil {
		return nil, err
	}

	transactionRaw, err := hex.DecodeString(transactionResponse.TransactionHex)

	if err != nil {
		return nil, err
	}

	transaction, err := btcutil.NewTxFromBytes(transactionRaw)

	if err != nil {
		return nil, err
	}

	return transaction.MsgTx(), nil
}

// TODO: implement
func (client BoltzChainClient) TransactionIsConfirmed(transactionId string) (bool, error) {
	return false, nil
}

func (client BoltzChainClient) BroadcastTransaction(transaction *wire.MsgTx) error {
	transactionHex, err := boltz.SerializeTransaction(transaction)

	if err != nil {
		return errors.New("could not serialize transaction: " + err.Error())
	}

	_, err = client.Boltz.BroadcastTransaction(transactionHex)

	if err != nil {
		return errors.New("could not broadcast transaction: " + err.Error())
	}

	return nil
}
