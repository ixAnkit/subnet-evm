// (c) 2019-2020, Ava Labs, Inc.
//
// This file is a derived work, based on the go-ethereum library whose original
// notices appear below.
//
// It is distributed under a license compatible with the licensing terms of the
// original code from which it is derived.
//
// Much love to the original authors for their work.
// **********
// Copyright 2016 The go-ethereum Authors
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

package params

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/MetalBlockchain/metalgo/utils/constants"
	"github.com/ethereum/go-ethereum/common"
	"github.com/shubhamdubey02/subnet-evm/commontype"
	"github.com/shubhamdubey02/subnet-evm/precompile/modules"
	"github.com/shubhamdubey02/subnet-evm/precompile/precompileconfig"
	"github.com/shubhamdubey02/subnet-evm/utils"
)

const maxJSONLen = 64 * 1024 * 1024 // 64MB

var (
	errNonGenesisForkByHeight = errors.New("subnet-evm only supports forking by height at the genesis block")

	SubnetEVMChainID = big.NewInt(43214)

	// For legacy tests
	MinGasPrice        int64 = 225_000_000_000
	TestInitialBaseFee int64 = 225_000_000_000
	TestMaxBaseFee     int64 = 225_000_000_000

	DynamicFeeExtraDataSize        = 80
	RollupWindow            uint64 = 10

	DefaultFeeConfig = commontype.FeeConfig{
		GasLimit:        big.NewInt(8_000_000),
		TargetBlockRate: 2, // in seconds

		MinBaseFee:               big.NewInt(25_000_000_000),
		TargetGas:                big.NewInt(15_000_000),
		BaseFeeChangeDenominator: big.NewInt(36),

		MinBlockGasCost:  big.NewInt(0),
		MaxBlockGasCost:  big.NewInt(1_000_000),
		BlockGasCostStep: big.NewInt(200_000),
	}
)

var (
	// SubnetEVMDefaultConfig is the default configuration
	// without any network upgrades.
	SubnetEVMDefaultChainConfig = &ChainConfig{
		ChainID:            SubnetEVMChainID,
		FeeConfig:          DefaultFeeConfig,
		AllowFeeRecipients: false,

		HomesteadBlock:      big.NewInt(0),
		EIP150Block:         big.NewInt(0),
		EIP155Block:         big.NewInt(0),
		EIP158Block:         big.NewInt(0),
		ByzantiumBlock:      big.NewInt(0),
		ConstantinopleBlock: big.NewInt(0),
		PetersburgBlock:     big.NewInt(0),
		IstanbulBlock:       big.NewInt(0),
		MuirGlacierBlock:    big.NewInt(0),
		NetworkUpgrades:     getDefaultNetworkUpgrades(constants.MainnetID), // This can be changed to correct network (local, test) via VM.
		GenesisPrecompiles:  Precompiles{},
	}

	TestChainConfig = &ChainConfig{
		AvalancheContext:    AvalancheContext{utils.TestSnowContext()},
		ChainID:             big.NewInt(1),
		FeeConfig:           DefaultFeeConfig,
		AllowFeeRecipients:  false,
		HomesteadBlock:      big.NewInt(0),
		EIP150Block:         big.NewInt(0),
		EIP155Block:         big.NewInt(0),
		EIP158Block:         big.NewInt(0),
		ByzantiumBlock:      big.NewInt(0),
		ConstantinopleBlock: big.NewInt(0),
		PetersburgBlock:     big.NewInt(0),
		IstanbulBlock:       big.NewInt(0),
		MuirGlacierBlock:    big.NewInt(0),
		NetworkUpgrades: NetworkUpgrades{
			SubnetEVMTimestamp: utils.NewUint64(0),
			DurangoTimestamp:   utils.NewUint64(0),
		},
		GenesisPrecompiles: Precompiles{},
		UpgradeConfig:      UpgradeConfig{},
	}

	TestSubnetEVMConfig = &ChainConfig{
		AvalancheContext:    AvalancheContext{utils.TestSnowContext()},
		ChainID:             big.NewInt(1),
		FeeConfig:           DefaultFeeConfig,
		AllowFeeRecipients:  false,
		HomesteadBlock:      big.NewInt(0),
		EIP150Block:         big.NewInt(0),
		EIP155Block:         big.NewInt(0),
		EIP158Block:         big.NewInt(0),
		ByzantiumBlock:      big.NewInt(0),
		ConstantinopleBlock: big.NewInt(0),
		PetersburgBlock:     big.NewInt(0),
		IstanbulBlock:       big.NewInt(0),
		MuirGlacierBlock:    big.NewInt(0),
		NetworkUpgrades: NetworkUpgrades{
			SubnetEVMTimestamp: utils.NewUint64(0),
		},
		GenesisPrecompiles: Precompiles{},
		UpgradeConfig:      UpgradeConfig{},
	}

	TestPreSubnetEVMConfig = &ChainConfig{
		AvalancheContext:    AvalancheContext{utils.TestSnowContext()},
		ChainID:             big.NewInt(1),
		FeeConfig:           DefaultFeeConfig,
		AllowFeeRecipients:  false,
		HomesteadBlock:      big.NewInt(0),
		EIP150Block:         big.NewInt(0),
		EIP155Block:         big.NewInt(0),
		EIP158Block:         big.NewInt(0),
		ByzantiumBlock:      big.NewInt(0),
		ConstantinopleBlock: big.NewInt(0),
		PetersburgBlock:     big.NewInt(0),
		IstanbulBlock:       big.NewInt(0),
		MuirGlacierBlock:    big.NewInt(0),
		NetworkUpgrades:     NetworkUpgrades{},
		GenesisPrecompiles:  Precompiles{},
		UpgradeConfig:       UpgradeConfig{},
	}

	TestRules = TestChainConfig.Rules(new(big.Int), 0)
)

