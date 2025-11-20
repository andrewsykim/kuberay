package get

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	kubefake "k8s.io/client-go/kubernetes/fake"

	"github.com/ray-project/kuberay/kubectl-plugin/pkg/util/client"

	rayv1 "github.com/ray-project/kuberay/ray-operator/apis/ray/v1"
	rayClientFake "github.com/ray-project/kuberay/ray-operator/pkg/client/clientset/versioned/fake"
)

func TestGetToken(t *testing.T) {
	testCases := []struct {
		name          string
		clusterName   string
		namespace     string
		cluster       *rayv1.RayCluster
		secret        *corev1.Secret
		expectedToken string
		expectedError string
	}{
		{
			name:        "successfully get token",
			clusterName: "test-cluster",
			namespace:   "default",
			cluster: &rayv1.RayCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: "default",
				},
			},
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: "default",
				},
				Data: map[string][]byte{
					"auth_token": []byte("test-token"),
				},
			},
			expectedToken: "test-token\n",
		},
		{
			name:          "RayCluster not found",
			clusterName:   "test-cluster",
			namespace:     "default",
			expectedError: "failed to get RayCluster test-cluster in namespace default: rayclusters.ray.io \"test-cluster\" not found",
		},
		{
			name:        "secret not found",
			clusterName: "test-cluster",
			namespace:   "default",
			cluster: &rayv1.RayCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: "default",
				},
			},
			expectedError: "failed to get secret test-cluster in namespace default: secrets \"test-cluster\" not found",
		},
		{
			name:        "secret does not contain auth_token",
			clusterName: "test-cluster",
			namespace:   "default",
			cluster: &rayv1.RayCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: "default",
				},
			},
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: "default",
				},
				Data: map[string][]byte{},
			},
			expectedError: "secret test-cluster in namespace default does not contain 'auth_token'",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var kubeObjs []runtime.Object
			if tc.secret != nil {
				kubeObjs = append(kubeObjs, tc.secret)
			}
			var rayObjs []runtime.Object
			if tc.cluster != nil {
				rayObjs = append(rayObjs, tc.cluster)
			}
			kubeClientSet := kubefake.NewSimpleClientset(kubeObjs...)
			rayClient := rayClientFake.NewSimpleClientset(rayObjs...)
			k8sClient := client.NewClientForTesting(kubeClientSet, rayClient)
			streams := genericclioptions.IOStreams{
				Out:    &bytes.Buffer{},
				ErrOut: &bytes.Buffer{},
			}

			options := &GetTokenOptions{
				ioStreams: &streams,
				namespace: tc.namespace,
				cluster:   tc.clusterName,
			}

			err := options.Run(context.Background(), k8sClient)
			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedToken, streams.Out.(*bytes.Buffer).String())
			}
		})
	}
}
