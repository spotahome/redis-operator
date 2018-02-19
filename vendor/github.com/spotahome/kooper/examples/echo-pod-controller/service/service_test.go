package service_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/spotahome/kooper/examples/echo-pod-controller/service"
)

type logKind int

const (
	infoKind logKind = iota
	warnignKind
	errorKind
)

type logEvent struct {
	kind logKind
	line string
}

type testLogger struct {
	events []logEvent
	sync.Mutex
}

func (t *testLogger) logLine(kind logKind, format string, args ...interface{}) {
	str := fmt.Sprintf(format, args...)
	t.events = append(t.events, logEvent{kind: kind, line: str})
}

func (t *testLogger) Infof(format string, args ...interface{}) {
	t.logLine(infoKind, format, args...)
}
func (t *testLogger) Warningf(format string, args ...interface{}) {
	t.logLine(warnignKind, format, args...)
}
func (t *testLogger) Errorf(format string, args ...interface{}) {
	t.logLine(errorKind, format, args...)
}

func TestEchoServiceEchoString(t *testing.T) {
	tests := []struct {
		name       string
		prefix     string
		msg        string
		expResults []logEvent
	}{
		{
			name:   "Logging a prefix and a string should log.",
			prefix: "test",
			msg:    "this is a test",
			expResults: []logEvent{
				logEvent{kind: infoKind, line: "[test] this is a test"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mocks.
			ml := &testLogger{events: []logEvent{}}

			// Create aservice and run.
			srv := service.NewSimpleEcho(ml)
			srv.EchoS(test.prefix, test.msg)

			// Check.
			assert.Equal(test.expResults, ml.events)
		})
	}
}

func TestEchoServiceEchoObj(t *testing.T) {
	tests := []struct {
		name       string
		prefix     string
		obj        runtime.Object
		expResults []logEvent
	}{
		{
			name:   "Logging a pod should print pod name.",
			prefix: "test",
			obj: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: "mypod",
				},
			},
			expResults: []logEvent{
				logEvent{kind: infoKind, line: "[test] mypod"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mocks.
			ml := &testLogger{events: []logEvent{}}

			// Create aservice and run.
			srv := service.NewSimpleEcho(ml)
			srv.EchoObj(test.prefix, test.obj)

			// Check.
			assert.Equal(test.expResults, ml.events)
		})
	}
}
