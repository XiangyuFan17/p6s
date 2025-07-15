package k8s

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Secret represents Kubernetes Secret information
type Secret struct {
	Name       string
	Namespace  string
	Type       string
	Data       map[string]string
	CreateTime time.Time
}

// GetSecrets gets all Secrets in specified namespace
func (k *K8sClient) GetSecrets(namespace string) ([]Secret, error) {
	if !k.IsConnected() {
		return nil, fmt.Errorf("not connected to Kubernetes cluster")
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get Secret list
	secretList, err := k.clientset.CoreV1().Secrets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get Secret list: %v", err)
	}

	// Convert to custom Secret structure
	result := make([]Secret, 0, len(secretList.Items))
	for _, s := range secretList.Items {
		// Convert Secret data
		decodedData := make(map[string]string)
		for key, value := range s.Data {
			// Secret.Data in Kubernetes Go client is []byte type, convert directly to string
			decodedData[key] = string(value)
		}

		secret := Secret{
			Name:       s.Name,
			Namespace:  s.Namespace,
			Type:       string(s.Type),
			Data:       decodedData,
			CreateTime: s.CreationTimestamp.Time,
		}

		result = append(result, secret)
	}

	return result, nil
}

// GetSecret gets specific Secret in specified namespace
func (k *K8sClient) GetSecret(namespace, name string) (*Secret, error) {
	if !k.IsConnected() {
		return nil, fmt.Errorf("not connected to Kubernetes cluster")
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get specified Secret
	s, err := k.clientset.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get Secret: %v", err)
	}

	// Convert Secret data
	decodedData := make(map[string]string)
	for key, value := range s.Data {
		// Secret.Data in Kubernetes Go client is []byte type, convert directly to string
		decodedData[key] = string(value)
	}

	secret := &Secret{
		Name:       s.Name,
		Namespace:  s.Namespace,
		Type:       string(s.Type),
		Data:       decodedData,
		CreateTime: s.CreationTimestamp.Time,
	}

	return secret, nil
}

// GetNamespaces gets all namespaces
func (k *K8sClient) GetNamespaces() ([]string, error) {
	if !k.IsConnected() {
		return nil, fmt.Errorf("not connected to Kubernetes cluster")
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get namespace list
	nsList, err := k.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get namespace list: %v", err)
	}

	// Extract namespace names
	result := make([]string, 0, len(nsList.Items))
	for _, ns := range nsList.Items {
		result = append(result, ns.Name)
	}

	return result, nil
}