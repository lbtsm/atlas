package new_marker

import (
	"crypto/ecdsa"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/mapprotocol/atlas/accounts/abi"
)

type Message struct {
	from        common.Address
	priKey      *ecdsa.PrivateKey
	value       *big.Int
	messageType string
	input       []byte
	abiMethod   string
	to          common.Address
	abi         *abi.ABI
	DoneCh      chan<- struct{}
	ret         interface{}
	solveResult func([]byte)
	gasLimit    uint64
}
