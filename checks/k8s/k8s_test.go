package k8s

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"

	"github.com/seastar-consulting/checkers/types"
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
		checkItem     types.CheckItem
		mockContext   string
		listPodsError error
		want          types.CheckResult
		wantErr       bool
	}{
		{
			name: "default namespace and context",
			checkItem: types.CheckItem{
				Name: "test-check",
				Type: "k8s.namespace_access",
				Parameters: map[string]string{},
			},
			mockContext: "test-context",
			want: types.CheckResult{
				Name:   "test-check",
				Type:   "k8s.namespace_access",
				Status: types.Success,
				Output: "Successfully verified access to namespace 'default' in context 'test-context'",
			},
		},
		{
			name: "custom namespace and context",
			checkItem: types.CheckItem{
				Name: "test-check",
				Type: "k8s.namespace_access",
				Parameters: map[string]string{
					"namespace": "custom-ns",
					"context":   "custom-context",
				},
			},
			mockContext: "custom-context",
			want: types.CheckResult{
				Name:   "test-check",
				Type:   "k8s.namespace_access",
				Status: types.Success,
				Output: "Successfully verified access to namespace 'custom-ns' in context 'custom-context'",
			},
		},
		{
			name: "permission denied error",
			checkItem: types.CheckItem{
				Name: "test-check",
				Type: "k8s.namespace_access",
				Parameters: map[string]string{
					"namespace": "restricted-ns",
				},
			},
			mockContext:   "test-context",
			listPodsError: fmt.Errorf("pods is forbidden: User \"test\" cannot list resource \"pods\" in API group \"\" in the namespace \"restricted-ns\""),
			want: types.CheckResult{
				Name:   "test-check",
				Type:   "k8s.namespace_access",
				Status: types.Failure,
				Output: "No access to namespace 'restricted-ns': pods is forbidden: User \"test\" cannot list resource \"pods\" in API group \"\" in the namespace \"restricted-ns\"",
			},
		},
		{
			name: "unauthorized error",
			checkItem: types.CheckItem{
				Name: "test-check",
				Type: "k8s.namespace_access",
				Parameters: map[string]string{
					"namespace": "secure-ns",
				},
			},
			mockContext:   "test-context",
			listPodsError: fmt.Errorf("unauthorized: unable to verify user \"test\" in namespace \"secure-ns\""),
			want: types.CheckResult{
				Name:   "test-check",
				Type:   "k8s.namespace_access",
				Status: types.Failure,
				Output: "No access to namespace 'secure-ns': unauthorized: unable to verify user \"test\" in namespace \"secure-ns\"",
			},
		},
		{
			name: "non-permission error",
			checkItem: types.CheckItem{
				Name: "test-check",
				Type: "k8s.namespace_access",
				Parameters: map[string]string{
					"namespace": "missing-ns",
				},
			},
			mockContext:   "test-context",
			listPodsError: fmt.Errorf("namespace \"missing-ns\" not found"),
			want: types.CheckResult{
				Name:   "test-check",
				Type:   "k8s.namespace_access",
				Status: types.Error,
				Error:  "error while accessing namespace 'missing-ns': namespace \"missing-ns\" not found",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock kubeconfig
			newKubeConfig = func(contextName string) (clientcmd.ClientConfig, error) {
				return clientcmd.NewDefaultClientConfig(api.Config{
					CurrentContext: tt.mockContext,
				}, nil), nil
			}

			// Mock clientset
			newClientset = func(config clientcmd.ClientConfig) (kubernetes.Interface, error) {
				if tt.listPodsError != nil {
					return &mockClientset{
						Clientset: fake.NewSimpleClientset(),
						err:      tt.listPodsError,
					}, nil
				}
				return fake.NewSimpleClientset(), nil
			}

			got, err := CheckNamespaceAccess(tt.checkItem)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckNamespaceAccess() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
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
		err:             m.err,
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
		err:          m.err,
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
