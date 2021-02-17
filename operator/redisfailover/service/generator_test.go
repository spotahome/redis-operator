package service_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	redisfailoverv1 "github.com/spotahome/redis-operator/api/redisfailover/v1"
	"github.com/spotahome/redis-operator/log"
	mK8SService "github.com/spotahome/redis-operator/mocks/service/k8s"
	rfservice "github.com/spotahome/redis-operator/operator/redisfailover/service"
)

func TestRedisStatefulSetStorageGeneration(t *testing.T) {
	configMapName := rfservice.GetRedisName(generateRF())
	shutdownConfigMapName := rfservice.GetRedisShutdownConfigMapName(generateRF())
	readinesConfigMapName := rfservice.GetRedisReadinessName(generateRF())
	executeMode := int32(0744)
	tests := []struct {
		name           string
		ownerRefs      []metav1.OwnerReference
		expectedSS     appsv1.StatefulSet
		rfRedisStorage redisfailoverv1.RedisStorage
	}{
		{
			name: "Default values",
			expectedSS: appsv1.StatefulSet{
				Spec: appsv1.StatefulSetSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									VolumeMounts: []corev1.VolumeMount{
										{
											Name:      "redis-config",
											MountPath: "/redis",
										},
										{
											Name:      "redis-shutdown-config",
											MountPath: "/redis-shutdown",
										},
										{
											Name:      "redis-readiness-config",
											MountPath: "/redis-readiness",
										},
										{
											Name:      "redis-data",
											MountPath: "/data",
										},
									},
								},
							},
							Volumes: []corev1.Volume{
								{
									Name: "redis-config",
									VolumeSource: corev1.VolumeSource{
										ConfigMap: &corev1.ConfigMapVolumeSource{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: configMapName,
											},
										},
									},
								},
								{
									Name: "redis-shutdown-config",
									VolumeSource: corev1.VolumeSource{
										ConfigMap: &corev1.ConfigMapVolumeSource{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: shutdownConfigMapName,
											},
											DefaultMode: &executeMode,
										},
									},
								},
								{
									Name: "redis-readiness-config",
									VolumeSource: corev1.VolumeSource{
										ConfigMap: &corev1.ConfigMapVolumeSource{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: readinesConfigMapName,
											},
											DefaultMode: &executeMode,
										},
									},
								},
								{
									Name: "redis-data",
									VolumeSource: corev1.VolumeSource{
										EmptyDir: &corev1.EmptyDirVolumeSource{},
									},
								},
							},
						},
					},
				},
			},
			rfRedisStorage: redisfailoverv1.RedisStorage{},
		},
		{
			name: "Defined an emptydir with storage on memory",
			expectedSS: appsv1.StatefulSet{
				Spec: appsv1.StatefulSetSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									VolumeMounts: []corev1.VolumeMount{
										{
											Name:      "redis-config",
											MountPath: "/redis",
										},
										{
											Name:      "redis-shutdown-config",
											MountPath: "/redis-shutdown",
										},
										{
											Name:      "redis-readiness-config",
											MountPath: "/redis-readiness",
										},
										{
											Name:      "redis-data",
											MountPath: "/data",
										},
									},
								},
							},
							Volumes: []corev1.Volume{
								{
									Name: "redis-config",
									VolumeSource: corev1.VolumeSource{
										ConfigMap: &corev1.ConfigMapVolumeSource{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: configMapName,
											},
										},
									},
								},
								{
									Name: "redis-shutdown-config",
									VolumeSource: corev1.VolumeSource{
										ConfigMap: &corev1.ConfigMapVolumeSource{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: shutdownConfigMapName,
											},
											DefaultMode: &executeMode,
										},
									},
								},
								{
									Name: "redis-readiness-config",
									VolumeSource: corev1.VolumeSource{
										ConfigMap: &corev1.ConfigMapVolumeSource{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: readinesConfigMapName,
											},
											DefaultMode: &executeMode,
										},
									},
								},
								{
									Name: "redis-data",
									VolumeSource: corev1.VolumeSource{
										EmptyDir: &corev1.EmptyDirVolumeSource{
											Medium: corev1.StorageMediumMemory,
										},
									},
								},
							},
						},
					},
				},
			},
			rfRedisStorage: redisfailoverv1.RedisStorage{
				EmptyDir: &corev1.EmptyDirVolumeSource{
					Medium: corev1.StorageMediumMemory,
				},
			},
		},
		{
			name: "Defined an persistentvolumeclaim",
			expectedSS: appsv1.StatefulSet{
				Spec: appsv1.StatefulSetSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									VolumeMounts: []corev1.VolumeMount{
										{
											Name:      "redis-config",
											MountPath: "/redis",
										},
										{
											Name:      "redis-shutdown-config",
											MountPath: "/redis-shutdown",
										},
										{
											Name:      "redis-readiness-config",
											MountPath: "/redis-readiness",
										},
										{
											Name:      "pvc-data",
											MountPath: "/data",
										},
									},
								},
							},
							Volumes: []corev1.Volume{
								{
									Name: "redis-config",
									VolumeSource: corev1.VolumeSource{
										ConfigMap: &corev1.ConfigMapVolumeSource{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: configMapName,
											},
										},
									},
								},
								{
									Name: "redis-shutdown-config",
									VolumeSource: corev1.VolumeSource{
										ConfigMap: &corev1.ConfigMapVolumeSource{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: shutdownConfigMapName,
											},
											DefaultMode: &executeMode,
										},
									},
								},
								{
									Name: "redis-readiness-config",
									VolumeSource: corev1.VolumeSource{
										ConfigMap: &corev1.ConfigMapVolumeSource{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: readinesConfigMapName,
											},
											DefaultMode: &executeMode,
										},
									},
								},
							},
						},
					},
					VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "pvc-data",
							},
							Spec: corev1.PersistentVolumeClaimSpec{
								AccessModes: []corev1.PersistentVolumeAccessMode{
									"ReadWriteOnce",
								},
								Resources: corev1.ResourceRequirements{
									Requests: corev1.ResourceList{
										corev1.ResourceStorage: resource.MustParse("1Gi"),
									},
								},
							},
						},
					},
				},
			},
			rfRedisStorage: redisfailoverv1.RedisStorage{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pvc-data",
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						AccessModes: []corev1.PersistentVolumeAccessMode{
							"ReadWriteOnce",
						},
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceStorage: resource.MustParse("1Gi"),
							},
						},
					},
				},
			},
		},
		{
			name: "Defined an persistentvolumeclaim with ownerRefs",
			ownerRefs: []metav1.OwnerReference{
				{
					Name: "testing",
				},
			},
			expectedSS: appsv1.StatefulSet{
				Spec: appsv1.StatefulSetSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									VolumeMounts: []corev1.VolumeMount{
										{
											Name:      "redis-config",
											MountPath: "/redis",
										},
										{
											Name:      "redis-shutdown-config",
											MountPath: "/redis-shutdown",
										},
										{
											Name:      "redis-readiness-config",
											MountPath: "/redis-readiness",
										},
										{
											Name:      "pvc-data",
											MountPath: "/data",
										},
									},
								},
							},
							Volumes: []corev1.Volume{
								{
									Name: "redis-config",
									VolumeSource: corev1.VolumeSource{
										ConfigMap: &corev1.ConfigMapVolumeSource{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: configMapName,
											},
										},
									},
								},
								{
									Name: "redis-shutdown-config",
									VolumeSource: corev1.VolumeSource{
										ConfigMap: &corev1.ConfigMapVolumeSource{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: shutdownConfigMapName,
											},
											DefaultMode: &executeMode,
										},
									},
								},
								{
									Name: "redis-readiness-config",
									VolumeSource: corev1.VolumeSource{
										ConfigMap: &corev1.ConfigMapVolumeSource{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: readinesConfigMapName,
											},
											DefaultMode: &executeMode,
										},
									},
								},
							},
						},
					},
					VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "pvc-data",
								OwnerReferences: []metav1.OwnerReference{
									{
										Name: "testing",
									},
								},
							},
							Spec: corev1.PersistentVolumeClaimSpec{
								AccessModes: []corev1.PersistentVolumeAccessMode{
									"ReadWriteOnce",
								},
								Resources: corev1.ResourceRequirements{
									Requests: corev1.ResourceList{
										corev1.ResourceStorage: resource.MustParse("1Gi"),
									},
								},
							},
						},
					},
				},
			},
			rfRedisStorage: redisfailoverv1.RedisStorage{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pvc-data",
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						AccessModes: []corev1.PersistentVolumeAccessMode{
							"ReadWriteOnce",
						},
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceStorage: resource.MustParse("1Gi"),
							},
						},
					},
				},
			},
		},
		{
			name: "Defined an persistentvolumeclaim with ownerRefs keeping the pvc",
			ownerRefs: []metav1.OwnerReference{
				{
					Name: "testing",
				},
			},
			expectedSS: appsv1.StatefulSet{
				Spec: appsv1.StatefulSetSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									VolumeMounts: []corev1.VolumeMount{
										{
											Name:      "redis-config",
											MountPath: "/redis",
										},
										{
											Name:      "redis-shutdown-config",
											MountPath: "/redis-shutdown",
										},
										{
											Name:      "redis-readiness-config",
											MountPath: "/redis-readiness",
										},
										{
											Name:      "pvc-data",
											MountPath: "/data",
										},
									},
								},
							},
							Volumes: []corev1.Volume{
								{
									Name: "redis-config",
									VolumeSource: corev1.VolumeSource{
										ConfigMap: &corev1.ConfigMapVolumeSource{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: configMapName,
											},
										},
									},
								},
								{
									Name: "redis-shutdown-config",
									VolumeSource: corev1.VolumeSource{
										ConfigMap: &corev1.ConfigMapVolumeSource{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: shutdownConfigMapName,
											},
											DefaultMode: &executeMode,
										},
									},
								},
								{
									Name: "redis-readiness-config",
									VolumeSource: corev1.VolumeSource{
										ConfigMap: &corev1.ConfigMapVolumeSource{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: readinesConfigMapName,
											},
											DefaultMode: &executeMode,
										},
									},
								},
							},
						},
					},
					VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "pvc-data",
							},
							Spec: corev1.PersistentVolumeClaimSpec{
								AccessModes: []corev1.PersistentVolumeAccessMode{
									"ReadWriteOnce",
								},
								Resources: corev1.ResourceRequirements{
									Requests: corev1.ResourceList{
										corev1.ResourceStorage: resource.MustParse("1Gi"),
									},
								},
							},
						},
					},
				},
			},
			rfRedisStorage: redisfailoverv1.RedisStorage{
				KeepAfterDeletion: true,
				PersistentVolumeClaim: &corev1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pvc-data",
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						AccessModes: []corev1.PersistentVolumeAccessMode{
							"ReadWriteOnce",
						},
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceStorage: resource.MustParse("1Gi"),
							},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		assert := assert.New(t)

		// Generate a default RedisFailover and attaching the required storage
		rf := generateRF()
		rf.Spec.Redis.Storage = test.rfRedisStorage

		generatedStatefulSet := appsv1.StatefulSet{}

		ms := &mK8SService.Services{}
		ms.On("CreateOrUpdatePodDisruptionBudget", namespace, mock.Anything).Once().Return(nil, nil)
		ms.On("CreateOrUpdateStatefulSet", namespace, mock.Anything).Once().Run(func(args mock.Arguments) {
			ss := args.Get(1).(*appsv1.StatefulSet)
			generatedStatefulSet = *ss
		}).Return(nil)

		client := rfservice.NewRedisFailoverKubeClient(ms, log.Dummy)
		err := client.EnsureRedisStatefulset(rf, nil, test.ownerRefs)

		// Check that the storage-related fields are as spected
		assert.Equal(test.expectedSS.Spec.Template.Spec.Volumes, generatedStatefulSet.Spec.Template.Spec.Volumes)
		assert.Equal(test.expectedSS.Spec.Template.Spec.Containers[0].VolumeMounts, generatedStatefulSet.Spec.Template.Spec.Containers[0].VolumeMounts)
		assert.Equal(test.expectedSS.Spec.VolumeClaimTemplates, generatedStatefulSet.Spec.VolumeClaimTemplates)
		assert.NoError(err)
	}
}

