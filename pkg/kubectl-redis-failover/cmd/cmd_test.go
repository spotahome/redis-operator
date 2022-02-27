package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	options "github.com/spotahome/redis-operator/pkg/kubectl-redis-failover/options/fake"
)

func TestCmdArgoRolloutsCmdUsage(t *testing.T) {
	tf, o := options.NewFakeRedisFailoverOptions()
	defer tf.Cleanup()
	cmd := NewCmdRedisFailover(o)
	cmd.PersistentPreRunE = o.PersistentPreRunE
	err := cmd.Execute()
	assert.Error(t, err)
	stdout := o.Out.(*bytes.Buffer).String()
	stderr := o.ErrOut.(*bytes.Buffer).String()
	assert.Empty(t, stdout)
	assert.Contains(t, stderr, "Usage:")
	assert.Contains(t, stderr, "kubectl-redis-failover COMMAND")
}
