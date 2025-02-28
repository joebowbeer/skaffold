/*
Copyright 2019 The Skaffold Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cluster

import (
	"testing"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/build/kaniko"
	latest_v1 "github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/latest/v1"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/util"
	"github.com/GoogleContainerTools/skaffold/testutil"
)

func TestKanikoArgs(t *testing.T) {
	tests := []struct {
		description        string
		artifact           *latest_v1.KanikoArtifact
		insecureRegistries map[string]bool
		tag                string
		shouldErr          bool
		expectedArgs       []string
	}{
		{
			description: "simple build",
			artifact: &latest_v1.KanikoArtifact{
				DockerfilePath: "Dockerfile",
			},
			expectedArgs: []string{},
		},
		{
			description: "cache layers",
			artifact: &latest_v1.KanikoArtifact{
				DockerfilePath: "Dockerfile",
				Cache:          &latest_v1.KanikoCache{},
			},
			expectedArgs: []string{kaniko.CacheFlag},
		},
		{
			description: "cache layers to specific repo",
			artifact: &latest_v1.KanikoArtifact{
				DockerfilePath: "Dockerfile",
				Cache: &latest_v1.KanikoCache{
					Repo: "repo",
				},
			},
			expectedArgs: []string{"--cache", kaniko.CacheRepoFlag, "repo"},
		},
		{
			description: "cache path",
			artifact: &latest_v1.KanikoArtifact{
				DockerfilePath: "Dockerfile",
				Cache: &latest_v1.KanikoCache{
					HostPath: "/cache",
				},
			},
			expectedArgs: []string{
				kaniko.CacheFlag,
				kaniko.CacheDirFlag, "/cache"},
		},
		{
			description: "target",
			artifact: &latest_v1.KanikoArtifact{
				DockerfilePath: "Dockerfile",
				Target:         "target",
			},
			expectedArgs: []string{kaniko.TargetFlag, "target"},
		},
		{
			description: "reproducible",
			artifact: &latest_v1.KanikoArtifact{
				DockerfilePath: "Dockerfile",
				Reproducible:   true,
			},
			expectedArgs: []string{kaniko.ReproducibleFlag},
		},
		{
			description: "build args",
			artifact: &latest_v1.KanikoArtifact{
				DockerfilePath: "Dockerfile",
				BuildArgs: map[string]*string{
					"nil_key":   nil,
					"empty_key": util.StringPtr(""),
					"value_key": util.StringPtr("value"),
				},
			},
			expectedArgs: []string{
				kaniko.BuildArgsFlag, "empty_key=",
				kaniko.BuildArgsFlag, "nil_key",
				kaniko.BuildArgsFlag, "value_key=value"},
		},
		{
			description: "invalid build args",
			artifact: &latest_v1.KanikoArtifact{
				DockerfilePath: "Dockerfile",
				BuildArgs: map[string]*string{
					"invalid": util.StringPtr("{{Invalid"),
				},
			},
			shouldErr: true,
		},
		{
			description: "insecure registries",
			artifact: &latest_v1.KanikoArtifact{
				DockerfilePath: "Dockerfile",
			},
			insecureRegistries: map[string]bool{"localhost:4000": true},
			expectedArgs:       []string{kaniko.InsecureRegistryFlag, "localhost:4000"},
		},
		{
			description: "skip tls",
			artifact: &latest_v1.KanikoArtifact{
				DockerfilePath: "Dockerfile",
				SkipTLS:        true,
			},
			expectedArgs: []string{
				kaniko.SkipTLSFlag,
				kaniko.SkipTLSVerifyRegistryFlag, "gcr.io",
			},
		},
		{
			description: "invalid registry",
			artifact: &latest_v1.KanikoArtifact{
				DockerfilePath: "Dockerfile",
				SkipTLS:        true,
			},
			tag:       "!!!!",
			shouldErr: true,
		},
	}
	for _, test := range tests {
		testutil.Run(t, test.description, func(t *testutil.T) {
			commonArgs := []string{"--destination", "gcr.io/tag", "--dockerfile", "Dockerfile", "--context", "dir:///kaniko/buildcontext"}

			tag := "gcr.io/tag"
			if test.tag != "" {
				tag = test.tag
			}
			args, err := kanikoArgs(test.artifact, tag, test.insecureRegistries)

			t.CheckError(test.shouldErr, err)
			if !test.shouldErr {
				t.CheckDeepEqual(append(commonArgs, test.expectedArgs...), args)
			}
		})
	}
}

func TestKanikoPodSpec(t *testing.T) {
	artifact := &latest_v1.KanikoArtifact{
		Image:          "image",
		DockerfilePath: "Dockerfile",
		InitImage:      "init/image",
		Env: []v1.EnvVar{{
			Name:  "KEY",
			Value: "VALUE",
		}},
		VolumeMounts: []v1.VolumeMount{
			{
				Name:      "cm-volume-1",
				ReadOnly:  true,
				MountPath: "/cm-test-mount-path",
				SubPath:   "/subpath",
			},
			{
				Name:      "secret-volume-1",
				ReadOnly:  true,
				MountPath: "/secret-test-mount-path",
				SubPath:   "/subpath",
			},
		},
	}

	var runAsUser int64 = 0

	builder := &Builder{
		cfg: &mockBuilderContext{},
		ClusterDetails: &latest_v1.ClusterDetails{
			Namespace:           "ns",
			PullSecretName:      "secret",
			PullSecretPath:      "kaniko-secret.json",
			PullSecretMountPath: "/secret",
			HTTPProxy:           "http://proxy",
			HTTPSProxy:          "https://proxy",
			ServiceAccountName:  "aVerySpecialSA",
			RunAsUser:           &runAsUser,
			Resources: &latest_v1.ResourceRequirements{
				Requests: &latest_v1.ResourceRequirement{
					CPU: "0.1",
				},
				Limits: &latest_v1.ResourceRequirement{
					CPU: "0.5",
				},
			},
			Volumes: []v1.Volume{
				{
					Name: "cm-volume-1",
					VolumeSource: v1.VolumeSource{
						ConfigMap: &v1.ConfigMapVolumeSource{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "cm-1",
							},
						},
					},
				},
				{
					Name: "secret-volume-1",
					VolumeSource: v1.VolumeSource{
						Secret: &v1.SecretVolumeSource{
							SecretName: "secret-1",
						},
					},
				},
			},
			Tolerations: []v1.Toleration{
				{
					Key:               "app",
					Operator:          "Equal",
					Value:             "skaffold",
					Effect:            "NoSchedule",
					TolerationSeconds: nil,
				},
			},
		},
	}
	pod, _ := builder.kanikoPodSpec(artifact, "tag")

	expectedPod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Annotations:  map[string]string{"test": "test"},
			GenerateName: "kaniko-",
			Labels:       map[string]string{"skaffold-kaniko": "skaffold-kaniko"},
			Namespace:    "ns",
		},
		Spec: v1.PodSpec{
			InitContainers: []v1.Container{{
				Name:    initContainer,
				Image:   "init/image",
				Command: []string{"sh", "-c", "while [ ! -f /tmp/complete ]; do sleep 1; done"},
				VolumeMounts: []v1.VolumeMount{{
					Name:      kaniko.DefaultEmptyDirName,
					MountPath: kaniko.DefaultEmptyDirMountPath,
				}, {
					Name:      "cm-volume-1",
					ReadOnly:  true,
					MountPath: "/cm-secret-mount-path",
					SubPath:   "/subpath",
				}, {
					Name:      "secret-volume-1",
					ReadOnly:  true,
					MountPath: "/secret-secret-mount-path",
					SubPath:   "/subpath",
				}},
				Resources: v1.ResourceRequirements{
					Requests: map[v1.ResourceName]resource.Quantity{
						v1.ResourceCPU: resource.MustParse("0.1"),
					},
					Limits: v1.ResourceList{
						v1.ResourceCPU: resource.MustParse("0.5"),
					},
				},
			}},
			Containers: []v1.Container{{
				Name:            kaniko.DefaultContainerName,
				Image:           "image",
				Args:            []string{"--dockerfile", "Dockerfile", "--context", "dir:///kaniko/buildcontext", "--destination", "tag", "-v", "info"},
				ImagePullPolicy: v1.PullIfNotPresent,
				Env: []v1.EnvVar{{
					Name:  "GOOGLE_APPLICATION_CREDENTIALS",
					Value: "/secret/kaniko-secret.json",
				}, {
					Name:  "UPSTREAM_CLIENT_TYPE",
					Value: "UpstreamClient(skaffold-)",
				}, {
					Name:  "KEY",
					Value: "VALUE",
				}, {
					Name:  "HTTP_PROXY",
					Value: "http://proxy",
				}, {
					Name:  "HTTPS_PROXY",
					Value: "https://proxy",
				}},
				VolumeMounts: []v1.VolumeMount{
					{
						Name:      kaniko.DefaultEmptyDirName,
						MountPath: kaniko.DefaultEmptyDirMountPath,
					},
					{
						Name:      kaniko.DefaultSecretName,
						MountPath: "/secret",
					},
					{
						Name:      "cm-volume-1",
						ReadOnly:  true,
						MountPath: "/cm-secret-mount-path",
						SubPath:   "/subpath",
					},
					{
						Name:      "secret-volume-1",
						ReadOnly:  true,
						MountPath: "/secret-secret-mount-path",
						SubPath:   "/subpath",
					},
				},
				Resources: v1.ResourceRequirements{
					Requests: map[v1.ResourceName]resource.Quantity{
						v1.ResourceCPU: resource.MustParse("0.1"),
					},
					Limits: v1.ResourceList{
						v1.ResourceCPU: resource.MustParse("0.5"),
					},
				},
			}},
			ServiceAccountName: "aVerySpecialSA",
			SecurityContext: &v1.PodSecurityContext{
				RunAsUser: &runAsUser,
			},
			RestartPolicy: v1.RestartPolicyNever,
			Volumes: []v1.Volume{
				{
					Name: kaniko.DefaultEmptyDirName,
					VolumeSource: v1.VolumeSource{
						EmptyDir: &v1.EmptyDirVolumeSource{},
					},
				},
				{
					Name: kaniko.DefaultSecretName,
					VolumeSource: v1.VolumeSource{
						Secret: &v1.SecretVolumeSource{
							SecretName: "secret",
						},
					},
				},
				{
					Name: "cm-volume-1",
					VolumeSource: v1.VolumeSource{
						ConfigMap: &v1.ConfigMapVolumeSource{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "cm-1",
							},
						},
					},
				},
				{
					Name: "secret-volume-1",
					VolumeSource: v1.VolumeSource{
						Secret: &v1.SecretVolumeSource{
							SecretName: "secret-1",
						},
					},
				},
			},
			Tolerations: []v1.Toleration{
				{
					Key:               "app",
					Operator:          "Equal",
					Value:             "skaffold",
					Effect:            "NoSchedule",
					TolerationSeconds: nil,
				},
			},
		},
	}

	testutil.CheckDeepEqual(t, expectedPod.Spec.Containers[0].Env, pod.Spec.Containers[0].Env)
}

func TestResourceRequirements(t *testing.T) {
	tests := []struct {
		description string
		initial     *latest_v1.ResourceRequirements
		expected    v1.ResourceRequirements
	}{
		{
			description: "no resource specified",
			initial:     &latest_v1.ResourceRequirements{},
			expected:    v1.ResourceRequirements{},
		},
		{
			description: "with resource specified",
			initial: &latest_v1.ResourceRequirements{
				Requests: &latest_v1.ResourceRequirement{
					CPU:              "0.5",
					Memory:           "1000",
					ResourceStorage:  "1000",
					EphemeralStorage: "1000",
				},
				Limits: &latest_v1.ResourceRequirement{
					CPU:              "1.0",
					Memory:           "2000",
					ResourceStorage:  "1000",
					EphemeralStorage: "1000",
				},
			},
			expected: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceCPU:              resource.MustParse("0.5"),
					v1.ResourceMemory:           resource.MustParse("1000"),
					v1.ResourceStorage:          resource.MustParse("1000"),
					v1.ResourceEphemeralStorage: resource.MustParse("1000"),
				},
				Limits: v1.ResourceList{
					v1.ResourceCPU:              resource.MustParse("1.0"),
					v1.ResourceMemory:           resource.MustParse("2000"),
					v1.ResourceStorage:          resource.MustParse("1000"),
					v1.ResourceEphemeralStorage: resource.MustParse("1000"),
				},
			},
		},
	}

	for _, test := range tests {
		testutil.Run(t, test.description, func(t *testutil.T) {
			actual := resourceRequirements(test.initial)
			t.CheckDeepEqual(test.expected, actual)
		})
	}
}