func TestRedisStatefulSetCommands(t *testing.T) {
	tests := []struct {
		name             string
		givenCommands    []string
		expectedCommands []string
	}{
		{
			name:          "Default values",
			givenCommands: []string{},
			expectedCommands: []string{
				"redis-server",
				"/redis/redis.conf",
			},
		},
		{
			name: "Given commands should be used in redis container",
			givenCommands: []string{
				"test",
				"command",
			},
			expectedCommands: []string{
				"test",
				"command",
			},
		},
	}

	for _, test := range tests {
		assert := assert.New(t)

		// Generate a default RedisFailover and attaching the required storage
		rf := generateRF()
		rf.Spec.Redis.Command = test.givenCommands

		gotCommands := []string{}

		ms := &mK8SService.Services{}
		ms.On("CreateOrUpdatePodDisruptionBudget", namespace, mock.Anything).Once().Return(nil, nil)
		ms.On("CreateOrUpdateStatefulSet", namespace, mock.Anything).Once().Run(func(args mock.Arguments) {
			ss := args.Get(1).(*appsv1.StatefulSet)
			gotCommands = ss.Spec.Template.Spec.Containers[0].Command
		}).Return(nil)

		client := rfservice.NewRedisFailoverKubeClient(ms, log.Dummy)
		err := client.EnsureRedisStatefulset(rf, nil, []metav1.OwnerReference{})

		assert.Equal(test.expectedCommands, gotCommands)
		assert.NoError(err)
	}
}