func getUpgradeTime(networkID uint32, upgradeTimes map[uint32]time.Time) *uint64 {
	if upgradeTime, ok := upgradeTimes[networkID]; ok {
		return utils.TimeToNewUint64(upgradeTime)
	}
	// If the upgrade time isn't specified, default being enabled in the
	// genesis.
	return utils.NewUint64(0)
}

// ChainConfig is the core config which determines the blockchain settings.
//
// ChainConfig is stored in the database on a per block basis. This means
// that any network, identified by its genesis block, can have its own
// set of configuration options.
type ChainConfig struct {
	ChainID *big.Int `json:"chainId"` // chainId identifies the current chain and is used for replay protection

	HomesteadBlock *big.Int `json:"homesteadBlock,omitempty"` // Homestead switch block (nil = no fork, 0 = already homestead)

	// EIP150 implements the Gas price changes (https://github.com/ethereum/EIPs/issues/150)
	EIP150Block *big.Int `json:"eip150Block,omitempty"` // EIP150 HF block (nil = no fork)
	EIP155Block *big.Int `json:"eip155Block,omitempty"` // EIP155 HF block
	EIP158Block *big.Int `json:"eip158Block,omitempty"` // EIP158 HF block

	ByzantiumBlock      *big.Int `json:"byzantiumBlock,omitempty"`      // Byzantium switch block (nil = no fork, 0 = already on byzantium)
	ConstantinopleBlock *big.Int `json:"constantinopleBlock,omitempty"` // Constantinople switch block (nil = no fork, 0 = already activated)
	PetersburgBlock     *big.Int `json:"petersburgBlock,omitempty"`     // Petersburg switch block (nil = same as Constantinople)
	IstanbulBlock       *big.Int `json:"istanbulBlock,omitempty"`       // Istanbul switch block (nil = no fork, 0 = already on istanbul)
	MuirGlacierBlock    *big.Int `json:"muirGlacierBlock,omitempty"`    // Eip-2384 (bomb delay) switch block (nil = no fork, 0 = already activated)

	// Cancun activates the Cancun upgrade from Ethereum. (nil = no fork, 0 = already activated)
	CancunTime *uint64 `json:"cancunTime,omitempty"`

	NetworkUpgrades // Config for timestamps that enable network upgrades. Skip encoding/decoding directly into ChainConfig.

	AvalancheContext `json:"-"` // Avalanche specific context set during VM initialization. Not serialized.

	FeeConfig          commontype.FeeConfig `json:"feeConfig"`                    // Set the configuration for the dynamic fee algorithm
	AllowFeeRecipients bool                 `json:"allowFeeRecipients,omitempty"` // Allows fees to be collected by block builders.

	GenesisPrecompiles Precompiles `json:"-"` // Config for enabling precompiles from genesis. JSON encode/decode will be handled by the custom marshaler/unmarshaler.
	UpgradeConfig      `json:"-"`  // Config specified in upgradeBytes (avalanche network upgrades or enable/disabling precompiles). Skip encoding/decoding directly into ChainConfig.
}

