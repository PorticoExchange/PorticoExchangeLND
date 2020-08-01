package scrooge

import (
	"fmt"
	"github.com/BoltzExchange/boltz-lnd/boltz"
	"github.com/BoltzExchange/boltz-lnd/chain"
	"github.com/BoltzExchange/boltz-lnd/database"
	"github.com/BoltzExchange/boltz-lnd/lnd"
	"github.com/BoltzExchange/boltz-lnd/utils"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"math"
)

// TODO: handle race conditions in which transactions confirm while a new one is constructed
// TODO: better logging in all of scrooge
// TODO: handle unconfirmed inputs
// TODO: add async lock to avoid race conditions

type Scrooge struct {
	chainParams *chaincfg.Params

	database *database.Database

	lnd         *lnd.LND
	chainClient chain.Client
}

func (scrooge *Scrooge) Init(
	chainParams *chaincfg.Params,
	database *database.Database,
	lnd *lnd.LND,
	chainClient chain.Client,
) {
	scrooge.chainParams = chainParams

	scrooge.database = database

	scrooge.lnd = lnd
	scrooge.chainClient = chainClient
}

func (scrooge *Scrooge) SendRefundTransaction(swapsToRefund []database.Swap) (*wire.MsgTx, error) {
	unconfirmedTransaction, swaps, reverseSwaps, err := scrooge.getUnconfirmedSwaps()

	if err != nil {
		return nil, err
	}

	swaps = append(swaps, swapsToRefund...)

	return scrooge.batchSwapTransactions(unconfirmedTransaction, swaps, reverseSwaps)
}

func (scrooge Scrooge) SendClaimTransaction(reverseSwap database.ReverseSwap) (*wire.MsgTx, error) {
	unconfirmedTransaction, swaps, reverseSwaps, err := scrooge.getUnconfirmedSwaps()

	if err != nil {
		return nil, err
	}

	reverseSwaps = append(reverseSwaps, reverseSwap)

	return scrooge.batchSwapTransactions(unconfirmedTransaction, swaps, reverseSwaps)
}

func (scrooge *Scrooge) batchSwapTransactions(
	unconfirmedTransaction *database.UnconfirmedTransaction,
	swaps []database.Swap,
	reverseSwaps []database.ReverseSwap,
) (*wire.MsgTx, error) {
	var inputDetails []boltz.InputDetails
	var outputDetails []boltz.OutputDetails

	swapInputs, swapInputSum, err := scrooge.prepareSwaps(swaps)

	if err != nil {
		return nil, err
	}

	inputDetails = append(inputDetails, swapInputs...)

	if swapInputSum > 0 {
		address, err := scrooge.lnd.NewAddress()

		if err != nil {
			return nil, err
		}

		decodedAddress, err := btcutil.DecodeAddress(address, scrooge.chainParams)

		if err != nil {
			return nil, err
		}

		outputDetails = append(outputDetails, boltz.OutputDetails{
			Address: decodedAddress,
			Value:   swapInputSum,
		})
	}

	reverseSwapInputs, reverseSwapOutputs, err := scrooge.prepareReverseSwaps(reverseSwaps)

	if err != nil {
		return nil, err
	}

	inputDetails = append(inputDetails, reverseSwapInputs...)
	outputDetails = append(outputDetails, reverseSwapOutputs...)

	_, newTransactionSize, err := boltz.ConstructTransaction(inputDetails, outputDetails)

	if err != nil {
		return nil, err
	}

	feeSatPerVByteEstimation, err := scrooge.getFeeEstimation()

	if err != nil {
		return nil, err
	}

	transactionFee := feeSatPerVByteEstimation * newTransactionSize

	if unconfirmedTransaction != nil {
		transactionFee = utils.Max(
			transactionFee,
			int64(unconfirmedTransaction.Fee)+(int64(unconfirmedTransaction.Size)-newTransactionSize),
		)
	}


	feePerOutput := int64(math.Ceil(float64(transactionFee) / float64(len(outputDetails))))

	fmt.Println(transactionFee)
	fmt.Println(feePerOutput)
	fmt.Println(outputDetails[0].Value)

	for index := range outputDetails {
		outputDetails[index].Value -= feePerOutput
	}

	fmt.Println(outputDetails[0].Value)


	transaction, transactionSize, err := boltz.ConstructTransaction(inputDetails, outputDetails)

	if err != nil {
		return nil, err
	}

	err = scrooge.chainClient.BroadcastTransaction(transaction)

	if err != nil {
		return nil, err
	}

	if unconfirmedTransaction != nil {
		err = scrooge.database.RemoveUnconfirmedTransaction(unconfirmedTransaction.Id)
	}

	if err != nil {
		return nil, err
	}

	err = scrooge.database.CreateUnconfirmedTransaction(database.UnconfirmedTransaction{
		Id:   transaction.TxHash().String(),
		Size: int(transactionSize),
		Fee:  int(transactionFee),
	})

	for _, swap := range swaps {
		err = scrooge.database.SetSwapRefundTransactionId(&swap, transaction.TxHash().String())

		if err != nil {
			return nil, err
		}
	}

	for _, reverseSwap := range reverseSwaps {
		err = scrooge.database.SetReverseSwapClaimTransactionId(&reverseSwap, transaction.TxHash().String())

		if err != nil {
			return nil, err
		}
	}

	if err != nil {
		return nil, err
	}

	return transaction, nil
}
