package k8s

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Pod represents Kubernetes Pod information
type Pod struct {
	Name       string
	Namespace  string
	Status     string
	PodIP      string
	Containers []Container
	CreateTime time.Time
}

// Container represents container information in Pod
type Container struct {
	Name  string
	Image string
	Ports []ContainerPort
}

// ContainerPort represents container port information
type ContainerPort struct {
	Name          string
	ContainerPort int32
	Protocol      string
}

// GetPods gets all Pods in specified namespace
func (k *K8sClient) GetPods(namespace string) ([]Pod, error) {
	if !k.IsConnected() {
		return nil, fmt.Errorf("not connected to Kubernetes cluster")
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get Pod list
	podList, err := k.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get Pod list: %v", err)
	}

	// Convert to custom Pod structure
	result := make([]Pod, 0, len(podList.Items))
	for _, p := range podList.Items {
		// Extract container information
		containers := make([]Container, 0, len(p.Spec.Containers))
		for _, c := range p.Spec.Containers {
			// Extract port information
			ports := make([]ContainerPort, 0, len(c.Ports))
			for _, port := range c.Ports {
				ports = append(ports, ContainerPort{
					Name:          port.Name,
					ContainerPort: port.ContainerPort,
					Protocol:      string(port.Protocol),
				})
			}

			containers = append(containers, Container{
				Name:  c.Name,
				Image: c.Image,
				Ports: ports,
			})
		}

		pod := Pod{
			Name:       p.Name,
			Namespace:  p.Namespace,
			Status:     string(p.Status.Phase),
			PodIP:      p.Status.PodIP,
			Containers: containers,
			CreateTime: p.CreationTimestamp.Time,
		}

		result = append(result, pod)
	}

	return result, nil
}

// GetPod gets specific Pod in specified namespace
func (k *K8sClient) GetPod(namespace, name string) (*Pod, error) {
	if !k.IsConnected() {
		return nil, fmt.Errorf("not connected to Kubernetes cluster")
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get specified Pod
	p, err := k.clientset.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get Pod: %v", err)
	}

	// Extract container information
	containers := make([]Container, 0, len(p.Spec.Containers))
	for _, c := range p.Spec.Containers {
		// Extract port information
		ports := make([]ContainerPort, 0, len(c.Ports))
		for _, port := range c.Ports {
			ports = append(ports, ContainerPort{
				Name:          port.Name,
				ContainerPort: port.ContainerPort,
				Protocol:      string(port.Protocol),
			})
		}

		containers = append(containers, Container{
			Name:  c.Name,
			Image: c.Image,
			Ports: ports,
		})
	}

	pod := &Pod{
		Name:       p.Name,
		Namespace:  p.Namespace,
		Status:     string(p.Status.Phase),
		PodIP:      p.Status.PodIP,
		Containers: containers,
		CreateTime: p.CreationTimestamp.Time,
	}

	return pod, nil
}