// Description returns a human-readable description of ChainConfig.
func (c *ChainConfig) Description() string {
	var banner string

	banner += fmt.Sprintf("Chain ID:  %v\n", c.ChainID)
	banner += "Consensus: Dummy Consensus Engine\n\n"

	// Create a list of forks with a short description of them. Forks that only
	// makes sense for mainnet should be optional at printing to avoid bloating
	// the output for testnets and private networks.
	banner += "Hard Forks (block based):\n"
	banner += fmt.Sprintf(" - Homestead:                   #%-8v (https://github.com/ethereum/execution-specs/blob/master/network-upgrades/mainnet-upgrades/homestead.md)\n", c.HomesteadBlock)
	banner += fmt.Sprintf(" - Tangerine Whistle (EIP 150): #%-8v (https://github.com/ethereum/execution-specs/blob/master/network-upgrades/mainnet-upgrades/tangerine-whistle.md)\n", c.EIP150Block)
	banner += fmt.Sprintf(" - Spurious Dragon/1 (EIP 155): #%-8v (https://github.com/ethereum/execution-specs/blob/master/network-upgrades/mainnet-upgrades/spurious-dragon.md)\n", c.EIP155Block)
	banner += fmt.Sprintf(" - Spurious Dragon/2 (EIP 158): #%-8v (https://github.com/ethereum/execution-specs/blob/master/network-upgrades/mainnet-upgrades/spurious-dragon.md)\n", c.EIP155Block)
	banner += fmt.Sprintf(" - Byzantium:                   #%-8v (https://github.com/ethereum/execution-specs/blob/master/network-upgrades/mainnet-upgrades/byzantium.md)\n", c.ByzantiumBlock)
	banner += fmt.Sprintf(" - Constantinople:              #%-8v (https://github.com/ethereum/execution-specs/blob/master/network-upgrades/mainnet-upgrades/constantinople.md)\n", c.ConstantinopleBlock)
	banner += fmt.Sprintf(" - Petersburg:                  #%-8v (https://github.com/ethereum/execution-specs/blob/master/network-upgrades/mainnet-upgrades/petersburg.md)\n", c.PetersburgBlock)
	banner += fmt.Sprintf(" - Istanbul:                    #%-8v (https://github.com/ethereum/execution-specs/blob/master/network-upgrades/mainnet-upgrades/istanbul.md)\n", c.IstanbulBlock)
	if c.MuirGlacierBlock != nil {
		banner += fmt.Sprintf(" - Muir Glacier:                #%-8v (https://github.com/ethereum/execution-specs/blob/master/network-upgrades/mainnet-upgrades/muir-glacier.md)\n", c.MuirGlacierBlock)
	}

	banner += "Hard forks (timestamp based):\n"
	banner += fmt.Sprintf(" - Cancun Timestamp:              @%-10v (https://github.com/MetalBlockchain/metalgo/releases/tag/v1.12.0)\n", ptrToString(c.CancunTime))

	banner += "Avalanche Upgrades (timestamp based):\n"
	banner += fmt.Sprintf(" - SubnetEVM Timestamp:           @%-10v (https://github.com/MetalBlockchain/metalgo/releases/tag/v1.10.0)\n", ptrToString(c.SubnetEVMTimestamp))
	banner += fmt.Sprintf(" - Durango Timestamp:            @%-10v (https://github.com/MetalBlockchain/metalgo/releases/tag/v1.11.0)\n", ptrToString(c.DurangoTimestamp))
	banner += "\n"

	precompileUpgradeBytes, err := json.Marshal(c.GenesisPrecompiles)
	if err != nil {
		precompileUpgradeBytes = []byte("cannot marshal PrecompileUpgrade")
	}
	banner += fmt.Sprintf("Precompile Upgrades: %s", string(precompileUpgradeBytes))
	banner += "\n"

	upgradeConfigBytes, err := json.Marshal(c.UpgradeConfig)
	if err != nil {
		upgradeConfigBytes = []byte("cannot marshal UpgradeConfig")
	}
	banner += fmt.Sprintf("Upgrade Config: %s", string(upgradeConfigBytes))
	banner += "\n"

	feeBytes, err := json.Marshal(c.FeeConfig)
	if err != nil {
		feeBytes = []byte("cannot marshal FeeConfig")
	}
	banner += fmt.Sprintf("Fee Config: %s", string(feeBytes))
	banner += "\n"

	banner += fmt.Sprintf("Allow Fee Recipients: %v", c.AllowFeeRecipients)
	banner += "\n"
	return banner
}

