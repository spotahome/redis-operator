package restart

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	v1 "github.com/spotahome/redis-operator/api/redisfailover/v1"
	clientset "github.com/spotahome/redis-operator/client/k8s/clientset/versioned/typed/redisfailover/v1"
	"github.com/spotahome/redis-operator/pkg/kubectl-redis-failover/options"
)

const (
	restartExample = `
	# Restart the pods of a RedisFailover
	%[1]s restart redis REDISFAILOVER_NAME

	# Restart only redis pods of a RedisFailover
	%[1]s restart redis REDISFAILOVER_NAME --redis

	# Restart only sentinel pods of a RedisFailover
	%[1]s restart redis REDISFAILOVER_NAME --sentinel
	`

	redisRestartPatch = `{
		"spec": {
			"redis": {
				"restartAt": "%s"
			}
		}
	}`
	sentinelRestartPatch = `{
		"spec": {
			"sentinel": {
				"restartAt": "%s"
			}
		}
	}`
)

func NewCmdRestartFailover(o *options.RedisFailoverOptions) *cobra.Command {
	var (
		redis    bool
		sentinel bool
	)
	var cmd = &cobra.Command{
		Use:          "restart REDISFAILOVER",
		Short:        "Restart the redis pods of a RedisFailover",
		Example:      o.Example(restartExample),
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) != 1 {
				return o.UsageErr(c)
			}
			restartAt := o.Now().UTC()
			name := args[0]
			redisfailoverIf := o.RedisFailoversClientset().DatabasesV1().RedisFailovers(o.Namespace())
			if (!redis && !sentinel) || (redis && sentinel) {
				if _, errR := RestartRedisesOfRedisFailover(redisfailoverIf, name, &restartAt); errR != nil {
					return errR
				}
				rf, errS := RestartSentinelsOfRedisFailover(redisfailoverIf, name, &restartAt)
				if errS != nil {
					return errS
				}
				fmt.Fprintf(o.Out, "redisfailover '%s' restarted\n", rf.Name)
			} else if redis {
				rf, err := RestartRedisesOfRedisFailover(redisfailoverIf, name, &restartAt)
				if err != nil {
					return err
				}
				fmt.Fprintf(o.Out, "redises of redisfailover '%s' restarted\n", rf.Name)
			} else if sentinel {
				rf, err := RestartSentinelsOfRedisFailover(redisfailoverIf, name, &restartAt)
				if err != nil {
					return err
				}
				fmt.Fprintf(o.Out, "sentinels of redisfailover '%s' restarted\n", rf.Name)
			}
			return nil
		},
	}
	cmd.Flags().BoolVarP(&redis, "redis", "", false, "Restart only redis pods")
	cmd.Flags().BoolVarP(&sentinel, "sentinel", "", false, "Restart only sentinel pods")
	return cmd
}

// RestartRedisesOfRedisFailover restarts the redis pods of a RedisFailover.
func RestartRedisesOfRedisFailover(redisfailoverIf clientset.RedisFailoverInterface, name string, restartAt *time.Time) (*v1.RedisFailover, error) {
	ctx := context.TODO()
	if restartAt == nil {
		t := time.Now().UTC()
		restartAt = &t
	}
	patch := fmt.Sprintf(redisRestartPatch, restartAt.Format(time.RFC3339))
	return redisfailoverIf.Patch(ctx, name, types.MergePatchType, []byte(patch), metav1.PatchOptions{})
}

// RestartSentinelsOfRedisFailover restarts the redis pods of a RedisFailover.
func RestartSentinelsOfRedisFailover(redisfailoverIf clientset.RedisFailoverInterface, name string, restartAt *time.Time) (*v1.RedisFailover, error) {
	ctx := context.TODO()
	if restartAt == nil {
		t := time.Now().UTC()
		restartAt = &t
	}
	patch := fmt.Sprintf(sentinelRestartPatch, restartAt.Format(time.RFC3339))
	return redisfailoverIf.Patch(ctx, name, types.MergePatchType, []byte(patch), metav1.PatchOptions{})
}
