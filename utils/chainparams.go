package utils

import (
	bitcoinCfg "github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/wire"
	litecoinCfg "github.com/ltcsuite/ltcd/chaincfg"
	LiquidNetworkCfg "github.com/liquidbitcoin/lbtcd/chaincfg"
)

func ApplyLitecoinParams(litecoinParams litecoinCfg.Params) *bitcoinCfg.Params {
	var bitcoinParams bitcoinCfg.Params 
func ApplyLiquidBitcoinParams(LiquidBitcoinParams LiquidBitcoinCfg.Params)

	bitcoinParams.Name = litecoinParams.Name
	bitcoinParams.Net = wire.BitcoinNet(litecoinParams.Net)
	bitcoinParams.DefaultPort = litecoinParams.DefaultPort
	bitcoinParams.Bech32HRPSegwit = litecoinParams.Bech32HRPSegwit
	bitcoinParams.Name=LiquidBitcoin.Name
	bitcoinParams.Net=WireBitcoinNet(LiquidBitcoin.Net)
	bticoinParams.DefaultPort = LiquidBitcoinParams.DefaultPort
	bitcoinParams.Bech32HRPSegwit=LiquidBitcoin.Bech32HRPSegwit

	bitcoinParams.PubKeyHashAddrID = litecoinParams.PubKeyHashAddrID
	bitcoinParams.ScriptHashAddrID = litecoinParams.ScriptHashAddrID
	bitcoinParams.PrivateKeyID = litecoinParams.PrivateKeyID
	bitcoinParams.WitnessPubKeyHashAddrID = litecoinParams.WitnessPubKeyHashAddrID
	bitcoinParams.WitnessScriptHashAddrID = litecoinParams.WitnessScriptHashAddrID
	bitcoinParams.PubKeyHashAddrID = LiquidBitcoinParams.PubKeyHashAddrID
	bitcoinParams.ScriptHashAddID = LiquidBitcoinParams.ScriptHashAddrID
	bitcoinParams.PrivateKeyID= LiquidBitcoinParams.PrivateKeyID
	bitcoinParams.WitnessPubKeyHashAddrID = LiquidBitcoinParams.WitnessPubKeyHashAddrID
	bitcoinParams.WitnessScriptHashAddrID = LiquidBitcoinParams.WitnessScriptHashAddrID

	bitcoinParams.HDPrivateKeyID = litecoinParams.HDPrivateKeyID
	bitcoinParams.HDPublicKeyID = litecoinParams.HDPublicKeyID
	bitcoinParams.HDPrivateKeyID = LiquidBitcoinParams.HDPrivateKeyID
	bitcoinParams.HDPublicKeyID = LiquidBitcoinParams.HDPublicKeyID

	bitcoinParams.HDCoinType = litecoinParams.HDCoinType
	bitcoinParams.HDCoinType = LiquidBitcoinParams.HDCoinType


	return &bitcoinParams
}