func (c *ChainConfig) SetNetworkUpgradeDefaults() {
	if c.HomesteadBlock == nil {
		c.HomesteadBlock = big.NewInt(0)
	}
	if c.EIP150Block == nil {
		c.EIP150Block = big.NewInt(0)
	}
	if c.EIP155Block == nil {
		c.EIP155Block = big.NewInt(0)
	}
	if c.EIP158Block == nil {
		c.EIP158Block = big.NewInt(0)
	}
	if c.ByzantiumBlock == nil {
		c.ByzantiumBlock = big.NewInt(0)
	}
	if c.ConstantinopleBlock == nil {
		c.ConstantinopleBlock = big.NewInt(0)
	}
	if c.PetersburgBlock == nil {
		c.PetersburgBlock = big.NewInt(0)
	}
	if c.IstanbulBlock == nil {
		c.IstanbulBlock = big.NewInt(0)
	}
	if c.MuirGlacierBlock == nil {
		c.MuirGlacierBlock = big.NewInt(0)
	}

	c.NetworkUpgrades.setDefaults(c.SnowCtx.NetworkID)

	// if c.CancunTime == nil {
	// 	c.CancunTime = c.EUpgrade
	// }
}

// IsHomestead returns whether num is either equal to the homestead block or greater.
func (c *ChainConfig) IsHomestead(num *big.Int) bool {
	return utils.IsBlockForked(c.HomesteadBlock, num)
}

// IsEIP150 returns whether num is either equal to the EIP150 fork block or greater.
func (c *ChainConfig) IsEIP150(num *big.Int) bool {
	return utils.IsBlockForked(c.EIP150Block, num)
}

// IsEIP155 returns whether num is either equal to the EIP155 fork block or greater.
func (c *ChainConfig) IsEIP155(num *big.Int) bool {
	return utils.IsBlockForked(c.EIP155Block, num)
}

// IsEIP158 returns whether num is either equal to the EIP158 fork block or greater.
func (c *ChainConfig) IsEIP158(num *big.Int) bool {
	return utils.IsBlockForked(c.EIP158Block, num)
}

// IsByzantium returns whether num is either equal to the Byzantium fork block or greater.
func (c *ChainConfig) IsByzantium(num *big.Int) bool {
	return utils.IsBlockForked(c.ByzantiumBlock, num)
}

// IsConstantinople returns whether num is either equal to the Constantinople fork block or greater.
func (c *ChainConfig) IsConstantinople(num *big.Int) bool {
	return utils.IsBlockForked(c.ConstantinopleBlock, num)
}

// IsMuirGlacier returns whether num is either equal to the Muir Glacier (EIP-2384) fork block or greater.
func (c *ChainConfig) IsMuirGlacier(num *big.Int) bool {
	return utils.IsBlockForked(c.MuirGlacierBlock, num)
}

// IsPetersburg returns whether num is either
// - equal to or greater than the PetersburgBlock fork block,
// - OR is nil, and Constantinople is active
func (c *ChainConfig) IsPetersburg(num *big.Int) bool {
	return utils.IsBlockForked(c.PetersburgBlock, num) || c.PetersburgBlock == nil && utils.IsBlockForked(c.ConstantinopleBlock, num)
}

// IsIstanbul returns whether num is either equal to the Istanbul fork block or greater.
func (c *ChainConfig) IsIstanbul(num *big.Int) bool {
	return utils.IsBlockForked(c.IstanbulBlock, num)
}

// IsSubnetEVM returns whether [time] represents a block
// with a timestamp after the SubnetEVM upgrade time.
func (c *ChainConfig) IsSubnetEVM(time uint64) bool {
	return utils.IsTimestampForked(c.SubnetEVMTimestamp, time)
}

