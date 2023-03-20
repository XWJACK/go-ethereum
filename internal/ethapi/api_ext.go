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

	"fmt"
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
	"github.com/ethereum/go-ethereum/common/math"

	// "github.com/ethereum/go-ethereum/common/math"
	// "github.com/ethereum/go-ethereum/consensus/ethash"
	// "github.com/ethereum/go-ethereum/consensus/misc"

	// "github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"

	// "github.com/ethereum/go-ethereum/crypto"
	// "github.com/ethereum/go-ethereum/eth/tracers/logger"

	// "github.com/ethereum/go-ethereum/p2p"

	// "github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"
	// "github.com/tyler-smith/go-bip39"
	// "golang.org/x/crypto/sha3"
)

// PublicEthereumAPI provides an API to access Ethereum related information.
// It offers only methods that operate on public data that is freely available to anyone.
/**
 * ext_hasCode(address, blockNumber/blockHash)
 * ext_getTransactionError(hash)
 * ext_pricedTransactions(limit = 300) -1: all, 0: [0~300]
 */
type ExtensionEthereumAPI struct {
	b               Backend
	ethereum        EthereumAPI
	blockChain      BlockChainAPI
	transactionPool TransactionAPI
	txpool          TxPoolAPI
}

// NewExtensionEthereumAPI creates custom Ethereum protocol API.
func NewExtensionEthereumAPI(b Backend, ethereum EthereumAPI, blockChain BlockChainAPI, transactionPool TransactionAPI, txpool TxPoolAPI) *ExtensionEthereumAPI {
	return &ExtensionEthereumAPI{b, ethereum, blockChain, transactionPool, txpool}
}

// HasCode returns boolean the code stored at the given address in the state for the given block number.
func (s *ExtensionEthereumAPI) HasCode(ctx context.Context, address common.Address, blockNrOrHash rpc.BlockNumberOrHash) (bool, error) {
	code, err := s.blockChain.GetCode(ctx, address, blockNrOrHash)
	if err != nil {
		return false, err
	}
	return code != nil, nil
}

// func (s *ExtensionEthereumAPI) PricedTransactions(limit int) []*RPCTransaction {
// 	curHeader := s.b.CurrentHeader()

// 	priced := s.b.TxPoolPriced()
// 	if limit == 0 {
// 		limit = 300
// 	}
// 	if len(priced) < limit {
// 		limit = len(priced)
// 	}
// 	priced = priced[:limit]
// 	transactions := make([]*RPCTransaction, 0, len(priced))
// 	for _, tx := range priced {
// 		transactions = append(transactions, newRPCPendingTransaction(tx, curHeader, s.b.ChainConfig()))
// 	}
// 	return transactions
// }

// GetBlockReceipts retrieves the receipts of a single block.
func (api *ExtensionEthereumAPI) GetBlockReceipts(ctx context.Context, blockNrOrHash rpc.BlockNumberOrHash) (map[string]interface{}, error) {
	var block *types.Block
	var hash common.Hash
	if h, ok := blockNrOrHash.Hash(); ok {
		hash = h
	} else {
		b, err := api.b.BlockByNumberOrHash(ctx, blockNrOrHash)
		if err != nil {
			return nil, err
		}
		hash = b.Hash()
		block = b
	}
	receipts, err := api.b.GetReceipts(ctx, hash)
	if err != nil {
		return nil, err
	}
	// Derive the sender.
	signer := types.MakeSigner(api.b.ChainConfig(), block.Number())

	txs := block.Transactions()

	if txs.Len() != receipts.Len() {
		return nil, fmt.Errorf("txs not equal receipts")
	}

	result := map[string]interface{}{
		"number":           (*hexutil.Big)(block.Number()),
		"hash":             block.Hash(),
		"parentHash":       block.ParentHash(),
		"nonce":            block.Nonce(),
		"mixHash":          block.MixDigest(),
		"sha3Uncles":       block.UncleHash(),
		"logsBloom":        block.Bloom(),
		"stateRoot":        block.Root(),
		"miner":            block.Coinbase(),
		"extraData":        hexutil.Bytes(block.Extra()),
		"size":             hexutil.Uint64(block.Size()),
		"gasLimit":         hexutil.Uint64(block.GasLimit()),
		"gasUsed":          hexutil.Uint64(block.GasUsed()),
		"timestamp":        hexutil.Uint64(block.Time()),
		"transactionsRoot": block.TxHash(),
		"receiptsRoot":     block.ReceiptHash(),
		"baseFeePerGas":    (*hexutil.Big)(block.BaseFee()),
	}

	transactions := make([]map[string]interface{}, receipts.Len())
	for index, receipt := range receipts {
		tx := txs[index]

		from, _ := types.Sender(signer, tx)

		logs := make([]map[string]interface{}, len(receipt.Logs))
		for i, log := range receipt.Logs {
			logs[i] = map[string]interface{}{
				"address":  log.Address,
				"topics":   log.Topics,
				"data":     hexutil.Bytes(log.Data),
				"logIndex": log.Index,
				"removed":  log.Removed,
			}
		}

		transactions[index] = map[string]interface{}{
			"hash":              tx.Hash(),
			"from":              from,
			"gas":               hexutil.Uint64(tx.Gas()),
			"gasPrice":          (*hexutil.Big)(tx.GasPrice()),
			"input":             hexutil.Bytes(tx.Data()),
			"nonce":             hexutil.Uint64(tx.Nonce()),
			"to":                tx.To(),
			"value":             (*hexutil.Big)(tx.Value()),
			"contractAddress":   receipt.ContractAddress,
			"transactionIndex":  hexutil.Uint64(index),
			"gasUsed":           hexutil.Uint64(receipt.GasUsed),
			"cumulativeGasUsed": hexutil.Uint64(receipt.CumulativeGasUsed),
			"logs":              logs,
			"type":              hexutil.Uint(tx.Type()),
			// "logsBloom":         receipt.Bloom,
		}

		switch tx.Type() {
		case types.LegacyTxType:
			// if a legacy transaction has an EIP-155 chain id, include it explicitly
			if id := tx.ChainId(); id.Sign() != 0 {
				transactions[index]["chainId"] = (*hexutil.Big)(id)
			}
		case types.AccessListTxType:
			al := tx.AccessList()
			transactions[index]["accessList"] = &al
			transactions[index]["chainId"] = (*hexutil.Big)(tx.ChainId())
		case types.DynamicFeeTxType:
			al := tx.AccessList()
			transactions[index]["accessList"] = &al
			transactions[index]["chainId"] = (*hexutil.Big)(tx.ChainId())

			transactions[index]["maxFeePerGas"] = (*hexutil.Big)(tx.GasFeeCap())
			transactions[index]["maxPriorityFeePerGas"] = (*hexutil.Big)(tx.GasTipCap())

			// if the transaction has been mined, compute the effective gas price
			if block.BaseFee() != nil && block.Hash() != (common.Hash{}) {
				// price = min(tip, gasFeeCap - baseFee) + baseFee
				price := math.BigMin(new(big.Int).Add(tx.GasTipCap(), block.BaseFee()), tx.GasFeeCap())
				transactions[index]["gasPrice"] = (*hexutil.Big)(price)
			} else {
				transactions[index]["gasPrice"] = (*hexutil.Big)(tx.GasFeeCap())
			}
		}
	}

	result["transactions"] = transactions
	return result, nil
}
