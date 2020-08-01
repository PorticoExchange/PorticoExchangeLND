package scrooge

import (
	"database/sql"
	"errors"
	"github.com/BoltzExchange/boltz-lnd/boltz"
	"github.com/BoltzExchange/boltz-lnd/database"
	"github.com/BoltzExchange/boltz-lnd/logger"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"math"
	"strconv"
)

func (scrooge *Scrooge) FindLockupVout(addressToFind string, outputs []*wire.TxOut) (uint32, error) {
	for vout, output := range outputs {
		_, outputAddresses, _, err := txscript.ExtractPkScriptAddrs(output.PkScript, scrooge.chainParams)

		// Just ignore outputs we can't decode
		if err != nil {
			continue
		}

		for _, outputAddress := range outputAddresses {
			if outputAddress.EncodeAddress() == addressToFind {
				return uint32(vout), nil
			}
		}
	}

	return 0, errors.New("could not find lockup vout")
}

func (scrooge *Scrooge) getUnconfirmedSwaps() (transaction *database.UnconfirmedTransaction, swaps []database.Swap, reverseSwaps []database.ReverseSwap, err error) {
	unconfirmedTransaction, err := scrooge.database.QueryUnconfirmedTransaction()

	if err != nil {
		if err == sql.ErrNoRows {
			logger.Info("Didn't find transaction to batch")
			return nil, nil, nil, nil
		}

		return nil, nil, nil, err
	}

	isConfirmed, err := scrooge.chainClient.TransactionIsConfirmed(unconfirmedTransaction.Id)

	if err != nil {
		return nil, nil, nil, err
	}

	if isConfirmed {
		logger.Info("Transaction to batch (" + unconfirmedTransaction.Id + ") did confirm already")
		err = scrooge.database.RemoveUnconfirmedTransaction(unconfirmedTransaction.Id)

		return nil, nil, nil, err
	}

	swaps, err = scrooge.database.QuerySwapsByRefundTransaction(unconfirmedTransaction.Id)

	if err != nil {
		return nil, nil, nil, err
	}

	if len(swaps) > 0 {
		logger.Info("Found " + strconv.Itoa(len(swaps)) + " Swap refund transactions to batch")
	}

	reverseSwaps, err = scrooge.database.QueryReverseSwapsByClaimTransaction(unconfirmedTransaction.Id)

	if err != nil {
		return nil, nil, nil, err
	}

	if len(reverseSwaps) > 0 {
		logger.Info("Found " + strconv.Itoa(len(reverseSwaps)) + " Reverse Swap claim transactions to batch")
	}

	return unconfirmedTransaction, swaps, reverseSwaps, nil
}

func (scrooge *Scrooge) prepareSwaps(swaps []database.Swap) (inputDetails []boltz.InputDetails, inputSum int64, err error) {
	for _, swap := range swaps {
		lockupTransaction, err := scrooge.chainClient.GetRawTransaction(swap.LockupTransactionId)

		if err != nil {
			return nil, 0, err
		}

		lockupVout, err := scrooge.FindLockupVout(swap.Address, lockupTransaction.TxOut)

		if err != nil {
			return nil, 0, err
		}

		inputSum += lockupTransaction.TxOut[lockupVout].Value

		inputDetails = append(inputDetails, boltz.InputDetails{
			LockupTransaction:  lockupTransaction,
			Vout:               lockupVout,
			OutputType:         boltz.Compatibility,
			RedeemScript:       swap.RedeemScript,
			PrivateKey:         swap.PrivateKey,
			TimeoutBlockHeight: uint32(swap.TimoutBlockHeight),
		})
	}

	return inputDetails, inputSum, nil
}

func (scrooge *Scrooge) prepareReverseSwaps(reverseSwaps []database.ReverseSwap) (inputDetails []boltz.InputDetails, outputDetails []boltz.OutputDetails, err error) {
	for _, reverseSwap := range reverseSwaps {
		lockupTransaction, err := scrooge.chainClient.GetRawTransaction(reverseSwap.LockupTransactionId)

		if err != nil {
			return nil, nil, err
		}

		lockupAddress, err := boltz.WitnessScriptHashAddress(scrooge.chainParams, reverseSwap.RedeemScript)

		if err != nil {
			return nil, nil, err
		}

		lockupVout, err := scrooge.FindLockupVout(lockupAddress, lockupTransaction.TxOut)

		if err != nil {
			return nil, nil, err
		}

		inputDetails = append(inputDetails, boltz.InputDetails{
			LockupTransaction: lockupTransaction,
			Vout:              lockupVout,
			OutputType:        boltz.SegWit,
			RedeemScript:      reverseSwap.RedeemScript,
			PrivateKey:        reverseSwap.PrivateKey,
			Preimage:          reverseSwap.Preimage,
		})

		claimAddress, err := btcutil.DecodeAddress(reverseSwap.ClaimAddress, scrooge.chainParams)

		if err != nil {
			return nil, nil, err
		}

		outputDetails = append(outputDetails, boltz.OutputDetails{
			Address: claimAddress,
			Value:   lockupTransaction.TxOut[lockupVout].Value,
		})
	}

	return inputDetails, outputDetails, nil
}

func (scrooge *Scrooge) getFeeEstimation() (int64, error) {
	feeResponse, err := scrooge.lnd.EstimateFee(2)

	if err != nil {
		return 0, err
	}

	// Divide by 4 to get the fee per kilo vbyte and by 1000 to get the fee per vbyte
	return int64(math.Round(float64(feeResponse.SatPerKw) / 4000)), nil
}