// TODO: move avalanche hardforks to network_upgrades.go
// IsDurango returns whether [time] represents a block
// with a timestamp after the Durango upgrade time.
func (c *ChainConfig) IsDurango(time uint64) bool {
	return utils.IsTimestampForked(c.DurangoTimestamp, time)
}

// IsCancun returns whether [time] represents a block
// with a timestamp after the Cancun upgrade time.
func (c *ChainConfig) IsCancun(num *big.Int, time uint64) bool {
	return utils.IsTimestampForked(c.CancunTime, time)
}

func (r *Rules) PredicatersExist() bool {
	return len(r.Predicaters) > 0
}

func (r *Rules) PredicaterExists(addr common.Address) bool {
	_, PredicaterExists := r.Predicaters[addr]
	return PredicaterExists
}

// IsPrecompileEnabled returns whether precompile with [address] is enabled at [timestamp].
func (c *ChainConfig) IsPrecompileEnabled(address common.Address, timestamp uint64) bool {
	config := c.getActivePrecompileConfig(address, timestamp)
	return config != nil && !config.IsDisabled()
}

// CheckCompatible checks whether scheduled fork transitions have been imported
// with a mismatching chain configuration.
func (c *ChainConfig) CheckCompatible(newcfg *ChainConfig, height uint64, time uint64) *ConfigCompatError {
	var (
		bhead = new(big.Int).SetUint64(height)
		btime = time
	)
	// Iterate checkCompatible to find the lowest conflict.
	var lasterr *ConfigCompatError
	for {
		err := c.checkCompatible(newcfg, bhead, btime)
		if err == nil || (lasterr != nil && err.RewindToBlock == lasterr.RewindToBlock && err.RewindToTime == lasterr.RewindToTime) {
			break
		}
		lasterr = err

		if err.RewindToTime > 0 {
			btime = err.RewindToTime
		} else {
			bhead.SetUint64(err.RewindToBlock)
		}
	}
	return lasterr
}

// Verify verifies chain config and returns error
func (c *ChainConfig) Verify() error {
	if err := c.FeeConfig.Verify(); err != nil {
		return err
	}

	// Verify the precompile upgrades are internally consistent given the existing chainConfig.
	if err := c.verifyPrecompileUpgrades(); err != nil {
		return fmt.Errorf("invalid precompile upgrades: %w", err)
	}

	// Verify the state upgrades are internally consistent given the existing chainConfig.
	if err := c.verifyStateUpgrades(); err != nil {
		return fmt.Errorf("invalid state upgrades: %w", err)
	}

	// Verify the network upgrades are internally consistent given the existing chainConfig.
	if err := c.VerifyNetworkUpgrades(c.SnowCtx.NetworkID); err != nil {
		return fmt.Errorf("invalid network upgrades: %w", err)
	}

	return nil
}

type fork struct {
	name      string
	block     *big.Int // some go-ethereum forks use block numbers
	timestamp *uint64  // Avalanche forks use timestamps
	optional  bool     // if true, the fork may be nil and next fork is still allowed
}

// CheckConfigForkOrder checks that we don't "skip" any forks, geth isn't pluggable enough
// to guarantee that forks can be implemented in a different order than on official networks
func (c *ChainConfig) CheckConfigForkOrder() error {
	ethForks := []fork{
		{name: "homesteadBlock", block: c.HomesteadBlock},
		{name: "eip150Block", block: c.EIP150Block},
		{name: "eip155Block", block: c.EIP155Block},
		{name: "eip158Block", block: c.EIP158Block},
		{name: "byzantiumBlock", block: c.ByzantiumBlock},
		{name: "constantinopleBlock", block: c.ConstantinopleBlock},
		{name: "petersburgBlock", block: c.PetersburgBlock},
		{name: "istanbulBlock", block: c.IstanbulBlock},
		{name: "muirGlacierBlock", block: c.MuirGlacierBlock, optional: true},
	}

	// Check that forks are enabled in order
	if err := checkForks(ethForks, true); err != nil {
		return err
	}

	// Note: In Avalanche, hard forks must take place via block timestamps instead
	// of block numbers since blocks are produced asynchronously. Therefore, we do not
	// check that the block timestamps in the same way as for
	// the block number forks since it would not be a meaningful comparison.
	// Instead, we check only that Phases are enabled in order.
	// Note: we do not add the optional stateful precompile configs in here because they are optional
	// and independent, such that the ordering they are enabled does not impact the correctness of the
	// chain config.
	if err := checkForks(c.forkOrder(), false); err != nil {
		return err
	}

	return nil
}

