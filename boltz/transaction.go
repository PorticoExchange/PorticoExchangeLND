package boltz

import (
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"math"
)

type OutputType int

const (
	SegWit OutputType = iota
	Compatibility
	Legacy
)

type InputDetails struct {
	LockupTransaction *wire.MsgTx
	Vout              uint32
	OutputType        OutputType

	RedeemScript []byte
	PrivateKey   *btcec.PrivateKey

	// Should be set to an empty array in case of a refund
	Preimage []byte

	// Can be zero in case of a claim transaction
	TimeoutBlockHeight uint32
}

type OutputDetails struct {
	Address btcutil.Address
	Value   int64
}

func ConstructTransaction(inputs []InputDetails, outputs []OutputDetails) (*wire.MsgTx, int64, error) {
	transaction, err := constructTransaction(inputs, outputs)

	if err != nil {
		return nil, 0, err
	}

	return transaction, calculateVByteSize(transaction), nil
}

func constructTransaction(inputs []InputDetails, outputs []OutputDetails) (*wire.MsgTx, error) {
	transaction := wire.NewMsgTx(wire.TxVersion)

	var inputSum int64

	for _, input := range inputs {
		// Set the highest timeout block height as locktime
		if input.TimeoutBlockHeight > transaction.LockTime {
			transaction.LockTime = input.TimeoutBlockHeight
		}

		// Calculate the sum of all inputs
		inputSum += input.LockupTransaction.TxOut[input.Vout].Value

		// Add the input to the transaction
		inputHash := input.LockupTransaction.TxHash()
		input := wire.NewTxIn(wire.NewOutPoint(&inputHash, input.Vout), nil, nil)

		// Enable RBF: https://github.com/bitcoin/bips/blob/master/bip-0125.mediawiki#summary
		input.Sequence = 0

		transaction.AddTxIn(input)
	}

	// Add the outputs
	for _, output := range outputs {
		outputScript, err := txscript.PayToAddrScript(output.Address)

		if err != nil {
			return nil, err
		}

		transaction.AddTxOut(&wire.TxOut{
			PkScript: outputScript,
			Value:    output.Value,
		})
	}

	// Construct the signature script and witnesses and sign the inputs
	for i, input := range inputs {
		if input.Preimage == nil {
			input.Preimage = []byte{}
		}

		switch input.OutputType {
		case Legacy:
			// Set the signed signature script for legacy output
			signature, err := txscript.RawTxInSignature(
				transaction,
				i,
				input.RedeemScript,
				txscript.SigHashAll,
				input.PrivateKey,
			)

			if err != nil {
				return nil, err
			}

			signatureScriptBuilder := txscript.NewScriptBuilder()
			signatureScriptBuilder.AddData(signature)
			signatureScriptBuilder.AddData(input.Preimage)
			signatureScriptBuilder.AddData(input.RedeemScript)

			signatureScript, err := signatureScriptBuilder.Script()

			if err != nil {
				return nil, err
			}

			transaction.TxIn[i].SignatureScript = signatureScript

		case Compatibility:
			// Set the signature script for compatibility outputs
			signatureScriptBuilder := txscript.NewScriptBuilder()
			signatureScriptBuilder.AddData(createNestedP2shScript(input.RedeemScript))

			signatureScript, err := signatureScriptBuilder.Script()

			if err != nil {
				return nil, err
			}

			transaction.TxIn[i].SignatureScript = signatureScript
		}

		// Add the signed witness in case the output is not a legacy one
		if input.OutputType != Legacy {
			signatureHash := txscript.NewTxSigHashes(transaction)
			signature, err := txscript.RawTxInWitnessSignature(
				transaction,
				signatureHash,
				i,
				input.LockupTransaction.TxOut[input.Vout].Value,
				input.RedeemScript,
				txscript.SigHashAll,
				input.PrivateKey,
			)

			if err != nil {
				return nil, err
			}

			transaction.TxIn[i].Witness = wire.TxWitness{signature, input.Preimage, input.RedeemScript}
		}
	}

	return transaction, nil
}

func calculateVByteSize(transaction *wire.MsgTx) int64 {
	witnessSize := transaction.SerializeSize() - transaction.SerializeSizeStripped()
	vByte := int64(transaction.SerializeSizeStripped()) + int64(math.Ceil(float64(witnessSize)/4))

	return vByte
}
