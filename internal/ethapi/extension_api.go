// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package ethapi

import (
	"context"
	// "crypto/rand"
	// "encoding/hex"
	// "errors"
	// "fmt"
	"math/big"
	// "strings"
	// "time"

	// "github.com/davecgh/go-spew/spew"
	// "github.com/ethereum/go-ethereum/accounts"
	// "github.com/ethereum/go-ethereum/accounts/abi"
	// "github.com/ethereum/go-ethereum/accounts/keystore"
	// "github.com/ethereum/go-ethereum/accounts/scwallet"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"

	// "github.com/ethereum/go-ethereum/common/math"
	// "github.com/ethereum/go-ethereum/consensus/ethash"
	// "github.com/ethereum/go-ethereum/consensus/misc"
	// "github.com/ethereum/go-ethereum/core"
	// "github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	// "github.com/ethereum/go-ethereum/core/vm"
	// "github.com/ethereum/go-ethereum/crypto"
	// "github.com/ethereum/go-ethereum/eth/tracers/logger"
	// "github.com/ethereum/go-ethereum/log"
	// "github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/params"
	// "github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"
	// "github.com/tyler-smith/go-bip39"
	// "golang.org/x/crypto/sha3"
)

// PublicEthereumAPI provides an API to access Ethereum related information.
// It offers only methods that operate on public data that is freely available to anyone.
type ExtensionEthereumAPI struct {
	b          Backend
	ethereum   PublicEthereumAPI
	blockChain PublicBlockChainAPI
}

// NewExtensionEthereumAPI creates custom Ethereum protocol API.
func NewExtensionEthereumAPI(b Backend, ethereum PublicEthereumAPI, blockChain PublicBlockChainAPI) *ExtensionEthereumAPI {
	return &ExtensionEthereumAPI{b, ethereum, blockChain}
}

// HasCode returns boolean the code stored at the given address in the state for the given block number.
func (s *ExtensionEthereumAPI) HasCode(ctx context.Context, address common.Address, blockNrOrHash rpc.BlockNumberOrHash) (bool, error) {
	code, err := s.blockChain.GetCode(ctx, address, blockNrOrHash)
	if err != nil {
		return false, err
	}
	return code != nil, nil
}

// Given a transaction hash, returns its raw revert reason.
func (s *ExtensionEthereumAPI) GetTransactionError(ctx context.Context, hash common.Hash, overrides *StateOverride) (hexutil.Bytes, error) {
	// Try to return an already finalized transaction
	tx, _, blockNumber, _, err := s.b.GetTransaction(ctx, hash)
	if err != nil {
		return nil, err
	}
	if tx != nil {
		args := TransactionToTransactionArgs(tx, blockNumber, s.b.ChainConfig())
		result, err := DoCall(ctx, s.b, args, rpc.BlockNumberOrHashWithNumber(rpc.BlockNumber(blockNumber)), overrides, s.b.RPCEVMTimeout(), s.b.RPCGasCap())
		if err != nil {
			return nil, err
		}

		// If the result contains a revert reason, try to unpack and return it.
		if len(result.Revert()) > 0 {
			return result.Revert(), nil
		}
	}
	// Transaction unknown, return as such
	return nil, nil
}

// // GetTransactionReceipt returns the transaction receipt for the given transaction hash.
// func (s *TransactionAPI) GetTransactionReceiptWithError(ctx context.Context, hash common.Hash, overrides *StateOverride) (map[string]interface{}, error) {
// 	tx, blockHash, blockNumber, index, err := s.b.GetTransaction(ctx, hash)
// 	if err != nil {
// 		// When the transaction doesn't exist, the RPC method should return JSON null
// 		// as per specification.
// 		return nil, nil
// 	}
// 	receipts, err := s.b.GetReceipts(ctx, blockHash)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if len(receipts) <= int(index) {
// 		return nil, nil
// 	}
// 	receipt := receipts[index]

// 	// Derive the sender.
// 	bigblock := new(big.Int).SetUint64(blockNumber)
// 	signer := types.MakeSigner(s.b.ChainConfig(), bigblock)
// 	from, _ := types.Sender(signer, tx)

// 	fields := map[string]interface{}{
// 		"blockHash":         blockHash,
// 		"blockNumber":       hexutil.Uint64(blockNumber),
// 		"transactionHash":   hash,
// 		"transactionIndex":  hexutil.Uint64(index),
// 		"from":              from,
// 		"to":                tx.To(),
// 		"gasUsed":           hexutil.Uint64(receipt.GasUsed),
// 		"cumulativeGasUsed": hexutil.Uint64(receipt.CumulativeGasUsed),
// 		"contractAddress":   nil,
// 		"logs":              receipt.Logs,
// 		"logsBloom":         receipt.Bloom,
// 		"type":              hexutil.Uint(tx.Type()),
// 	}
// 	// Assign the effective gas price paid
// 	if !s.b.ChainConfig().IsLondon(bigblock) {
// 		fields["effectiveGasPrice"] = hexutil.Uint64(tx.GasPrice().Uint64())
// 	} else {
// 		header, err := s.b.HeaderByHash(ctx, blockHash)
// 		if err != nil {
// 			return nil, err
// 		}
// 		gasPrice := new(big.Int).Add(header.BaseFee, tx.EffectiveGasTipValue(header.BaseFee))
// 		fields["effectiveGasPrice"] = hexutil.Uint64(gasPrice.Uint64())
// 	}
// 	// Assign receipt status or post state.
// 	if len(receipt.PostState) > 0 {
// 		fields["root"] = hexutil.Bytes(receipt.PostState)
// 	} else {
// 		fields["status"] = hexutil.Uint(receipt.Status)
// 	}
// 	if receipt.Logs == nil {
// 		fields["logs"] = []*types.Log{}
// 	}
// 	// If the ContractAddress is 20 0x0 bytes, assume it is not a contract creation
// 	if receipt.ContractAddress != (common.Address{}) {
// 		fields["contractAddress"] = receipt.ContractAddress
// 	}
// 	return fields, nil
// }

// convert transaction to transaction args
func TransactionToTransactionArgs(tx *types.Transaction, blockNumber uint64, config *params.ChainConfig) TransactionArgs {
	signer := types.MakeSigner(config, new(big.Int).SetUint64(blockNumber))
	from, _ := types.Sender(signer, tx)
	gas := hexutil.Uint64(tx.Gas())
	nonce := hexutil.Uint64(tx.Nonce())
	data := hexutil.Bytes(tx.Data())
	result := TransactionArgs{
		From:    &from,
		To:      tx.To(),
		Gas:     &gas,
		Value:   (*hexutil.Big)(tx.Value()),
		Nonce:   &nonce,
		Data:    &data,
		ChainID: (*hexutil.Big)(tx.ChainId()),
	}

	switch tx.Type() {
	case types.LegacyTxType:
		result.GasPrice = (*hexutil.Big)(tx.GasPrice())
	case types.AccessListTxType:
		al := tx.AccessList()
		result.AccessList = &al
	case types.DynamicFeeTxType:
		result.MaxFeePerGas = (*hexutil.Big)(tx.GasFeeCap())
		result.MaxPriorityFeePerGas = (*hexutil.Big)(tx.GasTipCap())
	}
	return result
}