// checkForks checks that forks are enabled in order and returns an error if not
// [blockFork] is true if the fork is a block number fork, false if it is a timestamp fork
func checkForks(forks []fork, blockFork bool) error {
	lastFork := fork{}
	for _, cur := range forks {
		if blockFork && cur.block != nil && common.Big0.Cmp(cur.block) != 0 {
			return errNonGenesisForkByHeight
		}
		if lastFork.name != "" {
			// Next one must be higher number
			if lastFork.timestamp == nil && cur.timestamp != nil {
				return fmt.Errorf("unsupported fork ordering: %v not enabled, but %v enabled at %v",
					lastFork.name, cur.name, *cur.timestamp)
			}
			if lastFork.timestamp != nil && cur.timestamp != nil {
				if *lastFork.timestamp > *cur.timestamp {
					return fmt.Errorf("unsupported fork ordering: %v enabled at %v, but %v enabled at %v",
						lastFork.name, *lastFork.timestamp, cur.name, *cur.timestamp)
				}
			}
		}
		// If it was optional and not set, then ignore it
		if !cur.optional || cur.timestamp != nil {
			lastFork = cur
		}
	}
	return nil
}

// checkCompatible confirms that [newcfg] is backwards compatible with [c] to upgrade with the given head block height and timestamp.
// This confirms that all Ethereum and Avalanche upgrades are backwards compatible as well as that the precompile config is backwards
// compatible.
func (c *ChainConfig) checkCompatible(newcfg *ChainConfig, height *big.Int, time uint64) *ConfigCompatError {
	if isForkBlockIncompatible(c.HomesteadBlock, newcfg.HomesteadBlock, height) {
		return newBlockCompatError("Homestead fork block", c.HomesteadBlock, newcfg.HomesteadBlock)
	}
	if isForkBlockIncompatible(c.EIP150Block, newcfg.EIP150Block, height) {
		return newBlockCompatError("EIP150 fork block", c.EIP150Block, newcfg.EIP150Block)
	}
	if isForkBlockIncompatible(c.EIP155Block, newcfg.EIP155Block, height) {
		return newBlockCompatError("EIP155 fork block", c.EIP155Block, newcfg.EIP155Block)
	}
	if isForkBlockIncompatible(c.EIP158Block, newcfg.EIP158Block, height) {
		return newBlockCompatError("EIP158 fork block", c.EIP158Block, newcfg.EIP158Block)
	}
	if c.IsEIP158(height) && !utils.BigNumEqual(c.ChainID, newcfg.ChainID) {
		return newBlockCompatError("EIP158 chain ID", c.EIP158Block, newcfg.EIP158Block)
	}
	if isForkBlockIncompatible(c.ByzantiumBlock, newcfg.ByzantiumBlock, height) {
		return newBlockCompatError("Byzantium fork block", c.ByzantiumBlock, newcfg.ByzantiumBlock)
	}
	if isForkBlockIncompatible(c.ConstantinopleBlock, newcfg.ConstantinopleBlock, height) {
		return newBlockCompatError("Constantinople fork block", c.ConstantinopleBlock, newcfg.ConstantinopleBlock)
	}
	if isForkBlockIncompatible(c.PetersburgBlock, newcfg.PetersburgBlock, height) {
		// the only case where we allow Petersburg to be set in the past is if it is equal to Constantinople
		// mainly to satisfy fork ordering requirements which state that Petersburg fork be set if Constantinople fork is set
		if isForkBlockIncompatible(c.ConstantinopleBlock, newcfg.PetersburgBlock, height) {
			return newBlockCompatError("Petersburg fork block", c.PetersburgBlock, newcfg.PetersburgBlock)
		}
	}
	if isForkBlockIncompatible(c.IstanbulBlock, newcfg.IstanbulBlock, height) {
		return newBlockCompatError("Istanbul fork block", c.IstanbulBlock, newcfg.IstanbulBlock)
	}
	if isForkBlockIncompatible(c.MuirGlacierBlock, newcfg.MuirGlacierBlock, height) {
		return newBlockCompatError("Muir Glacier fork block", c.MuirGlacierBlock, newcfg.MuirGlacierBlock)
	}

	if isForkTimestampIncompatible(c.CancunTime, newcfg.CancunTime, time) {
		return newTimestampCompatError("Cancun fork block timestamp", c.CancunTime, c.CancunTime)
	}

	// Check avalanche network upgrades
	if err := c.CheckNetworkUpgradesCompatible(&newcfg.NetworkUpgrades, time); err != nil {
		return err
	}

	// Check that the precompiles on the new config are compatible with the existing precompile config.
	if err := c.CheckPrecompilesCompatible(newcfg.PrecompileUpgrades, time); err != nil {
		return err
	}

	// Check that the state upgrades on the new config are compatible with the existing state upgrade config.
	if err := c.CheckStateUpgradesCompatible(newcfg.StateUpgrades, time); err != nil {
		return err
	}

	// TODO verify that the fee config is fully compatible between [c] and [newcfg].
	return nil
}