func TestSentinelDeploymentCommands(t *testing.T) {
	tests := []struct {
		name             string
		givenCommands    []string
		expectedCommands []string
	}{
		{
			name:          "Default values",
			givenCommands: []string{},
			expectedCommands: []string{
				"redis-server",
				"/redis/sentinel.conf",
				"--sentinel",
			},
		},
		{
			name: "Given commands should be used in sentinel container",
			givenCommands: []string{
				"test",
				"command",
			},
			expectedCommands: []string{
				"test",
				"command",
			},
		},
	}

	for _, test := range tests {
		assert := assert.New(t)

		// Generate a default RedisFailover and attaching the required storage
		rf := generateRF()
		rf.Spec.Sentinel.Command = test.givenCommands

		gotCommands := []string{}

		ms := &mK8SService.Services{}
		ms.On("CreateOrUpdatePodDisruptionBudget", namespace, mock.Anything).Once().Return(nil, nil)
		ms.On("CreateOrUpdateDeployment", namespace, mock.Anything).Once().Run(func(args mock.Arguments) {
			d := args.Get(1).(*appsv1.Deployment)
			gotCommands = d.Spec.Template.Spec.Containers[0].Command
		}).Return(nil)

		client := rfservice.NewRedisFailoverKubeClient(ms, log.Dummy)
		err := client.EnsureSentinelDeployment(rf, nil, []metav1.OwnerReference{})

		assert.Equal(test.expectedCommands, gotCommands)
		assert.NoError(err)
	}
}

