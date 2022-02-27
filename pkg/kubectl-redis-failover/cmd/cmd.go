package cmd

import (
	"github.com/spf13/cobra"

	"github.com/spotahome/redis-operator/pkg/kubectl-redis-failover/cmd/restart"
	"github.com/spotahome/redis-operator/pkg/kubectl-redis-failover/options"
)

const (
	example = `
  # Restart thed redis pods of test redisfailover
  %[1]s restart REDISFAILOVER --redis
  `
)

// NewCmdRedisFailover returns new instance of redis failover command.
func NewCmdRedisFailover(o *options.RedisFailoverOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "kubectl-redis-failover COMMAND",
		Short:             "Manage redis failovers",
		Long:              "This command consists of multiple subcommands which can be used to manage Redis Failovers.",
		Example:           o.Example(example),
		PersistentPreRunE: o.PersistentPreRunE,
		RunE: func(c *cobra.Command, args []string) error {
			return o.UsageErr(c)
		},
	}
	o.AddKubectlFlags(cmd)
	cmd.AddCommand(restart.NewCmdRestartFailover(o))

	return cmd
}
