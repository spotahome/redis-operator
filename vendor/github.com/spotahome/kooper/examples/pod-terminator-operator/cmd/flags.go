package main

import (
	"flag"
	"os"
	"path/filepath"
	"time"

	"k8s.io/client-go/util/homedir"

	"github.com/spotahome/kooper/examples/pod-terminator-operator/operator"
)

// Flags are the controller flags.
type Flags struct {
	flagSet *flag.FlagSet

	ResyncSec   int
	KubeConfig  string
	Development bool
}

// OperatorConfig converts the command line flag arguments to operator configuration.
func (f *Flags) OperatorConfig() operator.Config {
	return operator.Config{
		ResyncPeriod: time.Duration(f.ResyncSec) * time.Second,
	}
}

// NewFlags returns a new Flags.
func NewFlags() *Flags {
	f := &Flags{
		flagSet: flag.NewFlagSet(os.Args[0], flag.ExitOnError),
	}
	// Get the user kubernetes configuration in it's home directory.
	kubehome := filepath.Join(homedir.HomeDir(), ".kube", "config")

	// Init flags.
	f.flagSet.IntVar(&f.ResyncSec, "resync-seconds", 30, "The number of seconds the controller will resync the resources")
	f.flagSet.StringVar(&f.KubeConfig, "kubeconfig", kubehome, "kubernetes configuration path, only used when development mode enabled")
	f.flagSet.BoolVar(&f.Development, "development", false, "development flag will allow to run the operator outside a kubernetes cluster")

	f.flagSet.Parse(os.Args[1:])

	return f
}