func TestRedisStatefulSetPodAnnotations(t *testing.T) {
	tests := []struct {
		name                   string
		givenPodAnnotations    map[string]string
		expectedPodAnnotations map[string]string
	}{
		{
			name:                   "PodAnnotations was not defined",
			givenPodAnnotations:    nil,
			expectedPodAnnotations: nil,
		},
		{
			name: "PodAnnotations is defined",
			givenPodAnnotations: map[string]string{
				"some":               "annotation",
				"path/to/annotation": "here",
			},
			expectedPodAnnotations: map[string]string{
				"some":               "annotation",
				"path/to/annotation": "here",
			},
		},
	}

	for _, test := range tests {
		assert := assert.New(t)

		// Generate a default RedisFailover and attaching the required annotations
		rf := generateRF()
		rf.Spec.Redis.PodAnnotations = test.givenPodAnnotations

		gotPodAnnotations := map[string]string{}

		ms := &mK8SService.Services{}
		ms.On("CreateOrUpdatePodDisruptionBudget", namespace, mock.Anything).Once().Return(nil, nil)
		ms.On("CreateOrUpdateStatefulSet", namespace, mock.Anything).Once().Run(func(args mock.Arguments) {
			ss := args.Get(1).(*appsv1.StatefulSet)
			gotPodAnnotations = ss.Spec.Template.ObjectMeta.Annotations
		}).Return(nil)

		client := rfservice.NewRedisFailoverKubeClient(ms, log.Dummy)
		err := client.EnsureRedisStatefulset(rf, nil, []metav1.OwnerReference{})

		assert.Equal(test.expectedPodAnnotations, gotPodAnnotations)
		assert.NoError(err)
	}
}

func TestSentinelDeploymentPodAnnotations(t *testing.T) {
	tests := []struct {
		name                   string
		givenPodAnnotations    map[string]string
		expectedPodAnnotations map[string]string
	}{
		{
			name:                   "PodAnnotations was not defined",
			givenPodAnnotations:    nil,
			expectedPodAnnotations: nil,
		},
		{
			name: "PodAnnotations is defined",
			givenPodAnnotations: map[string]string{
				"some":               "annotation",
				"path/to/annotation": "here",
			},
			expectedPodAnnotations: map[string]string{
				"some":               "annotation",
				"path/to/annotation": "here",
			},
		},
	}

	for _, test := range tests {
		assert := assert.New(t)

		// Generate a default RedisFailover and attaching the required annotations
		rf := generateRF()
		rf.Spec.Sentinel.PodAnnotations = test.givenPodAnnotations

		gotPodAnnotations := map[string]string{}

		ms := &mK8SService.Services{}
		ms.On("CreateOrUpdatePodDisruptionBudget", namespace, mock.Anything).Once().Return(nil, nil)
		ms.On("CreateOrUpdateDeployment", namespace, mock.Anything).Once().Run(func(args mock.Arguments) {
			d := args.Get(1).(*appsv1.Deployment)
			gotPodAnnotations = d.Spec.Template.ObjectMeta.Annotations
		}).Return(nil)

		client := rfservice.NewRedisFailoverKubeClient(ms, log.Dummy)
		err := client.EnsureSentinelDeployment(rf, nil, []metav1.OwnerReference{})

		assert.Equal(test.expectedPodAnnotations, gotPodAnnotations)
		assert.NoError(err)
	}
}

func TestRedisStatefulSetServiceAccountName(t *testing.T) {
	tests := []struct {
		name                       string
		givenServiceAccountName    string
		expectedServiceAccountName string
	}{
		{
			name:                       "ServiceAccountName was not defined",
			givenServiceAccountName:    "",
			expectedServiceAccountName: "",
		},
		{
			name:                       "ServiceAccountName is defined",
			givenServiceAccountName:    "redis-sa",
			expectedServiceAccountName: "redis-sa",
		},
	}

	for _, test := range tests {
		assert := assert.New(t)

		// Generate a default RedisFailover and attaching the required Service Account
		rf := generateRF()
		rf.Spec.Redis.ServiceAccountName = test.givenServiceAccountName

		gotServiceAccountName := ""

		ms := &mK8SService.Services{}
		ms.On("CreateOrUpdatePodDisruptionBudget", namespace, mock.Anything).Once().Return(nil, nil)
		ms.On("CreateOrUpdateStatefulSet", namespace, mock.Anything).Once().Run(func(args mock.Arguments) {
			ss := args.Get(1).(*appsv1.StatefulSet)
			gotServiceAccountName = ss.Spec.Template.Spec.ServiceAccountName
		}).Return(nil)

		client := rfservice.NewRedisFailoverKubeClient(ms, log.Dummy)
		err := client.EnsureRedisStatefulset(rf, nil, []metav1.OwnerReference{})

		assert.Equal(test.expectedServiceAccountName, gotServiceAccountName)
		assert.NoError(err)
	}
}

func TestSentinelDeploymentServiceAccountName(t *testing.T) {
	tests := []struct {
		name                       string
		givenServiceAccountName    string
		expectedServiceAccountName string
	}{
		{
			name:                       "ServiceAccountName was not defined",
			givenServiceAccountName:    "",
			expectedServiceAccountName: "",
		},
		{
			name:                       "ServiceAccountName is defined",
			givenServiceAccountName:    "sentinel-sa",
			expectedServiceAccountName: "sentinel-sa",
		},
	}

	for _, test := range tests {
		assert := assert.New(t)

		// Generate a default RedisFailover and attaching the required Service Account
		rf := generateRF()
		rf.Spec.Sentinel.ServiceAccountName = test.givenServiceAccountName

		gotServiceAccountName := ""

		ms := &mK8SService.Services{}
		ms.On("CreateOrUpdatePodDisruptionBudget", namespace, mock.Anything).Once().Return(nil, nil)
		ms.On("CreateOrUpdateDeployment", namespace, mock.Anything).Once().Run(func(args mock.Arguments) {
			d := args.Get(1).(*appsv1.Deployment)
			gotServiceAccountName = d.Spec.Template.Spec.ServiceAccountName
		}).Return(nil)

		client := rfservice.NewRedisFailoverKubeClient(ms, log.Dummy)
		err := client.EnsureSentinelDeployment(rf, nil, []metav1.OwnerReference{})

		assert.Equal(test.expectedServiceAccountName, gotServiceAccountName)
		assert.NoError(err)
	}
}

