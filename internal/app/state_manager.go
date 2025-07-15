package app

import (
	"p6s/internal/k8s"
	v1 "k8s.io/api/core/v1"
)

// StateManager manages application state
type StateManager struct {
	currentNamespace string
	currentPods      []k8s.Pod
	selectedPod      *v1.Pod
	selectedContainer *v1.Container
	currentSecrets   []k8s.Secret
	selectedSecret   *k8s.Secret
}

// NewStateManager creates a new state manager
func NewStateManager() *StateManager {
	return &StateManager{}
}

// SetCurrentNamespace sets the current namespace
func (sm *StateManager) SetCurrentNamespace(namespace string) {
	sm.currentNamespace = namespace
}

// GetCurrentNamespace gets the current namespace
func (sm *StateManager) GetCurrentNamespace() string {
	return sm.currentNamespace
}

// SetCurrentPods sets the current pod list
func (sm *StateManager) SetCurrentPods(pods []k8s.Pod) {
	sm.currentPods = pods

	sm.selectedPod = nil
	sm.selectedContainer = nil
}

// GetCurrentPods gets the current pod list
func (sm *StateManager) GetCurrentPods() []k8s.Pod {
	return sm.currentPods
}

// SetSelectedPod sets the selected pod
func (sm *StateManager) SetSelectedPod(pod *v1.Pod) {
	sm.selectedPod = pod

	sm.selectedContainer = nil
}

// GetSelectedPod gets the selected pod
func (sm *StateManager) GetSelectedPod() *v1.Pod {
	return sm.selectedPod
}

// SetSelectedContainer sets the selected container
func (sm *StateManager) SetSelectedContainer(container *v1.Container) {
	sm.selectedContainer = container
}

// GetSelectedContainer gets the selected container
func (sm *StateManager) GetSelectedContainer() *v1.Container {
	return sm.selectedContainer
}

// SetCurrentSecrets sets the current secret list
func (sm *StateManager) SetCurrentSecrets(secrets []k8s.Secret) {
	sm.currentSecrets = secrets

	sm.selectedSecret = nil
}

// GetCurrentSecrets gets the current secret list
func (sm *StateManager) GetCurrentSecrets() []k8s.Secret {
	return sm.currentSecrets
}

// SetSelectedSecret sets the selected secret
func (sm *StateManager) SetSelectedSecret(secret *k8s.Secret) {
	sm.selectedSecret = secret
}

// GetSelectedSecret gets the selected secret
func (sm *StateManager) GetSelectedSecret() *k8s.Secret {
	return sm.selectedSecret
}

// Reset resets all state
func (sm *StateManager) Reset() {
	sm.currentNamespace = ""
	sm.currentPods = nil
	sm.selectedPod = nil
	sm.selectedContainer = nil
	sm.currentSecrets = nil
	sm.selectedSecret = nil
}