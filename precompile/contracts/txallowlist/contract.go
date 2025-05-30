// (c) 2019-2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package txallowlist

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/shubhamdubey02/subnet-evm/precompile/allowlist"
	"github.com/shubhamdubey02/subnet-evm/precompile/contract"
)

// Singleton StatefulPrecompiledContract for W/R access to the tx allow list.
var TxAllowListPrecompile contract.StatefulPrecompiledContract = allowlist.CreateAllowListPrecompile(ContractAddress)

// GetTxAllowListStatus returns the role of [address] for the tx allow list.
func GetTxAllowListStatus(stateDB contract.StateDB, address common.Address) allowlist.Role {
	return allowlist.GetAllowListStatus(stateDB, ContractAddress, address)
}

// SetTxAllowListStatus sets the permissions of [address] to [role] for the
// tx allow list.
// assumes [role] has already been verified as valid.
func SetTxAllowListStatus(stateDB contract.StateDB, address common.Address, role allowlist.Role) {
	allowlist.SetAllowListRole(stateDB, ContractAddress, address, role)
}