// isForkBlockIncompatible returns true if a fork scheduled at s1 cannot be rescheduled to
// block s2 because head is already past the fork.
func isForkBlockIncompatible(s1, s2, head *big.Int) bool {
	return (utils.IsBlockForked(s1, head) || utils.IsBlockForked(s2, head)) && !configBlockEqual(s1, s2)
}

func configBlockEqual(x, y *big.Int) bool {
	if x == nil {
		return y == nil
	}
	if y == nil {
		return x == nil
	}
	return x.Cmp(y) == 0
}

// isForkTimestampIncompatible returns true if a fork scheduled at timestamp s1
// cannot be rescheduled to timestamp s2 because head is already past the fork.
func isForkTimestampIncompatible(s1, s2 *uint64, head uint64) bool {
	return (utils.IsTimestampForked(s1, head) || utils.IsTimestampForked(s2, head)) && !configTimestampEqual(s1, s2)
}

func configTimestampEqual(x, y *uint64) bool {
	if x == nil {
		return y == nil
	}
	if y == nil {
		return x == nil
	}
	return *x == *y
}

// ConfigCompatError is raised if the locally-stored blockchain is initialised with a
// ChainConfig that would alter the past.
type ConfigCompatError struct {
	What string

	// block numbers of the stored and new configurations if block based forking
	StoredBlock, NewBlock *big.Int

	// timestamps of the stored and new configurations if time based forking
	StoredTime, NewTime *uint64

	// the block number to which the local chain must be rewound to correct the error
	RewindToBlock uint64

	// the timestamp to which the local chain must be rewound to correct the error
	RewindToTime uint64
}

func newBlockCompatError(what string, storedblock, newblock *big.Int) *ConfigCompatError {
	var rew *big.Int
	switch {
	case storedblock == nil:
		rew = newblock
	case newblock == nil || storedblock.Cmp(newblock) < 0:
		rew = storedblock
	default:
		rew = newblock
	}
	err := &ConfigCompatError{
		What:          what,
		StoredBlock:   storedblock,
		NewBlock:      newblock,
		RewindToBlock: 0,
	}
	if rew != nil && rew.Sign() > 0 {
		err.RewindToBlock = rew.Uint64() - 1
	}
	return err
}

func newTimestampCompatError(what string, storedtime, newtime *uint64) *ConfigCompatError {
	var rew *uint64
	switch {
	case storedtime == nil:
		rew = newtime
	case newtime == nil || *storedtime < *newtime:
		rew = storedtime
	default:
		rew = newtime
	}
	err := &ConfigCompatError{
		What:         what,
		StoredTime:   storedtime,
		NewTime:      newtime,
		RewindToTime: 0,
	}
	if rew != nil && *rew > 0 {
		err.RewindToTime = *rew - 1
	}
	return err
}

func (err *ConfigCompatError) Error() string {
	if err.StoredBlock != nil {
		return fmt.Sprintf("mismatching %s in database (have block %d, want block %d, rewindto block %d)", err.What, err.StoredBlock, err.NewBlock, err.RewindToBlock)
	}
	return fmt.Sprintf("mismatching %s in database (have timestamp %s, want timestamp %s, rewindto timestamp %d)", err.What, ptrToString(err.StoredTime), ptrToString(err.NewTime), err.RewindToTime)
}

