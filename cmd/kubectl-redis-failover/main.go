package main

import (
	"os"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	_ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/klog/v2"

	"github.com/spotahome/redis-operator/pkg/kubectl-redis-failover/cmd"
	"github.com/spotahome/redis-operator/pkg/kubectl-redis-failover/options"
)

func main() {
	klog.InitFlags(nil)
	streams := genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}
	o := options.NewRedisFailoverOptions(streams)
	root := cmd.NewCmdRedisFailover(o)
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
