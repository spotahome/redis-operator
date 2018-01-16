package failover

import (
	"errors"

	"github.com/spotahome/redis-operator/pkg/log"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/api/v1"
)

func (r *RedisFailoverKubeClient) waitForPod(name string, namespace string, logger log.Logger) error {
	t := r.clock.NewTicker(loopInterval)
	to := r.clock.After(waitTimeout)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			logger.Debug("Waiting for pod to be ready")
			pod, _ := r.Client.CoreV1().Pods(namespace).Get(name, metav1.GetOptions{})
			for _, condition := range pod.Status.Conditions {
				if condition.Type == "Ready" && condition.Status == v1.ConditionTrue {
					return nil
				}
			}
		case <-to:
			return errors.New("timeout waiting the condition")
		}
	}
}

func (r *RedisFailoverKubeClient) waitForService(name string, namespace string, logger log.Logger) error {
	t := r.clock.NewTicker(loopInterval)
	to := r.clock.After(waitTimeout)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			logger.Debug("Waiting for service to find bootstrap pod")
			endpoints, _ := r.Client.CoreV1().Endpoints(namespace).Get(name, metav1.GetOptions{})
			addresses := 0
			for _, subset := range endpoints.Subsets {
				addresses += len(subset.Addresses)
			}
			if addresses > 0 {
				return nil
			}
		case <-to:
			return errors.New("timeout waiting the condition")
		}
	}
}

func (r *RedisFailoverKubeClient) waitForDeployment(name string, namespace string, replicas int32, logger log.Logger) error {
	t := r.clock.NewTicker(loopInterval)
	to := r.clock.After(waitTimeout)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			logger.Debug("Waiting for Sentinel deployment to be fully operative")
			deployment, _ := r.Client.AppsV1beta1().Deployments(namespace).Get(name, metav1.GetOptions{})
			if deployment.Status.ReadyReplicas == replicas && deployment.Status.Replicas == replicas {
				return nil
			}
		case <-to:
			return errors.New("timeout waiting the condition")
		}
	}
}

func (r *RedisFailoverKubeClient) waitForStatefulset(name string, namespace string, replicas int32, logger log.Logger) error {
	t := r.clock.NewTicker(loopInterval)
	to := r.clock.After(waitTimeout)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			logger.Debug("Waiting for Redis statefulset to be fully operative")
			statefulset, _ := r.Client.AppsV1beta1().StatefulSets(namespace).Get(name, metav1.GetOptions{})
			if statefulset.Status.ReadyReplicas == replicas && statefulset.Status.Replicas == replicas {
				return nil
			}
		case <-to:
			return errors.New("timeout waiting the condition")
		}
	}
}

func (r *RedisFailoverKubeClient) waitForPodDeletion(name string, namespace string, logger log.Logger) error {
	t := r.clock.NewTicker(loopInterval)
	to := r.clock.After(waitTimeout)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			logger.Debug("Waiting for pod to terminate")
			podList, _ := r.Client.CoreV1().Pods(namespace).List(metav1.ListOptions{})
			found := false
			for _, pod := range podList.Items {
				if pod.Name == name {
					found = true
				}
			}
			if !found {
				return nil
			}
		case <-to:
			return errors.New("timeout waiting the condition")
		}
	}
}

func (r *RedisFailoverKubeClient) waitForStatefulsetDeletion(name string, namespace string, logger log.Logger) error {
	t := r.clock.NewTicker(loopInterval)
	to := r.clock.After(waitTimeout)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			logger.Debug("Waiting for statefulset to terminate")
			statefulsetList, _ := r.Client.AppsV1beta1().StatefulSets(namespace).List(metav1.ListOptions{})
			found := false
			for _, statefulset := range statefulsetList.Items {
				if statefulset.Name == name {
					found = true
				}
			}
			if !found {
				return nil
			}
		case <-to:
			return errors.New("timeout waiting the condition")
		}
	}
}

func (r *RedisFailoverKubeClient) waitForDeploymentDeletion(name string, namespace string, logger log.Logger) error {
	t := r.clock.NewTicker(loopInterval)
	to := r.clock.After(waitTimeout)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			logger.Debug("Waiting for deployment to terminate")
			deploymentList, _ := r.Client.Apps().Deployments(namespace).List(metav1.ListOptions{})
			found := false
			for _, deployment := range deploymentList.Items {
				if deployment.Name == name {
					found = true
				}
			}
			if !found {
				return nil
			}
		case <-to:
			return errors.New("timeout waiting the condition")
		}
	}
}

func (r *RedisFailoverKubeClient) waitForServiceDeletion(name string, namespace string, logger log.Logger) error {
	t := r.clock.NewTicker(loopInterval)
	to := r.clock.After(waitTimeout)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			logger.Debug("Waiting for service to disappear")
			serviceList, _ := r.Client.Core().Services(namespace).List(metav1.ListOptions{})
			found := false
			for _, service := range serviceList.Items {
				if service.Name == name {
					found = true
				}
			}
			if !found {
				return nil
			}
		case <-to:
			return errors.New("timeout waiting the condition")
		}
	}
}
