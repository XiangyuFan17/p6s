package k8s

import (
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// K8sClient encapsulates Kubernetes client operations
type K8sClient struct {
	clientset *kubernetes.Clientset
	context   string
}

// NewK8sClient creates a new Kubernetes client
func NewK8sClient() *K8sClient {
	return &K8sClient{}
}

// Connect connects to Kubernetes cluster
func (k *K8sClient) Connect() error {
	// Get kubeconfig file path
	home := homedir.HomeDir()
	kubeconfig := filepath.Join(home, ".kube", "config")

	// Check if file exists
	if _, err := os.Stat(kubeconfig); os.IsNotExist(err) {
		return fmt.Errorf("kubeconfig file does not exist: %s", kubeconfig)
	}

	// Load kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return fmt.Errorf("unable to load kubeconfig: %v", err)
	}

	// Create clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("unable to create Kubernetes client: %v", err)
	}

	// Get current context
	rawConfig, err := clientcmd.LoadFromFile(kubeconfig)
	if err != nil {
		return fmt.Errorf("unable to read kubeconfig: %v", err)
	}

	k.clientset = clientset
	k.context = rawConfig.CurrentContext

	return nil
}

// GetCurrentContext gets current Kubernetes context
func (k *K8sClient) GetCurrentContext() string {
	return k.context
}

// IsConnected checks if connected to cluster
func (k *K8sClient) IsConnected() bool {
	return k.clientset != nil
}

// GetClientset gets Kubernetes clientset
func (k *K8sClient) GetClientset() *kubernetes.Clientset {
	return k.clientset
}

// Close closes Kubernetes client connection
func (k *K8sClient) Close() {
	// client-go has no explicit close method, just set to nil
	k.clientset = nil
}