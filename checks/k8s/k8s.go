package k8s

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

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
	checks.Register("k8s.namespace_access", "Verifies access to a Kubernetes namespace", NamespaceAccess)
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

// NamespaceAccess checks if the current user has access to list pods in the specified namespace
func NamespaceAccess(params map[string]interface{}) (map[string]interface{}, error) {
	// Get parameters with defaults
	namespace := "default"
	if ns, ok := params["namespace"].(string); ok && ns != "" {
		namespace = ns
	}

	contextName := ""
	if ctx, ok := params["context"].(string); ok && ctx != "" {
		contextName = ctx
	}

	// Create kubernetes config
	config, err := newKubeConfig(contextName)
	if err != nil {
		return nil, fmt.Errorf("error creating kubernetes config: %v", err)
	}

	// Create kubernetes clientset
	clientset, err := newClientset(config)
	if err != nil {
		return nil, fmt.Errorf("error creating kubernetes clientset: %v", err)
	}

	// Try to list pods in the namespace to verify access
	ctx := context.Background()
	_, err = clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{Limit: 1})
	if err != nil {
		return map[string]interface{}{
			"status": "Failure",
			"output": fmt.Sprintf("Error accessing namespace %s: %v", namespace, err),
		}, nil
	}

	// Get current context name
	rawConfig, err := config.RawConfig()
	if err != nil {
		return nil, fmt.Errorf("error getting current context: %v", err)
	}
	currentContext := rawConfig.CurrentContext

	return map[string]interface{}{
		"status": "Success",
		"output": fmt.Sprintf("Successfully verified access to namespace %s in context %s", namespace, currentContext),
	}, nil
}