func TestSentinelService(t *testing.T) {
	tests := []struct {
		name            string
		rfName          string
		rfNamespace     string
		rfLabels        map[string]string
		rfAnnotations   map[string]string
		expectedService corev1.Service
	}{
		{
			name: "with defaults",
			expectedService: corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      sentinelName,
					Namespace: namespace,
					Labels: map[string]string{
						"app.kubernetes.io/component": "sentinel",
						"app.kubernetes.io/name":      name,
						"app.kubernetes.io/part-of":   "redis-failover",
					},
					OwnerReferences: []metav1.OwnerReference{
						{
							Name: "testing",
						},
					},
				},
				Spec: corev1.ServiceSpec{
					Selector: map[string]string{
						"app.kubernetes.io/component": "sentinel",
						"app.kubernetes.io/name":      name,
						"app.kubernetes.io/part-of":   "redis-failover",
					},
					Ports: []corev1.ServicePort{
						{
							Name:       "sentinel",
							Port:       26379,
							TargetPort: intstr.FromInt(26379),
							Protocol:   "TCP",
						},
					},
				},
			},
		},
		{
			name:   "with Name provided",
			rfName: "custom-name",
			expectedService: corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "rfs-custom-name",
					Namespace: namespace,
					Labels: map[string]string{
						"app.kubernetes.io/component": "sentinel",
						"app.kubernetes.io/name":      "custom-name",
						"app.kubernetes.io/part-of":   "redis-failover",
					},
					OwnerReferences: []metav1.OwnerReference{
						{
							Name: "testing",
						},
					},
				},
				Spec: corev1.ServiceSpec{
					Selector: map[string]string{
						"app.kubernetes.io/component": "sentinel",
						"app.kubernetes.io/name":      "custom-name",
						"app.kubernetes.io/part-of":   "redis-failover",
					},
					Ports: []corev1.ServicePort{
						{
							Name:       "sentinel",
							Port:       26379,
							TargetPort: intstr.FromInt(26379),
							Protocol:   "TCP",
						},
					},
				},
			},
		},
		{
			name:        "with Namespace provided",
			rfNamespace: "custom-namespace",
			expectedService: corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      sentinelName,
					Namespace: "custom-namespace",
					Labels: map[string]string{
						"app.kubernetes.io/component": "sentinel",
						"app.kubernetes.io/name":      name,
						"app.kubernetes.io/part-of":   "redis-failover",
					},
					OwnerReferences: []metav1.OwnerReference{
						{
							Name: "testing",
						},
					},
				},
				Spec: corev1.ServiceSpec{
					Selector: map[string]string{
						"app.kubernetes.io/component": "sentinel",
						"app.kubernetes.io/name":      name,
						"app.kubernetes.io/part-of":   "redis-failover",
					},
					Ports: []corev1.ServicePort{
						{
							Name:       "sentinel",
							Port:       26379,
							TargetPort: intstr.FromInt(26379),
							Protocol:   "TCP",
						},
					},
				},
			},
		},
		{
			name:     "with Labels provided",
			rfLabels: map[string]string{"some": "label"},
			expectedService: corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      sentinelName,
					Namespace: namespace,
					Labels: map[string]string{
						"app.kubernetes.io/component": "sentinel",
						"app.kubernetes.io/name":      name,
						"app.kubernetes.io/part-of":   "redis-failover",
						"some":                        "label",
					},
					OwnerReferences: []metav1.OwnerReference{
						{
							Name: "testing",
						},
					},
				},
				Spec: corev1.ServiceSpec{
					Selector: map[string]string{
						"app.kubernetes.io/component": "sentinel",
						"app.kubernetes.io/name":      name,
						"app.kubernetes.io/part-of":   "redis-failover",
					},
					Ports: []corev1.ServicePort{
						{
							Name:       "sentinel",
							Port:       26379,
							TargetPort: intstr.FromInt(26379),
							Protocol:   "TCP",
						},
					},
				},
			},
		},
		{
			name:          "with Annotations provided",
			rfAnnotations: map[string]string{"some": "annotation"},
			expectedService: corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      sentinelName,
					Namespace: namespace,
					Labels: map[string]string{
						"app.kubernetes.io/component": "sentinel",
						"app.kubernetes.io/name":      name,
						"app.kubernetes.io/part-of":   "redis-failover",
					},
					Annotations: map[string]string{"some": "annotation"},
					OwnerReferences: []metav1.OwnerReference{
						{
							Name: "testing",
						},
					},
				},
				Spec: corev1.ServiceSpec{
					Selector: map[string]string{
						"app.kubernetes.io/component": "sentinel",
						"app.kubernetes.io/name":      name,
						"app.kubernetes.io/part-of":   "redis-failover",
					},
					Ports: []corev1.ServicePort{
						{
							Name:       "sentinel",
							Port:       26379,
							TargetPort: intstr.FromInt(26379),
							Protocol:   "TCP",
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)

			// Generate a default RedisFailover and attaching the required annotations
			rf := generateRF()
			if test.rfName != "" {
				rf.Name = test.rfName
			}
			if test.rfNamespace != "" {
				rf.Namespace = test.rfNamespace
			}
			rf.Spec.Sentinel.ServiceAnnotations = test.rfAnnotations

			generatedService := corev1.Service{}

			ms := &mK8SService.Services{}
			ms.On("CreateIfNotExistsService", rf.Namespace, mock.Anything).Once().Run(func(args mock.Arguments) {
				s := args.Get(1).(*corev1.Service)
				generatedService = *s
			}).Return(nil)

			client := rfservice.NewRedisFailoverKubeClient(ms, log.Dummy)
			err := client.EnsureSentinelService(rf, test.rfLabels, []metav1.OwnerReference{{Name: "testing"}})

			assert.Equal(test.expectedService, generatedService)
			assert.NoError(err)
		})
	}
}

