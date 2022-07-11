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
/**
 * ext_hasCode(address, blockNumber/blockHash)
 * ext_getTransactionError(hash)
 * ext_pricedTransactions(limit = 300) -1: all, 0: [0~300]
 */
type ExtensionEthereumAPI struct {
	b               Backend
	ethereum        PublicEthereumAPI
	blockChain      PublicBlockChainAPI
	transactionPool PublicTransactionPoolAPI
	txpool          PublicTxPoolAPI
}

// NewExtensionEthereumAPI creates custom Ethereum protocol API.
func NewExtensionEthereumAPI(b Backend, ethereum PublicEthereumAPI, blockChain PublicBlockChainAPI, transactionPool PublicTransactionPoolAPI, txpool PublicTxPoolAPI) *ExtensionEthereumAPI {
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

// // PendingTransactions returns the transactions that are in the transaction pool
// // and have a from address that is one of the accounts this node manages.
// func (s *ExtensionEthereumAPI) SearchTransactions(address common.Address, from rpc.BlockNumberOrHash, to rpc.BlockNumberOrHash) ([]*RPCTransaction, error) {

// 	// pending, err := s.b.GetPoolTransactions()
// 	// if err != nil {
// 	// 	return nil, err
// 	// }
// 	// // accounts := make(map[common.Address]struct{})
// 	// // for _, wallet := range s.b.AccountManager().Wallets() {
// 	// // 	for _, account := range wallet.Accounts() {
// 	// // 		accounts[account.Address] = struct{}{}
// 	// // 	}
// 	// // }
// 	// curHeader := s.b.CurrentHeader()
// 	// transactions := make([]*RPCTransaction, 0, len(pending))
// 	// for _, tx := range pending {
// 	// 	from, _ := types.Sender(s.signer, tx)
// 	// 	if _, exists := accounts[from]; exists {
// 	// 		transactions = append(transactions, newRPCPendingTransaction(tx, curHeader, s.b.ChainConfig()))
// 	// 	}
// 	// }
// 	// return transactions, nil
// }

func (s *ExtensionEthereumAPI) PricedTransactions(limit int) []*RPCTransaction {
	curHeader := s.b.CurrentHeader()

	if limit == 0 {
		limit = 300
	}
	priced := s.b.TxPoolPriced()[:limit]
	transactions := make([]*RPCTransaction, 0, len(priced))
	for _, tx := range priced {
		transactions = append(transactions, newRPCPendingTransaction(tx, curHeader, s.b.ChainConfig()))
	}
	return transactions
}

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
