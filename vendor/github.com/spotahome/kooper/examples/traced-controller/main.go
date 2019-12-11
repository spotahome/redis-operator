package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	jaeger "github.com/uber/jaeger-client-go"
	jaegerconfig "github.com/uber/jaeger-client-go/config"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	"github.com/spotahome/kooper/log"
	"github.com/spotahome/kooper/operator/controller"
	"github.com/spotahome/kooper/operator/handler"
	"github.com/spotahome/kooper/operator/retrieve"
)

// Important. Run a jaeger development instance to see the traces.
//
// docker run --rm -it -p5775:5775/udp -p6831:6831/udp -p6832:6832/udp -p5778:5778 -p16686:16686 -p14268:14268 -p9411:9411 jaegertracing/all-in-one:latest

const (
	namespaceDef             = "default"
	controllerNameDef        = "traced-pod-controller"
	resyncIntervalSecondsDef = 120
)

// Flags are the flags of the program.
type Flags struct {
	ControllerName        string
	ResyncIntervalSeconds int
	Namespace             string
}

// NewFlags returns the flags of the commandline.
func NewFlags() *Flags {
	flags := &Flags{}
	fl := flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	fl.IntVar(&flags.ResyncIntervalSeconds, "resync-interval", resyncIntervalSecondsDef, "resync seconds of the controller")
	fl.StringVar(&flags.Namespace, "namespace", namespaceDef, "kubernetes namespace where the controller is running")
	fl.StringVar(&flags.ControllerName, "controller-name", controllerNameDef, "controller name (service name)")

	fl.Parse(os.Args[1:])

	return flags
}

// Main runs the main application.
func Main() error {
	// Flags
	fl := NewFlags()

	// Initialize logger.
	logger := &log.Std{}

	// Get k8s client.
	k8scfg, err := rest.InClusterConfig()
	if err != nil {
		// No in cluster? letr's try locally
		kubehome := filepath.Join(homedir.HomeDir(), ".kube", "config")
		k8scfg, err = clientcmd.BuildConfigFromFlags("", kubehome)
		if err != nil {
			return fmt.Errorf("error loading kubernetes configuration: %s", err)
		}
	}
	k8scli, err := kubernetes.NewForConfig(k8scfg)
	if err != nil {
		return fmt.Errorf("error creating kubernetes client: %s", err)
	}

	// Create AWS service.
	aws, closer, err := createFakeService("aws")
	if err != nil {
		return err
	}
	defer closer.Close()

	// Create Redis service.
	redis, closer, err := createFakeService("redis")
	if err != nil {
		return err
	}
	defer closer.Close()

	// Create Redis service.
	github, closer, err := createFakeService("github")
	if err != nil {
		return err
	}
	defer closer.Close()

	// Create controller tracer.
	tracer, closer, err := createTracer(fl.ControllerName)
	if err != nil {
		return err
	}
	defer closer.Close()

	// Create our retriever so the controller knows how to get/listen for pod events.
	retr := &retrieve.Resource{
		Object: &corev1.Pod{},
		ListerWatcher: &cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				return k8scli.CoreV1().Pods("").List(options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				return k8scli.CoreV1().Pods("").Watch(options)
			},
		},
	}

	// Our domain logic that will print every add/sync/update and delete event we .
	hand := &handler.HandlerFunc{
		AddFunc: func(ctx context.Context, obj runtime.Object) error {
			pod := obj.(*corev1.Pod)
			logger.Infof("Pod added: %s/%s", pod.Namespace, pod.Name)

			// Create span
			pSpan := opentracing.SpanFromContext(ctx)
			span := tracer.StartSpan("AddFunc", opentracing.ChildOf(pSpan.Context()))
			defer span.Finish()
			ctx = opentracing.ContextWithSpan(ctx, span)

			// Execute.
			if _, err := redis.makeOperation(ctx, "getOrg", 10*time.Millisecond, true); err != nil {
				return err
			}
			if _, err := github.makeOperation(ctx, "getOrg", 75*time.Millisecond, true); err != nil {
				return err
			}

			if _, err := redis.makeOperation(ctx, "getEc2List", 14*time.Millisecond, true); err != nil {
				return err
			}

			// AWS stuff
			{
				var newCtx context.Context
				newCtx, err := aws.makeOperation(ctx, "ensureAWSResources", 120*time.Millisecond, true)
				if err != nil {
					return err
				}

				if _, err := aws.makeOperation(newCtx, "getEc2List", 672*time.Millisecond, true); err != nil {
					return err
				}

				if _, err := aws.makeOperation(newCtx, "getALBs", 232*time.Millisecond, true); err != nil {
					return err
				}
				if _, err := aws.makeOperation(newCtx, "linkSGInALB", 157*time.Millisecond, true); err != nil {
					return err
				}
			}

			if _, err := redis.makeOperation(ctx, "storeResult", 34*time.Millisecond, true); err != nil {
				return err
			}

			return nil
		},
		DeleteFunc: func(ctx context.Context, s string) error {
			logger.Infof("Pod deleted: %s", s)

			// Create span
			pSpan := opentracing.SpanFromContext(ctx)
			span := tracer.StartSpan("DeleteFunc", opentracing.ChildOf(pSpan.Context()))
			defer span.Finish()
			ctx = opentracing.ContextWithSpan(ctx, span)

			// Execute.
			if _, err := aws.makeOperation(ctx, "deleteCloudformationStack", 698*time.Millisecond, true); err != nil {
				return err
			}
			if _, err := redis.makeOperation(ctx, "storeResult", 26*time.Millisecond, true); err != nil {
				return err
			}

			return nil
		},
	}

	// Create the controller and run.
	cfg := &controller.Config{
		Name:                 fl.ControllerName,
		ProcessingJobRetries: 5,
		ResyncInterval:       time.Duration(fl.ResyncIntervalSeconds) * time.Second,
		ConcurrentWorkers:    30,
	}
	ctrl := controller.New(cfg, hand, retr, nil, tracer, nil, logger)
	stopC := make(chan struct{})
	errC := make(chan error)
	go func() {
		errC <- ctrl.Run(stopC)
	}()

	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, syscall.SIGTERM, syscall.SIGINT)

	select {
	case err := <-errC:
		if err != nil {
			logger.Infof("controller finished with error: %s", err)
			return err
		}
		logger.Infof("controller finished successfuly")
	case s := <-sigC:
		logger.Infof("signal %s received", s)
		close(stopC)
	}

	time.Sleep(5 * time.Second)

	return nil
}

func createFakeService(service string) (*fakeService, io.Closer, error) {
	t, cl, err := createTracer(service)
	return &fakeService{
		tracer: t,
	}, cl, err
}

func createTracer(service string) (opentracing.Tracer, io.Closer, error) {
	cfg := &jaegerconfig.Configuration{
		ServiceName: service,
		Sampler: &jaegerconfig.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &jaegerconfig.ReporterConfig{
			LogSpans: true,
		},
	}
	tracer, closer, err := cfg.NewTracer(jaegerconfig.Logger(jaeger.NullLogger))
	if err != nil {
		return nil, nil, fmt.Errorf("cannot init Jaeger: %s", err)
	}
	return tracer, closer, nil
}

func main() {
	if err := Main(); err != nil {
		fmt.Fprintf(os.Stderr, "error executing controller: %s", err)
		os.Exit(1)
	}
	os.Exit(0)
}
