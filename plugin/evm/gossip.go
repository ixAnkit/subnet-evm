// Copyright (C) 2019-2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package evm

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/MetalBlockchain/metalgo/ids"
	"github.com/MetalBlockchain/metalgo/utils/logging"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/MetalBlockchain/metalgo/network/p2p"
	"github.com/MetalBlockchain/metalgo/network/p2p/gossip"

	"github.com/shubhamdubey02/subnet-evm/core"
	"github.com/shubhamdubey02/subnet-evm/core/txpool"
	"github.com/shubhamdubey02/subnet-evm/core/types"
	"github.com/shubhamdubey02/subnet-evm/eth"
)

const pendingTxsBuffer = 10

var (
	_ p2p.Handler = (*txGossipHandler)(nil)

	_ gossip.Gossipable               = (*GossipEthTx)(nil)
	_ gossip.Marshaller[*GossipEthTx] = (*GossipEthTxMarshaller)(nil)
	_ gossip.Set[*GossipEthTx]        = (*GossipEthTxPool)(nil)

	_ eth.PushGossiper = (*EthPushGossiper)(nil)
)

func newTxGossipHandler[T gossip.Gossipable](
	log logging.Logger,
	marshaller gossip.Marshaller[T],
	mempool gossip.Set[T],
	metrics gossip.Metrics,
	maxMessageSize int,
	throttlingPeriod time.Duration,
	throttlingLimit int,
	validators *p2p.Validators,
) txGossipHandler {
	// push gossip messages can be handled from any peer
	handler := gossip.NewHandler(
		log,
		marshaller,
		mempool,
		metrics,
		maxMessageSize,
	)

	// pull gossip requests are filtered by validators and are throttled
	// to prevent spamming
	validatorHandler := p2p.NewValidatorHandler(
		p2p.NewThrottlerHandler(
			handler,
			p2p.NewSlidingWindowThrottler(throttlingPeriod, throttlingLimit),
			log,
		),
		validators,
		log,
	)

	return txGossipHandler{
		appGossipHandler:  handler,
		appRequestHandler: validatorHandler,
	}
}

type txGossipHandler struct {
	appGossipHandler  p2p.Handler
	appRequestHandler p2p.Handler
}

func (t txGossipHandler) AppGossip(ctx context.Context, nodeID ids.NodeID, gossipBytes []byte) {
	t.appGossipHandler.AppGossip(ctx, nodeID, gossipBytes)
}

func (t txGossipHandler) AppRequest(ctx context.Context, nodeID ids.NodeID, deadline time.Time, requestBytes []byte) ([]byte, error) {
	return t.appRequestHandler.AppRequest(ctx, nodeID, deadline, requestBytes)
}
func (t txGossipHandler) CrossChainAppRequest(context.Context, ids.ID, time.Time, []byte) ([]byte, error) {
	return nil, nil
}

func NewGossipEthTxPool(mempool *txpool.TxPool, registerer prometheus.Registerer) (*GossipEthTxPool, error) {
	bloom, err := gossip.NewBloomFilter(registerer, "eth_tx_bloom_filter", txGossipBloomMinTargetElements, txGossipBloomTargetFalsePositiveRate, txGossipBloomResetFalsePositiveRate)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize bloom filter: %w", err)
	}

	return &GossipEthTxPool{
		mempool:    mempool,
		pendingTxs: make(chan core.NewTxsEvent, pendingTxsBuffer),
		bloom:      bloom,
	}, nil
}

type GossipEthTxPool struct {
	mempool    *txpool.TxPool
	pendingTxs chan core.NewTxsEvent

	bloom *gossip.BloomFilter
	lock  sync.RWMutex
}

func (g *GossipEthTxPool) Subscribe(ctx context.Context) {
	g.mempool.SubscribeNewTxsEvent(g.pendingTxs)

	for {
		select {
		case <-ctx.Done():
			log.Debug("shutting down subscription")
			return
		case pendingTxs := <-g.pendingTxs:
			g.lock.Lock()
			optimalElements := (g.mempool.PendingSize(false) + len(pendingTxs.Txs)) * txGossipBloomChurnMultiplier
			for _, pendingTx := range pendingTxs.Txs {
				tx := &GossipEthTx{Tx: pendingTx}
				g.bloom.Add(tx)
				reset, err := gossip.ResetBloomFilterIfNeeded(g.bloom, optimalElements)
				if err != nil {
					log.Error("failed to reset bloom filter", "err", err)
					continue
				}

				if reset {
					log.Debug("resetting bloom filter", "reason", "reached max filled ratio")

					g.mempool.IteratePending(func(tx *txpool.Transaction) bool {
						g.bloom.Add(&GossipEthTx{Tx: tx.Tx})
						return true
					})
				}
			}
			g.lock.Unlock()
		}
	}
}

// Add enqueues the transaction to the mempool. Subscribe should be called
// to receive an event if tx is actually added to the mempool or not.
func (g *GossipEthTxPool) Add(tx *GossipEthTx) error {
	return g.mempool.Add([]*txpool.Transaction{{Tx: tx.Tx}}, false, false)[0]
}

// Has should just return whether or not the [txID] is still in the mempool,
// not whether it is in the mempool AND pending.
func (g *GossipEthTxPool) Has(txID ids.ID) bool {
	return g.mempool.Has(common.Hash(txID))
}

func (g *GossipEthTxPool) Iterate(f func(tx *GossipEthTx) bool) {
	g.mempool.IteratePending(func(tx *txpool.Transaction) bool {
		return f(&GossipEthTx{Tx: tx.Tx})
	})
}

func (g *GossipEthTxPool) GetFilter() ([]byte, []byte) {
	g.lock.RLock()
	defer g.lock.RUnlock()

	return g.bloom.Marshal()
}

type GossipEthTxMarshaller struct{}

func (g GossipEthTxMarshaller) MarshalGossip(tx *GossipEthTx) ([]byte, error) {
	return tx.Tx.MarshalBinary()
}

func (g GossipEthTxMarshaller) UnmarshalGossip(bytes []byte) (*GossipEthTx, error) {
	tx := &GossipEthTx{
		Tx: &types.Transaction{},
	}

	return tx, tx.Tx.UnmarshalBinary(bytes)
}

type GossipEthTx struct {
	Tx *types.Transaction
}

func (tx *GossipEthTx) GossipID() ids.ID {
	return ids.ID(tx.Tx.Hash())
}

// EthPushGossiper is used by the ETH backend to push transactions issued over
// the RPC and added to the mempool to peers.
type EthPushGossiper struct {
	vm *VM
}

func (e *EthPushGossiper) Add(tx *types.Transaction) {
	// eth.Backend is initialized before the [ethTxPushGossiper] is created, so
	// we just ignore any gossip requests until it is set.
	ethTxPushGossiper := e.vm.ethTxPushGossiper.Get()
	if ethTxPushGossiper == nil {
		return
	}
	ethTxPushGossiper.Add(&GossipEthTx{tx})
}