func ptrToString(val *uint64) string {
	if val == nil {
		return "nil"
	}
	return fmt.Sprintf("%d", *val)
}

// Rules wraps ChainConfig and is merely syntactic sugar or can be used for functions
// that do not have or require information about the block.
//
// Rules is a one time interface meaning that it shouldn't be used in between transition
// phases.
type Rules struct {
	ChainID                                                 *big.Int
	IsHomestead, IsEIP150, IsEIP155, IsEIP158               bool
	IsByzantium, IsConstantinople, IsPetersburg, IsIstanbul bool
	IsCancun                                                bool

	// Rules for Avalanche releases
	IsSubnetEVM bool
	IsDurango   bool

	// ActivePrecompiles maps addresses to stateful precompiled contracts that are enabled
	// for this rule set.
	// Note: none of these addresses should conflict with the address space used by
	// any existing precompiles.
	ActivePrecompiles map[common.Address]precompileconfig.Config
	// Predicaters maps addresses to stateful precompile Predicaters
	// that are enabled for this rule set.
	Predicaters map[common.Address]precompileconfig.Predicater
	// AccepterPrecompiles map addresses to stateful precompile accepter functions
	// that are enabled for this rule set.
	AccepterPrecompiles map[common.Address]precompileconfig.Accepter
}

// IsPrecompileEnabled returns true if the precompile at [addr] is enabled for this rule set.
func (r *Rules) IsPrecompileEnabled(addr common.Address) bool {
	_, ok := r.ActivePrecompiles[addr]
	return ok
}

// Rules ensures c's ChainID is not nil.
func (c *ChainConfig) rules(num *big.Int, timestamp uint64) Rules {
	chainID := c.ChainID
	if chainID == nil {
		chainID = new(big.Int)
	}
	return Rules{
		ChainID:          new(big.Int).Set(chainID),
		IsHomestead:      c.IsHomestead(num),
		IsEIP150:         c.IsEIP150(num),
		IsEIP155:         c.IsEIP155(num),
		IsEIP158:         c.IsEIP158(num),
		IsByzantium:      c.IsByzantium(num),
		IsConstantinople: c.IsConstantinople(num),
		IsPetersburg:     c.IsPetersburg(num),
		IsIstanbul:       c.IsIstanbul(num),
		IsCancun:         c.IsCancun(num, timestamp),
	}
}

// Rules returns the Avalanche modified rules to support Avalanche
// network upgrades
func (c *ChainConfig) Rules(blockNum *big.Int, timestamp uint64) Rules {
	rules := c.rules(blockNum, timestamp)

	rules.IsSubnetEVM = c.IsSubnetEVM(timestamp)
	rules.IsDurango = c.IsDurango(timestamp)

	// Initialize the stateful precompiles that should be enabled at [blockTimestamp].
	rules.ActivePrecompiles = make(map[common.Address]precompileconfig.Config)
	rules.Predicaters = make(map[common.Address]precompileconfig.Predicater)
	rules.AccepterPrecompiles = make(map[common.Address]precompileconfig.Accepter)
	for _, module := range modules.RegisteredModules() {
		if config := c.getActivePrecompileConfig(module.Address, timestamp); config != nil && !config.IsDisabled() {
			rules.ActivePrecompiles[module.Address] = config
			if predicater, ok := config.(precompileconfig.Predicater); ok {
				rules.Predicaters[module.Address] = predicater
			}
			if precompileAccepter, ok := config.(precompileconfig.Accepter); ok {
				rules.AccepterPrecompiles[module.Address] = precompileAccepter
			}
		}
	}

	return rules
}

// GetFeeConfig returns the original FeeConfig contained in the genesis ChainConfig.
// Implements precompile.ChainConfig interface.
func (c *ChainConfig) GetFeeConfig() commontype.FeeConfig {
	return c.FeeConfig
}

// AllowedFeeRecipients returns the original AllowedFeeRecipients parameter contained in the genesis ChainConfig.
// Implements precompile.ChainConfig interface.
func (c *ChainConfig) AllowedFeeRecipients() bool {
	return c.AllowFeeRecipients
}