func TestRedisService(t *testing.T) {
	tests := []struct {
		name            string
		rfName          string
		rfNamespace     string
		rfLabels        map[string]string
		rfAnnotations   map[string]string
		expectedService corev1.Service
	}{
		{
			name: "with defaults",
			expectedService: corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      redisName,
					Namespace: namespace,
					Labels: map[string]string{
						"app.kubernetes.io/component": "redis",
						"app.kubernetes.io/name":      name,
						"app.kubernetes.io/part-of":   "redis-failover",
					},
					Annotations: map[string]string{
						"prometheus.io/scrape": "true",
						"prometheus.io/path":   "/metrics",
						"prometheus.io/port":   "http",
					},
					OwnerReferences: []metav1.OwnerReference{
						{
							Name: "testing",
						},
					},
				},
				Spec: corev1.ServiceSpec{
					Type:      corev1.ServiceTypeClusterIP,
					ClusterIP: corev1.ClusterIPNone,
					Selector: map[string]string{
						"app.kubernetes.io/component": "redis",
						"app.kubernetes.io/name":      name,
						"app.kubernetes.io/part-of":   "redis-failover",
					},
					Ports: []corev1.ServicePort{
						{
							Name:     "http-metrics",
							Port:     9121,
							Protocol: corev1.ProtocolTCP,
						},
					},
				},
			},
		},
		{
			name:   "with Name provided",
			rfName: "custom-name",
			expectedService: corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "rfr-custom-name",
					Namespace: namespace,
					Labels: map[string]string{
						"app.kubernetes.io/component": "redis",
						"app.kubernetes.io/name":      "custom-name",
						"app.kubernetes.io/part-of":   "redis-failover",
					},
					Annotations: map[string]string{
						"prometheus.io/scrape": "true",
						"prometheus.io/path":   "/metrics",
						"prometheus.io/port":   "http",
					},
					OwnerReferences: []metav1.OwnerReference{
						{
							Name: "testing",
						},
					},
				},
				Spec: corev1.ServiceSpec{
					Type:      corev1.ServiceTypeClusterIP,
					ClusterIP: corev1.ClusterIPNone,
					Selector: map[string]string{
						"app.kubernetes.io/component": "redis",
						"app.kubernetes.io/name":      "custom-name",
						"app.kubernetes.io/part-of":   "redis-failover",
					},
					Ports: []corev1.ServicePort{
						{
							Name:     "http-metrics",
							Port:     9121,
							Protocol: corev1.ProtocolTCP,
						},
					},
				},
			},
		},
		{
			name:        "with Namespace provided",
			rfNamespace: "custom-namespace",
			expectedService: corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      redisName,
					Namespace: "custom-namespace",
					Labels: map[string]string{
						"app.kubernetes.io/component": "redis",
						"app.kubernetes.io/name":      name,
						"app.kubernetes.io/part-of":   "redis-failover",
					},
					Annotations: map[string]string{
						"prometheus.io/scrape": "true",
						"prometheus.io/path":   "/metrics",
						"prometheus.io/port":   "http",
					},
					OwnerReferences: []metav1.OwnerReference{
						{
							Name: "testing",
						},
					},
				},
				Spec: corev1.ServiceSpec{
					Type:      corev1.ServiceTypeClusterIP,
					ClusterIP: corev1.ClusterIPNone,
					Selector: map[string]string{
						"app.kubernetes.io/component": "redis",
						"app.kubernetes.io/name":      name,
						"app.kubernetes.io/part-of":   "redis-failover",
					},
					Ports: []corev1.ServicePort{
						{
							Name:     "http-metrics",
							Port:     9121,
							Protocol: corev1.ProtocolTCP,
						},
					},
				},
			},
		},
		{
			name:     "with Labels provided",
			rfLabels: map[string]string{"some": "label"},
			expectedService: corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      redisName,
					Namespace: namespace,
					Labels: map[string]string{
						"app.kubernetes.io/component": "redis",
						"app.kubernetes.io/name":      name,
						"app.kubernetes.io/part-of":   "redis-failover",
						"some":                        "label",
					},
					Annotations: map[string]string{
						"prometheus.io/scrape": "true",
						"prometheus.io/path":   "/metrics",
						"prometheus.io/port":   "http",
					},
					OwnerReferences: []metav1.OwnerReference{
						{
							Name: "testing",
						},
					},
				},
				Spec: corev1.ServiceSpec{
					Type:      corev1.ServiceTypeClusterIP,
					ClusterIP: corev1.ClusterIPNone,
					Selector: map[string]string{
						"app.kubernetes.io/component": "redis",
						"app.kubernetes.io/name":      name,
						"app.kubernetes.io/part-of":   "redis-failover",
					},
					Ports: []corev1.ServicePort{
						{
							Name:     "http-metrics",
							Port:     9121,
							Protocol: corev1.ProtocolTCP,
						},
					},
				},
			},
		},
		{
			name:          "with Annotations provided",
			rfAnnotations: map[string]string{"some": "annotation"},
			expectedService: corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      redisName,
					Namespace: namespace,
					Labels: map[string]string{
						"app.kubernetes.io/component": "redis",
						"app.kubernetes.io/name":      name,
						"app.kubernetes.io/part-of":   "redis-failover",
					},
					Annotations: map[string]string{
						"prometheus.io/scrape": "true",
						"prometheus.io/path":   "/metrics",
						"prometheus.io/port":   "http",
						"some":                 "annotation",
					},
					OwnerReferences: []metav1.OwnerReference{
						{
							Name: "testing",
						},
					},
				},
				Spec: corev1.ServiceSpec{
					Type:      corev1.ServiceTypeClusterIP,
					ClusterIP: corev1.ClusterIPNone,
					Selector: map[string]string{
						"app.kubernetes.io/component": "redis",
						"app.kubernetes.io/name":      name,
						"app.kubernetes.io/part-of":   "redis-failover",
					},
					Ports: []corev1.ServicePort{
						{
							Name:     "http-metrics",
							Port:     9121,
							Protocol: corev1.ProtocolTCP,
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)

			// Generate a default RedisFailover and attaching the required annotations
			rf := generateRF()
			if test.rfName != "" {
				rf.Name = test.rfName
			}
			if test.rfNamespace != "" {
				rf.Namespace = test.rfNamespace
			}
			rf.Spec.Redis.ServiceAnnotations = test.rfAnnotations

			generatedService := corev1.Service{}

			ms := &mK8SService.Services{}
			ms.On("CreateIfNotExistsService", rf.Namespace, mock.Anything).Once().Run(func(args mock.Arguments) {
				s := args.Get(1).(*corev1.Service)
				generatedService = *s
			}).Return(nil)

			client := rfservice.NewRedisFailoverKubeClient(ms, log.Dummy)
			err := client.EnsureRedisService(rf, test.rfLabels, []metav1.OwnerReference{{Name: "testing"}})

			assert.Equal(test.expectedService, generatedService)
			assert.NoError(err)
		})
	}
}

