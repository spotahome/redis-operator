package restart

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kubetesting "k8s.io/client-go/testing"

	v1 "github.com/spotahome/redis-operator/api/redisfailover/v1"
	fakeclientset "github.com/spotahome/redis-operator/client/k8s/clientset/versioned/fake"
	options "github.com/spotahome/redis-operator/pkg/kubectl-redis-failover/options/fake"
)

func TestRestartCmdUsage(t *testing.T) {
	tf, o := options.NewFakeRedisFailoverOptions()
	defer tf.Cleanup()
	cmd := NewCmdRestartFailover(o)
	cmd.PersistentPreRunE = o.PersistentPreRunE
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	assert.Error(t, err)
	stdout := o.Out.(*bytes.Buffer).String()
	stderr := o.ErrOut.(*bytes.Buffer).String()
	assert.Empty(t, stdout)
	assert.Contains(t, stderr, "Usage:")
	assert.Contains(t, stderr, "restart REDISFAILOVER")
}

func TestRestartCmdOnlyRedisSuccess(t *testing.T) {
	rf := v1.RedisFailover{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: metav1.NamespaceDefault,
		},
	}

	tf, o := options.NewFakeRedisFailoverOptions(&rf)
	defer tf.Cleanup()
	now := metav1.Now()
	o.Now = func() metav1.Time {
		return now
	}
	fakeClient := o.RedisFailoverClient.(*fakeclientset.Clientset)
	fakeClient.PrependReactor("patch", "*", func(action kubetesting.Action) (handled bool, ret runtime.Object, err error) {
		if patchAction, ok := action.(kubetesting.PatchAction); ok {
			if string(patchAction.GetPatch()) == fmt.Sprintf(redisRestartPatch, now.UTC().Format(time.RFC3339)) {
				rf.Spec.Redis.RestartAt = now.DeepCopy()
			}
		}
		return true, &rf, nil
	})

	cmd := NewCmdRestartFailover(o)
	cmd.PersistentPreRunE = o.PersistentPreRunE
	o.AddKubectlFlags(cmd)
	cmd.SetArgs([]string{"test", "--redis"})
	err := cmd.Execute()
	assert.Nil(t, err)

	expectedTime := metav1.NewTime(now.UTC())
	assert.True(t, rf.Spec.Redis.RestartAt.Equal(&expectedTime))
	stdout := o.Out.(*bytes.Buffer).String()
	stderr := o.ErrOut.(*bytes.Buffer).String()
	assert.Equal(t, "redises of redisfailover 'test' restarted\n", stdout)
	assert.Empty(t, stderr)
}

func TestRestartCmdOnlySentinelSuccess(t *testing.T) {
	rf := v1.RedisFailover{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: metav1.NamespaceDefault,
		},
	}

	tf, o := options.NewFakeRedisFailoverOptions(&rf)
	defer tf.Cleanup()
	now := metav1.Now()
	o.Now = func() metav1.Time {
		return now
	}
	fakeClient := o.RedisFailoverClient.(*fakeclientset.Clientset)
	fakeClient.PrependReactor("patch", "*", func(action kubetesting.Action) (handled bool, ret runtime.Object, err error) {
		if patchAction, ok := action.(kubetesting.PatchAction); ok {
			if string(patchAction.GetPatch()) == fmt.Sprintf(sentinelRestartPatch, now.UTC().Format(time.RFC3339)) {
				rf.Spec.Sentinel.RestartAt = now.DeepCopy()
			}
		}
		return true, &rf, nil
	})

	cmd := NewCmdRestartFailover(o)
	cmd.PersistentPreRunE = o.PersistentPreRunE
	o.AddKubectlFlags(cmd)
	cmd.SetArgs([]string{"test", "--sentinel"})
	err := cmd.Execute()
	assert.Nil(t, err)

	expectedTime := metav1.NewTime(now.UTC())
	assert.True(t, rf.Spec.Sentinel.RestartAt.Equal(&expectedTime))
	stdout := o.Out.(*bytes.Buffer).String()
	stderr := o.ErrOut.(*bytes.Buffer).String()
	assert.Equal(t, "sentinels of redisfailover 'test' restarted\n", stdout)
	assert.Empty(t, stderr)
}

