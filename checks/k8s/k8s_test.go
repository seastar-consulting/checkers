package k8s

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"

	"github.com/seastar-consulting/checkers/checks"
)

// Save original functions for testing
var (
	originalNewKubeConfig = newKubeConfig
	originalNewClientset  = newClientset
)

func TestNamespaceAccess(t *testing.T) {
	// Save original functions and restore them after test
	defer func() {
		newKubeConfig = originalNewKubeConfig
		newClientset = originalNewClientset
	}()

	tests := []struct {
		name          string
		params        map[string]interface{}
		mockContext   string
		listPodsError error
		want          map[string]interface{}
		wantErr       bool
	}{
		{
			name:        "default namespace and context",
			params:      map[string]interface{}{},
			mockContext: "test-context",
			want: map[string]interface{}{
				"status": "Success",
				"output": "Successfully verified access to namespace default in context test-context",
			},
		},
		{
			name: "custom namespace and context",
			params: map[string]interface{}{
				"namespace": "test-ns",
				"context":   "custom-context",
			},
			mockContext: "custom-context",
			want: map[string]interface{}{
				"status": "Success",
				"output": "Successfully verified access to namespace test-ns in context custom-context",
			},
		},
		{
			name: "access denied error",
			params: map[string]interface{}{
				"namespace": "restricted",
			},
			mockContext:   "test-context",
			listPodsError: fmt.Errorf("access denied"),
			want: map[string]interface{}{
				"status": "Failure",
				"output": "Error accessing namespace restricted: access denied",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock kubernetes config
			newKubeConfig = func(context string) (clientcmd.ClientConfig, error) {
				return clientcmd.NewDefaultClientConfig(api.Config{
					CurrentContext: tt.mockContext,
				}, &clientcmd.ConfigOverrides{}), nil
			}

			// Mock kubernetes clientset
			newClientset = func(config clientcmd.ClientConfig) (kubernetes.Interface, error) {
				if tt.listPodsError != nil {
					// Create a clientset that will return an error when listing pods
					clientset := fake.NewSimpleClientset()
					// Inject the error by returning it directly
					return &mockClientset{
						Clientset: clientset,
						err:       tt.listPodsError,
					}, nil
				}
				return fake.NewSimpleClientset(), nil
			}

			// Get the check
			check, err := checks.Get("k8s.namespace_access")
			require.NoError(t, err)
			require.Equal(t, "k8s.namespace_access", check.Name)
			require.Equal(t, "Verifies access to a Kubernetes namespace", check.Description)

			// Run the check
			result, err := check.Func(tt.params)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, result)
		})
	}
}

// mockClientset wraps a fake clientset and injects errors
type mockClientset struct {
	*fake.Clientset
	err error
}

// CoreV1 returns a mocked CoreV1Client that returns errors when listing pods
func (m *mockClientset) CoreV1() corev1.CoreV1Interface {
	return &mockCoreV1Client{
		CoreV1Interface: m.Clientset.CoreV1(),
		err:            m.err,
	}
}

// mockCoreV1Client wraps a fake CoreV1Client and injects errors
type mockCoreV1Client struct {
	corev1.CoreV1Interface
	err error
}

// Pods returns a mocked PodInterface that returns errors when listing pods
func (m *mockCoreV1Client) Pods(namespace string) corev1.PodInterface {
	return &mockPodInterface{
		PodInterface: m.CoreV1Interface.Pods(namespace),
		err:         m.err,
	}
}

// mockPodInterface wraps a fake PodInterface and injects errors
type mockPodInterface struct {
	corev1.PodInterface
	err error
}

// List returns the injected error
func (m *mockPodInterface) List(ctx context.Context, opts metav1.ListOptions) (*v1.PodList, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.PodInterface.List(ctx, opts)
}
