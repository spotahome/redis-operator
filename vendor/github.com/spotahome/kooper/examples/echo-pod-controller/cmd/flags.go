package main

import (
	"flag"
	"os"
	"path/filepath"
	"time"

	"github.com/spotahome/kooper/examples/echo-pod-controller/controller"
	"k8s.io/client-go/util/homedir"
)

// Flags are the controller flags.
type Flags struct {
	flagSet *flag.FlagSet

	Namespace   string
	ResyncSec   int
	KubeConfig  string
	Development bool
}

// ControllerConfig converts the command line flag arguments to controller configuration.
func (f *Flags) ControllerConfig() controller.Config {
	return controller.Config{
		Namespace:    f.Namespace,
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
	f.flagSet.StringVar(&f.Namespace, "namespace", "", "kubernetes namespace where this app is running")
	f.flagSet.IntVar(&f.ResyncSec, "resync-seconds", 30, "The number of seconds the controller will resync the resources")
	f.flagSet.StringVar(&f.KubeConfig, "kubeconfig", kubehome, "kubernetes configuration path, only used when development mode enabled")
	f.flagSet.BoolVar(&f.Development, "development", false, "development flag will allow to run the operator outside a kubernetes cluster")

	f.flagSet.Parse(os.Args[1:])

	return f
}