func TestRedisHostNetworkAndDnsPolicy(t *testing.T) {
	tests := []struct {
		name                string
		hostNetwork         bool
		expectedHostNetwork bool
		dnsPolicy           corev1.DNSPolicy
		expectedDnsPolicy   corev1.DNSPolicy
	}{
		{
			name:                "Default",
			expectedHostNetwork: false,
			expectedDnsPolicy:   corev1.DNSClusterFirst,
		},
		{
			name:                "Custom",
			hostNetwork:         true,
			expectedHostNetwork: true,
			dnsPolicy:           corev1.DNSClusterFirstWithHostNet,
			expectedDnsPolicy:   corev1.DNSClusterFirstWithHostNet,
		},
	}

	for _, test := range tests {
		assert := assert.New(t)

		rf := generateRF()
		rf.Spec.Redis.HostNetwork = test.hostNetwork
		rf.Spec.Redis.DNSPolicy = test.dnsPolicy

		var actualHostNetwork bool
		var actualDnsPolicy corev1.DNSPolicy

		ms := &mK8SService.Services{}
		ms.On("CreateOrUpdatePodDisruptionBudget", namespace, mock.Anything).Once().Return(nil, nil)
		ms.On("CreateOrUpdateStatefulSet", namespace, mock.Anything).Once().Run(func(args mock.Arguments) {
			ss := args.Get(1).(*appsv1.StatefulSet)
			actualHostNetwork = ss.Spec.Template.Spec.HostNetwork
			actualDnsPolicy = ss.Spec.Template.Spec.DNSPolicy
		}).Return(nil)

		client := rfservice.NewRedisFailoverKubeClient(ms, log.Dummy)
		err := client.EnsureRedisStatefulset(rf, nil, []metav1.OwnerReference{})
		assert.NoError(err)

		assert.Equal(test.expectedHostNetwork, actualHostNetwork)
		assert.Equal(test.expectedDnsPolicy, actualDnsPolicy)
	}
}

