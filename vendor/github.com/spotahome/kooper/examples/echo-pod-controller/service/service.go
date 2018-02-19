package service

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/spotahome/kooper/examples/echo-pod-controller/log"
)

// Echo is simple echo service.
type Echo interface {
	// EchoObj echoes the received object.
	EchoObj(prefix string, obj runtime.Object)
	// EchoS echoes the received string.
	EchoS(prefix string, s string)
}

// SimpleEcho echoes the received object name.
type SimpleEcho struct {
	logger log.Logger
}

// NewSimpleEcho returns a new SimpleEcho.
func NewSimpleEcho(logger log.Logger) *SimpleEcho {
	return &SimpleEcho{
		logger: logger,
	}
}

func (s *SimpleEcho) getObjInfo(obj runtime.Object) (string, error) {
	objMeta, ok := obj.(metav1.Object)
	if !ok {
		return "", fmt.Errorf("could not print object information")
	}
	return fmt.Sprintf("%s", objMeta.GetName()), nil
}

func (s *SimpleEcho) echo(prefix string, str string) {
	s.logger.Infof("[%s] %s", prefix, str)
}

// EchoObj satisfies service.Echo interface.
func (s *SimpleEcho) EchoObj(prefix string, obj runtime.Object) {
	// Get object string with all the information.
	objInfo, err := s.getObjInfo(obj)
	if err != nil {
		s.logger.Errorf("error on echo: %s", err)
	}

	s.echo(prefix, objInfo)
}

// EchoS satisfies service.Echo interface.
func (s *SimpleEcho) EchoS(prefix string, str string) {
	s.echo(prefix, str)
}
