package utils

import (
	"flag"
	"path/filepath"

	"github.com/spotahome/redis-operator/operator/redisfailover"
	"k8s.io/client-go/util/homedir"
)

// CMDFlags are the flags used by the cmd
// TODO: improve flags.
type CMDFlags struct {
	KubeConfig  string
	Development bool
	Debug       bool
	ListenAddr  string
	MetricsPath string
	Concurrency int
}

// Init initializes and parse the flags
func (c *CMDFlags) Init() {
	kubehome := filepath.Join(homedir.HomeDir(), ".kube", "config")
	// register flags
	flag.StringVar(&c.KubeConfig, "kubeconfig", kubehome, "kubernetes configuration path, only used when development mode enabled")
	flag.BoolVar(&c.Development, "development", false, "development flag will allow to run the operator outside a kubernetes cluster")
	flag.BoolVar(&c.Debug, "debug", false, "enable debug mode")
	flag.StringVar(&c.ListenAddr, "listen-address", ":9710", "Address to listen on for metrics.")
	flag.StringVar(&c.MetricsPath, "metrics-path", "/metrics", "Path to serve the metrics.")
	// default is 3 for conccurency because kooper also defines 3 as default
	// reference: https://github.com/spotahome/kooper/blob/master/controller/controller.go#L89
	flag.IntVar(&c.Concurrency, "concurrency", 3, "Number of conccurent workers meant to process events")
	// Parse flags
	flag.Parse()
}

// ToRedisOperatorConfig convert the flags to redisfailover config
func (c *CMDFlags) ToRedisOperatorConfig() redisfailover.Config {
	return redisfailover.Config{
		ListenAddress: c.ListenAddr,
		MetricsPath:   c.MetricsPath,
		Concurrency:   c.Concurrency,
	}
}