func TestSentinelHostNetworkAndDnsPolicy(t *testing.T) {
	tests := []struct {
		name                string
		hostNetwork         bool
		expectedHostNetwork bool
		dnsPolicy           corev1.DNSPolicy
		expectedDnsPolicy   corev1.DNSPolicy
	}{
		{
			name:                "Default",
			expectedHostNetwork: false,
			expectedDnsPolicy:   corev1.DNSClusterFirst,
		},
		{
			name:                "Custom",
			hostNetwork:         true,
			expectedHostNetwork: true,
			dnsPolicy:           corev1.DNSClusterFirstWithHostNet,
			expectedDnsPolicy:   corev1.DNSClusterFirstWithHostNet,
		},
	}

	for _, test := range tests {
		assert := assert.New(t)

		rf := generateRF()
		rf.Spec.Sentinel.HostNetwork = test.hostNetwork
		rf.Spec.Sentinel.DNSPolicy = test.dnsPolicy

		var actualHostNetwork bool
		var actualDnsPolicy corev1.DNSPolicy

		ms := &mK8SService.Services{}
		ms.On("CreateOrUpdatePodDisruptionBudget", namespace, mock.Anything).Once().Return(nil, nil)
		ms.On("CreateOrUpdateDeployment", namespace, mock.Anything).Once().Run(func(args mock.Arguments) {
			d := args.Get(1).(*appsv1.Deployment)
			actualHostNetwork = d.Spec.Template.Spec.HostNetwork
			actualDnsPolicy = d.Spec.Template.Spec.DNSPolicy
		}).Return(nil)

		client := rfservice.NewRedisFailoverKubeClient(ms, log.Dummy)
		err := client.EnsureSentinelDeployment(rf, nil, []metav1.OwnerReference{})
		assert.NoError(err)

		assert.Equal(test.expectedHostNetwork, actualHostNetwork)
		assert.Equal(test.expectedDnsPolicy, actualDnsPolicy)
	}
}

func TestRedisImagePullPolicy(t *testing.T) {
	tests := []struct {
		name                   string
		policy                 corev1.PullPolicy
		exporterPolicy         corev1.PullPolicy
		expectedPolicy         corev1.PullPolicy
		expectedExporterPolicy corev1.PullPolicy
	}{
		{
			name:                   "Default",
			expectedPolicy:         corev1.PullAlways,
			expectedExporterPolicy: corev1.PullAlways,
		},
		{
			name:                   "Custom",
			policy:                 corev1.PullIfNotPresent,
			exporterPolicy:         corev1.PullNever,
			expectedPolicy:         corev1.PullIfNotPresent,
			expectedExporterPolicy: corev1.PullNever,
		},
	}

	for _, test := range tests {
		assert := assert.New(t)

		var policy corev1.PullPolicy
		var exporterPolicy corev1.PullPolicy

		rf := generateRF()
		rf.Spec.Redis.ImagePullPolicy = test.policy
		rf.Spec.Redis.Exporter.Enabled = true
		rf.Spec.Redis.Exporter.ImagePullPolicy = test.expectedExporterPolicy

		ms := &mK8SService.Services{}
		ms.On("CreateOrUpdatePodDisruptionBudget", namespace, mock.Anything).Once().Return(nil, nil)
		ms.On("CreateOrUpdateStatefulSet", namespace, mock.Anything).Once().Run(func(args mock.Arguments) {
			ss := args.Get(1).(*appsv1.StatefulSet)
			policy = ss.Spec.Template.Spec.Containers[0].ImagePullPolicy
			exporterPolicy = ss.Spec.Template.Spec.Containers[1].ImagePullPolicy
		}).Return(nil)

		client := rfservice.NewRedisFailoverKubeClient(ms, log.Dummy)
		err := client.EnsureRedisStatefulset(rf, nil, []metav1.OwnerReference{})

		assert.NoError(err)
		assert.Equal(string(test.expectedPolicy), string(policy))
		assert.Equal(string(test.expectedExporterPolicy), string(exporterPolicy))
	}
}

func TestSentinelImagePullPolicy(t *testing.T) {
	tests := []struct {
		name                 string
		policy               corev1.PullPolicy
		expectedPolicy       corev1.PullPolicy
		expectedConfigPolicy corev1.PullPolicy
	}{
		{
			name:                 "Default",
			expectedPolicy:       corev1.PullAlways,
			expectedConfigPolicy: corev1.PullAlways,
		},
		{
			name:                 "Custom",
			policy:               corev1.PullIfNotPresent,
			expectedPolicy:       corev1.PullIfNotPresent,
			expectedConfigPolicy: corev1.PullIfNotPresent,
		},
	}

	for _, test := range tests {
		assert := assert.New(t)

		var policy corev1.PullPolicy
		var configPolicy corev1.PullPolicy

		rf := generateRF()
		rf.Spec.Sentinel.ImagePullPolicy = test.policy

		ms := &mK8SService.Services{}
		ms.On("CreateOrUpdatePodDisruptionBudget", namespace, mock.Anything).Once().Return(nil, nil)
		ms.On("CreateOrUpdateDeployment", namespace, mock.Anything).Once().Run(func(args mock.Arguments) {
			d := args.Get(1).(*appsv1.Deployment)
			policy = d.Spec.Template.Spec.Containers[0].ImagePullPolicy
			configPolicy = d.Spec.Template.Spec.InitContainers[0].ImagePullPolicy
		}).Return(nil)

		client := rfservice.NewRedisFailoverKubeClient(ms, log.Dummy)
		err := client.EnsureSentinelDeployment(rf, nil, []metav1.OwnerReference{})

		assert.NoError(err)
		assert.Equal(string(test.expectedPolicy), string(policy))
		assert.Equal(string(test.expectedConfigPolicy), string(configPolicy))
	}
}