func TestRestartCmdAllWithoutFlagsSuccess(t *testing.T) {
	rf := v1.RedisFailover{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: metav1.NamespaceDefault,
		},
	}

	tf, o := options.NewFakeRedisFailoverOptions(&rf)
	defer tf.Cleanup()
	now := metav1.Now()
	o.Now = func() metav1.Time {
		return now
	}
	fakeClient := o.RedisFailoverClient.(*fakeclientset.Clientset)
	fakeClient.PrependReactor("patch", "*", func(action kubetesting.Action) (handled bool, ret runtime.Object, err error) {
		if patchAction, ok := action.(kubetesting.PatchAction); ok {
			if string(patchAction.GetPatch()) == fmt.Sprintf(sentinelRestartPatch, now.UTC().Format(time.RFC3339)) {
				rf.Spec.Sentinel.RestartAt = now.DeepCopy()
				rf.Spec.Redis.RestartAt = now.DeepCopy()
			}
		}
		return true, &rf, nil
	})

	cmd := NewCmdRestartFailover(o)
	cmd.PersistentPreRunE = o.PersistentPreRunE
	o.AddKubectlFlags(cmd)
	cmd.SetArgs([]string{"test"})
	err := cmd.Execute()
	assert.Nil(t, err)

	expectedTime := metav1.NewTime(now.UTC())
	assert.True(t, rf.Spec.Redis.RestartAt.Equal(&expectedTime))
	assert.True(t, rf.Spec.Sentinel.RestartAt.Equal(&expectedTime))
	stdout := o.Out.(*bytes.Buffer).String()
	stderr := o.ErrOut.(*bytes.Buffer).String()
	assert.Equal(t, "redisfailover 'test' restarted\n", stdout)
	assert.Empty(t, stderr)
}

func TestRestartCmdAllWithFlagsSuccess(t *testing.T) {
	rf := v1.RedisFailover{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: metav1.NamespaceDefault,
		},
	}

	tf, o := options.NewFakeRedisFailoverOptions(&rf)
	defer tf.Cleanup()
	now := metav1.Now()
	o.Now = func() metav1.Time {
		return now
	}
	fakeClient := o.RedisFailoverClient.(*fakeclientset.Clientset)
	fakeClient.PrependReactor("patch", "*", func(action kubetesting.Action) (handled bool, ret runtime.Object, err error) {
		if patchAction, ok := action.(kubetesting.PatchAction); ok {
			if string(patchAction.GetPatch()) == fmt.Sprintf(sentinelRestartPatch, now.UTC().Format(time.RFC3339)) {
				rf.Spec.Sentinel.RestartAt = now.DeepCopy()
				rf.Spec.Redis.RestartAt = now.DeepCopy()
			}
		}
		return true, &rf, nil
	})

	cmd := NewCmdRestartFailover(o)
	cmd.PersistentPreRunE = o.PersistentPreRunE
	o.AddKubectlFlags(cmd)
	cmd.SetArgs([]string{"test", "--redis", "--sentinel"})
	err := cmd.Execute()
	assert.Nil(t, err)

	expectedTime := metav1.NewTime(now.UTC())
	assert.True(t, rf.Spec.Redis.RestartAt.Equal(&expectedTime))
	assert.True(t, rf.Spec.Sentinel.RestartAt.Equal(&expectedTime))
	stdout := o.Out.(*bytes.Buffer).String()
	stderr := o.ErrOut.(*bytes.Buffer).String()
	assert.Equal(t, "redisfailover 'test' restarted\n", stdout)
	assert.Empty(t, stderr)
}

func TestRestartCmdPatchError(t *testing.T) {
	rf := v1.RedisFailover{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: metav1.NamespaceDefault,
		},
	}

	tf, o := options.NewFakeRedisFailoverOptions(&rf)
	defer tf.Cleanup()
	cmd := NewCmdRestartFailover(o)
	o.AddKubectlFlags(cmd)
	cmd.PersistentPreRunE = o.PersistentPreRunE
	cmd.SetArgs([]string{"test"})
	fakeClient := o.RedisFailoverClient.(*fakeclientset.Clientset)
	fakeClient.PrependReactor("patch", "*", func(action kubetesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, fmt.Errorf("Intentional Error")
	})

	err := cmd.Execute()
	assert.Error(t, err)
	stdout := o.Out.(*bytes.Buffer).String()
	stderr := o.ErrOut.(*bytes.Buffer).String()
	assert.Empty(t, stdout)
	assert.Equal(t, "Error: Intentional Error\n", stderr)
}

func TestRestartCmdNotFoundError(t *testing.T) {
	tf, o := options.NewFakeRedisFailoverOptions(&v1.RedisFailover{})
	defer tf.Cleanup()
	cmd := NewCmdRestartFailover(o)
	o.AddKubectlFlags(cmd)
	cmd.PersistentPreRunE = o.PersistentPreRunE
	cmd.SetArgs([]string{"doesnotexist"})
	err := cmd.Execute()
	assert.Error(t, err)
	stdout := o.Out.(*bytes.Buffer).String()
	stderr := o.ErrOut.(*bytes.Buffer).String()
	assert.Empty(t, stdout)
	assert.Equal(t, "Error: redisfailovers.databases.spotahome.com \"doesnotexist\" not found\n", stderr)
}
