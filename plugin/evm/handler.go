// (c) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package evm

import (
	"github.com/MetalBlockchain/metalgo/ids"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/shubhamdubey02/subnet-evm/core/txpool"
	"github.com/shubhamdubey02/subnet-evm/core/types"
	"github.com/shubhamdubey02/subnet-evm/plugin/evm/message"
)

// GossipHandler handles incoming gossip messages
type GossipHandler struct {
	vm     *VM
	txPool *txpool.TxPool
	stats  GossipStats
}

func NewGossipHandler(vm *VM, stats GossipStats) *GossipHandler {
	return &GossipHandler{
		vm:     vm,
		txPool: vm.txPool,
		stats:  stats,
	}
}

func (h *GossipHandler) HandleEthTxs(nodeID ids.NodeID, msg message.EthTxsGossip) error {
	log.Trace(
		"AppGossip called with EthTxsGossip",
		"peerID", nodeID,
		"size(txs)", len(msg.Txs),
	)

	if len(msg.Txs) == 0 {
		log.Trace(
			"AppGossip received empty EthTxsGossip Message",
			"peerID", nodeID,
		)
		return nil
	}

	// The maximum size of this encoded object is enforced by the codec.
	txs := make([]*types.Transaction, 0)
	if err := rlp.DecodeBytes(msg.Txs, &txs); err != nil {
		log.Trace(
			"AppGossip provided invalid txs",
			"peerID", nodeID,
			"err", err,
		)
		return nil
	}
	h.stats.IncEthTxsGossipReceived()
	wrapped := make([]*txpool.Transaction, len(txs))
	for i, tx := range txs {
		wrapped[i] = &txpool.Transaction{Tx: tx}
	}
	errs := h.txPool.Add(wrapped, false, false)
	for i, err := range errs {
		if err != nil {
			log.Trace(
				"AppGossip failed to add to mempool",
				"err", err,
				"tx", txs[i].Hash(),
			)
			if err == txpool.ErrAlreadyKnown {
				h.stats.IncEthTxsGossipReceivedKnown()
			} else {
				h.stats.IncEthTxsGossipReceivedError()
			}
			continue
		}
		h.stats.IncEthTxsGossipReceivedNew()
	}
	return nil
}
