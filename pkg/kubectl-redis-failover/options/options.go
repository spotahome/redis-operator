package options

import (
	"errors"
	"flag"
	"fmt"
	"strconv"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"

	"github.com/spf13/cobra"
	clientset "github.com/spotahome/redis-operator/client/k8s/clientset/versioned"
	"github.com/spotahome/redis-operator/log"
)

const (
	cliName   = "kubectl redis failover"
	infoLevel = "info"
)

type RedisFailoverOptions struct {
	CLIName             string
	RESTClientGetter    genericclioptions.RESTClientGetter
	ConfigFlags         *genericclioptions.ConfigFlags
	KlogLevel           int
	LogLevel            string
	RedisFailoverClient clientset.Interface
	KubeClient          kubernetes.Interface
	DynamicClient       dynamic.Interface

	Log log.Logger
	genericclioptions.IOStreams

	Now func() metav1.Time
}

// NewRedisFailoverOptions provides an instance of RedisFailoverOptions with default values
func NewRedisFailoverOptions(streams genericclioptions.IOStreams) *RedisFailoverOptions {
	logger := log.Base()
	logger.SetOutput(streams.ErrOut)
	klog.SetOutput(streams.ErrOut)
	configFlags := genericclioptions.NewConfigFlags(true)

	return &RedisFailoverOptions{
		CLIName:          cliName,
		RESTClientGetter: configFlags,
		ConfigFlags:      configFlags,
		IOStreams:        streams,
		Log:              logger,
		LogLevel:         infoLevel,
		Now:              metav1.Now,
	}
}

// Example returns the example string with the CLI command replaced in the example
func (o *RedisFailoverOptions) Example(example string) string {
	return strings.Trim(fmt.Sprintf(example, cliName), "\n")
}

func (o *RedisFailoverOptions) UsageErr(c *cobra.Command) error {
	c.Usage()
	c.SilenceErrors = true
	return errors.New(c.UsageString())
}

// PersistentPreRunE contains common logic which will be executed for all commands
func (o *RedisFailoverOptions) PersistentPreRunE(c *cobra.Command, args []string) error {
	// NOTE: we set the output of the cobra command to stderr because the only thing that should
	// emit to this are returned errors from command.RunE
	c.SetOut(o.ErrOut)
	c.SetErr(o.ErrOut)
	o.Log.Set(log.Level(o.LogLevel))
	if flag.Lookup("v") != nil {
		// the '-v' flag is set by klog.Init(), which we only call in main.go
		err := flag.Set("v", strconv.Itoa(o.KlogLevel))
		if err != nil {
			return err
		}
	}
	return nil
}

// AddKubectlFlags adds kubectl related flags to the command
func (o *RedisFailoverOptions) AddKubectlFlags(cmd *cobra.Command) {
	flags := cmd.PersistentFlags()
	o.ConfigFlags.AddFlags(flags)
	flags.IntVarP(&o.KlogLevel, "kloglevel", "v", 0, "Log level for kubernetes client library")
	flags.StringVar(&o.LogLevel, "loglevel", infoLevel, "Log level for kubectl redis failover")
}

// RedisFailoversClientset returns a Rollout client interface based on client flags
func (o *RedisFailoverOptions) RedisFailoversClientset() clientset.Interface {
	if o.RedisFailoverClient != nil {
		return o.RedisFailoverClient
	}
	config, err := o.RESTClientGetter.ToRESTConfig()
	if err != nil {
		panic(err)
	}
	redisFailoverClient, err := clientset.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	o.RedisFailoverClient = redisFailoverClient
	return o.RedisFailoverClient
}

// KubeClientset returns a Kubernetes client interface based on client flags
func (o *RedisFailoverOptions) KubeClientset() kubernetes.Interface {
	if o.KubeClient != nil {
		return o.KubeClient
	}
	config, err := o.RESTClientGetter.ToRESTConfig()
	if err != nil {
		panic(err)
	}
	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	o.KubeClient = kubeClient
	return o.KubeClient
}

// DynamicClientset returns a Dynamic client interface based on client flags
func (o *RedisFailoverOptions) DynamicClientset() dynamic.Interface {
	if o.DynamicClient != nil {
		return o.DynamicClient
	}
	config, err := o.RESTClientGetter.ToRESTConfig()
	if err != nil {
		panic(err)
	}
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	o.DynamicClient = dynamicClient
	return o.DynamicClient
}

// Namespace returns the namespace based on client flags or kube context
func (o *RedisFailoverOptions) Namespace() string {
	namespace, _, err := o.RESTClientGetter.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		panic(err)
	}
	return namespace
}
