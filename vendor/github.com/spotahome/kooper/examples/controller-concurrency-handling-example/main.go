package main

import (
	"context"
	"flag"
	"os"
	"path/filepath"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	"github.com/spotahome/kooper/log"
	"github.com/spotahome/kooper/operator/controller"
	"github.com/spotahome/kooper/operator/handler"
	"github.com/spotahome/kooper/operator/retrieve"
)

var (
	concurrentWorkers int
	sleepMS           int
)

func initFlags() error {
	fg := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fg.IntVar(&concurrentWorkers, "concurrency", 0, "The number of concurrent event handling")
	fg.IntVar(&sleepMS, "sleep-ms", 25, "The number of milliseconds to sleep on each event handling")
	return fg.Parse(os.Args[1:])
}

func sleep() {
	time.Sleep(time.Duration(sleepMS) * time.Millisecond)
}

func main() {
	// Initialize logger.
	log := &log.Std{}

	// Init flags.
	if err := initFlags(); err != nil {
		log.Errorf("error parsing arguments: %s", err)
		os.Exit(1)
	}

	// Get k8s client.
	k8scfg, err := rest.InClusterConfig()
	if err != nil {
		// No in cluster? letr's try locally
		kubehome := filepath.Join(homedir.HomeDir(), ".kube", "config")
		k8scfg, err = clientcmd.BuildConfigFromFlags("", kubehome)
		if err != nil {
			log.Errorf("error loading kubernetes configuration: %s", err)
			os.Exit(1)
		}
	}
	k8scli, err := kubernetes.NewForConfig(k8scfg)
	if err != nil {
		log.Errorf("error creating kubernetes client: %s", err)
		os.Exit(1)
	}

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
		AddFunc: func(_ context.Context, obj runtime.Object) error {
			pod := obj.(*corev1.Pod)
			sleep()
			log.Infof("Pod added: %s/%s", pod.Namespace, pod.Name)
			return nil
		},
		DeleteFunc: func(_ context.Context, s string) error {
			sleep()
			log.Infof("Pod deleted: %s", s)
			return nil
		},
	}

	// Create the controller that will refresh every 30 seconds.
	var ctrl controller.Controller
	if concurrentWorkers < 2 {
		log.Infof("sequential controller created")
		ctrl = controller.NewSequential(30*time.Second, hand, retr, nil, log)
	} else {
		log.Infof("sequential controller created")
		ctrl, _ = controller.NewConcurrent(concurrentWorkers, 30*time.Second, hand, retr, nil, log)
	}

	// Start our controller.
	stopC := make(chan struct{})
	if err := ctrl.Run(stopC); err != nil {
		log.Errorf("error running controller: %s", err)
		os.Exit(1)
	}
	os.Exit(0)
}
