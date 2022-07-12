// Copyright 2014 The go-ethereum Authors
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

package core

import (
	// "errors"
	// "math"
	// "math/big"
	// "sort"
	// "sync"
	// "sync/atomic"
	// "time"

	// "github.com/ethereum/go-ethereum/common"
	// "github.com/ethereum/go-ethereum/common/prque"
	// "github.com/ethereum/go-ethereum/consensus/misc"
	// "github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	// "github.com/ethereum/go-ethereum/event"
	// "github.com/ethereum/go-ethereum/log"
	// "github.com/ethereum/go-ethereum/metrics"
	// "github.com/ethereum/go-ethereum/params"
)

// return pending txs sorted by price and nonce (may be included in the future)
func (pool *TxPool) Priced() types.Transactions {
	pending := pool.Pending(true)

	// Sort the transactions and cross check the nonce ordering
	txset := types.NewTransactionsByPriceAndNonce(pool.signer, pending, pool.priced.urgent.baseFee)

	txs := types.Transactions{}
	for tx := txset.Peek(); tx != nil; tx = txset.Peek() {
		txs = append(txs, tx)
		txset.Shift()
	}
	return txs
}
