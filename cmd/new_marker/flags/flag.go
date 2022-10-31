package flags

import (
	"gopkg.in/urfave/cli.v1"
)

var (
	Key = cli.StringFlag{
		Name:  "key",
		Usage: "Private key file path",
		Value: "",
	}
	KeyStore = cli.StringFlag{
		Name:  "keystore",
		Usage: "Keystore file path",
	}
	Name = cli.StringFlag{
		Name:  "name",
		Usage: "name of account",
	}
	URL = cli.StringFlag{
		Name:  "url",
		Usage: "metadata url of account",
	}
	Commission = cli.Uint64Flag{
		Name:  "commission",
		Usage: "register validator param",
	}
	Relayerf = cli.StringFlag{
		Name:  "relayerf",
		Usage: "register validator param",
	}
	VoteNum = cli.Int64Flag{
		Name:  "voteNum",
		Usage: "The amount of gold to use to vote",
	}
	TopNum = cli.Int64Flag{
		Name:  "topNum",
		Usage: "topNum of validator",
	}
	LockedNum = cli.Int64Flag{
		Name:  "lockedNum",
		Usage: "The amount of map to lock 、unlock、relock、withdraw ",
	}
	MAPValue = cli.Int64Flag{
		Name:  "mapValue",
		Usage: "validator address",
	}
	WithdrawIndex = cli.Int64Flag{
		Name:  "withdrawIndex",
		Usage: "use for withdraw",
	}
	RelockIndex = cli.Int64Flag{
		Name:  "relockIndex",
		Usage: "use for relock",
	}

	Verbosity = cli.Int64Flag{
		Name:  "Verbosity",
		Usage: "Verbosity of log level",
	}

	RPCListenAddr = cli.StringFlag{
		Name:  "rpcaddr",
		Usage: "HTTP-RPC server listening interface",
		Value: "localhost",
	}
	Value = cli.Uint64Flag{
		Name:  "value",
		Usage: "value units one eth",
		Value: 0,
	}
	Amount = cli.StringFlag{
		Name:  "amount",
		Usage: "transfer amount, unit (wei)",
		Value: "0",
	}
	Duration = cli.Int64Flag{
		Name:  "duration",
		Usage: "duration The time (in seconds) that these requirements persist for.",
		Value: 0,
	}
	TargetAddress = cli.StringFlag{
		Name:  "target",
		Usage: "Target query address",
		Value: "",
	}

	ValidatorAddress = cli.StringFlag{
		Name:  "validator",
		Usage: "validator address",
		Value: "",
	}
	SignerPriv = cli.StringFlag{
		Name:  "signerPriv",
		Usage: "signer private",
		Value: "",
	}
	Signer = cli.StringFlag{
		Name:  "signer",
		Usage: "signer address",
		Value: "",
	}
	Signature = cli.StringFlag{
		Name:  "signature",
		Usage: "ECDSA Signature",
		Value: "",
	}
	Proof = cli.StringFlag{
		Name:  "proof",
		Usage: "signer proof",
		Value: "",
	}
	AccountAddress = cli.StringFlag{
		Name:  "accountAddress",
		Usage: "account address",
		Value: "",
	}
	ContractAddress = cli.StringFlag{
		Name:  "contractAddress",
		Usage: "set contract Address",
		Value: "",
	}
	ImplementationAddress = cli.StringFlag{
		Name:  "implementationAddress",
		Usage: "set implementation Address",
		Value: "",
	}
	GasLimit = cli.Int64Flag{
		Name:  "gasLimit",
		Usage: "use for sendContractTransaction gasLimit",
		Value: 0,
	}
)

var (
	Default = []cli.Flag{
		KeyStore,
		GasLimit,
		RPCListenAddr,
	}
)
