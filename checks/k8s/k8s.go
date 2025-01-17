package k8s

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/seastar-consulting/checkers/types"

	"github.com/seastar-consulting/checkers/checks"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var (
	newKubeConfig = defaultNewKubeConfig
	newClientset  = defaultNewClientset
)

func init() {
	checks.Register("k8s.namespace_access", "Verifies access to a Kubernetes namespace", CheckNamespaceAccess)
}

// defaultNewKubeConfig creates a new kubernetes config from the given context
func defaultNewKubeConfig(contextName string) (clientcmd.ClientConfig, error) {
	kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
	if envVar := os.Getenv("KUBECONFIG"); envVar != "" {
		kubeconfig = envVar
	}

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.ExplicitPath = kubeconfig
	configOverrides := &clientcmd.ConfigOverrides{}

	if contextName != "" {
		configOverrides.CurrentContext = contextName
	}

	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides), nil
}

// defaultNewClientset creates a new kubernetes clientset from the given config
func defaultNewClientset(config clientcmd.ClientConfig) (kubernetes.Interface, error) {
	c, err := config.ClientConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(c)
}

// CheckNamespaceAccess checks if the current user has access to list pods in the specified namespace
// CheckNamespaceAccess implements the CheckFunc interface and verifies access to a Kubernetes namespace
func CheckNamespaceAccess(item types.CheckItem) (types.CheckResult, error) {
	const defaultNamespace = "default"

	// Helper function to retrieve string parameters with a default fallback
	getStringParam := func(key, defaultValue string) string {
		if value, ok := item.Parameters[key]; ok && value != "" {
			return value
		}
		return defaultValue
	}

	// Retrieve parameters
	namespaceParam := getStringParam("namespace", defaultNamespace)
	contextParam := getStringParam("context", "")

	// Create Kubernetes config
	kubeConfig, err := newKubeConfig(contextParam)
	if err != nil {
		return types.CheckResult{
			Name:   item.Name,
			Type:   item.Type,
			Status: types.Error,
			Error:  fmt.Sprintf("failed to create Kubernetes config: %v", err),
		}, nil
	}

	// Get current context early, before using the config
	rawConfig, err := kubeConfig.RawConfig()
	if err != nil {
		return types.CheckResult{
			Name:   item.Name,
			Type:   item.Type,
			Status: types.Error,
			Error:  fmt.Sprintf("failed to retrieve current context from config: %v", err),
		}, nil
	}
	currentContext := rawConfig.CurrentContext

	// Create Kubernetes clientset
	clientset, err := newClientset(kubeConfig)
	if err != nil {
		return types.CheckResult{
			Name:   item.Name,
			Type:   item.Type,
			Status: types.Error,
			Error:  fmt.Sprintf("failed to create Kubernetes clientset: %v", err),
		}, nil
	}

	// Attempt to list pods in the specified namespace
	ctx := context.Background()
	_, err = clientset.CoreV1().Pods(namespaceParam).List(ctx, metav1.ListOptions{Limit: 1})
	if err != nil {
		// Check if this is a permission-related error
		if strings.Contains(err.Error(), "forbidden") || 
		   strings.Contains(err.Error(), "unauthorized") ||
		   strings.Contains(err.Error(), "access denied") {
			return types.CheckResult{
				Name:   item.Name,
				Type:   item.Type,
				Status: types.Failure,
				Output: fmt.Sprintf("No access to namespace '%s': %v", namespaceParam, err),
			}, nil
		}
		// For other errors (like namespace not found, network issues, etc.), return error
		return types.CheckResult{
			Name:   item.Name,
			Type:   item.Type,
			Status: types.Error,
			Error:  fmt.Sprintf("error while accessing namespace '%s': %v", namespaceParam, err),
		}, nil
	}

	// Return success with access details
	return types.CheckResult{
		Name:   item.Name,
		Type:   item.Type,
		Status: types.Success,
		Output: fmt.Sprintf("Successfully verified access to namespace '%s' in context '%s'", namespaceParam, currentContext),
	}, nil
}
