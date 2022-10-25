package constant

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/mapprotocol/atlas/cmd/marker/mapprotocol"
)

var (
	AddressOfValidator          = mapprotocol.MustProxyAddressFor("Validators")
	AddressOfLockedGold         = mapprotocol.MustProxyAddressFor("LockedGold")
	AddressOfAccounts           = mapprotocol.MustProxyAddressFor("Accounts")
	AddressOfElection           = mapprotocol.MustProxyAddressFor("Election")
	AddressOfGoldToken          = mapprotocol.MustProxyAddressFor("GoldToken")
	AddressOfEpochRewards       = mapprotocol.MustProxyAddressFor("EpochRewards")
	AddressOfTestPoc2Parameters = common.HexToAddress("0xb586DC60e9e39F87c9CB8B7D7E30b2f04D40D14c")
)

var (
	AbiOfValidators         = mapprotocol.AbiFor("Validators")
	AbiOfLockedGold         = mapprotocol.AbiFor("LockedGold")
	AbiOfAccounts           = mapprotocol.AbiFor("Accounts")
	AbiOfElection           = mapprotocol.AbiFor("Election")
	AbiOfGoldToken          = mapprotocol.AbiFor("GoldToken")
	AbiOfEpochRewards       = mapprotocol.AbiFor("EpochRewards")
	AbiOfTestPoc2Parameters = mapprotocol.AbiFor("TestPoc2")
)
