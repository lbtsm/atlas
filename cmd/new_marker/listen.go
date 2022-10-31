package new_marker

import (
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/mapprotocol/atlas/cmd/marker/config"
)

type listener struct {
	cfg    *config.Config
	conn   *ethclient.Client
	writer Writer
	msgCh  chan struct{} // wait for msg handles
}
