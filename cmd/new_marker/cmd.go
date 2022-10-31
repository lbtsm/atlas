package new_marker

import (
	"github.com/mapprotocol/atlas/cmd/new_marker/flags"
	"gopkg.in/urfave/cli.v1"
)

type Command struct {
	cli.Command
}

type Option func(cmd *Command)

type Action func(ctx *cli.Context, config *listener) error

var (
	withAction = func(action Action) Option {
		return func(cmd *Command) {
			cmd.Action = action
		}
	}
	withDefaultFlags = func() Option {
		return func(cmd *Command) {
			cmd.Flags = append(cmd.Flags, flags.Default...)
		}
	}
	withCustomFlags = func(customFlags []cli.Flag) Option {
		return func(cmd *Command) {
			cmd.Flags = append(cmd.Flags, customFlags...)
		}
	}
	withUsage = func(usage string) Option {
		return func(cmd *Command) {
			cmd.Command.Usage = usage
		}
	}
)

func NewCommand() *Command {
	return &Command{
		Command: cli.Command{},
	}
}
