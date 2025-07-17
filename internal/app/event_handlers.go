package app

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/rivo/tview"
	"p6s/internal/k8s"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EventHandlers event handler collection
type EventHandlers struct {
	app          *App
	errorHandler *ErrorHandler
}

// NewEventHandlers creates new event handlers
func NewEventHandlers(app *App) *EventHandlers {
	return &EventHandlers{
		app:          app,
		errorHandler: NewErrorHandler(app),
	}
}

// HandleNamespaceSelection handles namespace selection event
func (eh *EventHandlers) HandleNamespaceSelection(namespace string, namespaceDropdown, podDropdown, containerDropdown, portDropdown, secretDropdown, secretKeyDropdown *tview.DropDown, hostField, portField, usernameField, passwordField *tview.InputField) {

	
	// Set current namespace to state manager
	eh.app.stateManager.SetCurrentNamespace(namespace)
	
	// Reset all dropdowns
	eh.resetDropdowns(podDropdown, containerDropdown, portDropdown)
	eh.resetFields(hostField, portField)
	eh.resetSecretDropdowns(secretDropdown, secretKeyDropdown)
	// Remove resetAuthFields call to avoid clearing manually entered auth info when switching namespaces
	// eh.resetAuthFields(usernameField, passwordField)
	
	// Get Pod list
	pods, err := eh.app.k8sClient.GetPods(namespace)
	if err != nil {
		eh.errorHandler.HandleError(err, "Get Pod list")
		podDropdown.SetOptions([]string{fmt.Sprintf(MsgGetPodsError, err)}, nil)
		return
	}
	
	if len(pods) == 0 {
		podDropdown.SetOptions([]string{fmt.Sprintf(MsgNoPods, namespace)}, nil)
		return
	}
	
	// Update Pod list to state manager
	eh.app.stateManager.SetCurrentPods(pods)

	podNames := make([]string, len(pods))
	for i, p := range pods {
		podNames[i] = p.Name
	}
	
	podDropdown.SetOptions(podNames, nil)
	if len(podNames) > 0 {
		podDropdown.SetCurrentOption(0)
	}
	
	// Get Secret list
	secrets, err := eh.app.k8sClient.GetSecrets(namespace)
	if err != nil {
		eh.errorHandler.HandleError(err, "Get Secret list")
		secretDropdown.SetOptions([]string{fmt.Sprintf("Failed to get Secret: %v", err)}, nil)
	} else if len(secrets) == 0 {
		secretDropdown.SetOptions([]string{fmt.Sprintf("No Secret found in namespace %s", namespace)}, nil)
	} else {
		// Update Secret list to state manager
		eh.app.stateManager.SetCurrentSecrets(secrets)
		
		secretNames := make([]string, len(secrets))
		for i, s := range secrets {
			secretNames[i] = s.Name
		}
		
		secretDropdown.SetOptions(secretNames, nil)
		
		// Rebind Secret dropdown event handler
		secretDropdown.SetSelectedFunc(func(text string, index int) {
			eh.HandleSecretSelection(text, index, secretKeyDropdown, secretDropdown, usernameField, passwordField)
		})
	}
	
	// Reset Pod dropdown event handler as SetOptions clears previous handlers
	eh.setupPodDropdownHandler(podDropdown, containerDropdown, portDropdown, hostField, portField, namespaceDropdown, secretDropdown, secretKeyDropdown, usernameField, passwordField)
}

// HandlePodSelection handles Pod selection event
func (eh *EventHandlers) HandlePodSelection(podIndex int, containerDropdown, portDropdown *tview.DropDown, hostField, portField *tview.InputField, secretDropdown, secretKeyDropdown *tview.DropDown, usernameField, passwordField *tview.InputField) {

	
	// Get current Pod list
	pods := eh.app.stateManager.GetCurrentPods()
	if podIndex < 0 || podIndex >= len(pods) {

		return
	}
	
	// Get selected Pod (k8s.Pod type)
	selectedKubePod := pods[podIndex]
	// Convert to v1.Pod type for state management
	v1Pod := eh.convertK8sPodToV1Pod(selectedKubePod)
	eh.app.stateManager.SetSelectedPod(&v1Pod)
	
	// Update container dropdown
	eh.updateContainerDropdown(containerDropdown, v1Pod)
	
	// Reset port related UI
	portDropdown.SetOptions([]string{MsgSelectContainerFirst}, nil)
	portField.SetText("")
	
	
	// Set Pod IP address to host field
	if v1Pod.Status.PodIP != "" {
		hostField.SetText(v1Pod.Status.PodIP)
	} else {
		hostField.SetText("")
	}
	
	// Get auth info directly from Pod environment variables and auto-fill
	username, password := eh.getPodEnvCredentials(&v1Pod)
	if username != "" {
		usernameField.SetText(username)
	}
	if password != "" {
		passwordField.SetText(password)
	}
	
	// Get Secret list used by current Pod (preserve original functionality)
	podSecrets := eh.getPodSecrets(&v1Pod)
	if len(podSecrets) == 0 {
		secretDropdown.SetOptions([]string{"Current Pod does not use any Secret"}, nil)
	} else {
		// Update Secret dropdown, only show Secrets used by current Pod
		secretNames := make([]string, len(podSecrets))
		for i, secret := range podSecrets {
			secretNames[i] = secret
		}
		secretDropdown.SetOptions(secretNames, nil)
		
		// Rebind Secret dropdown event handler
		secretDropdown.SetSelectedFunc(func(text string, index int) {
			eh.HandleSecretSelection(text, index, secretKeyDropdown, secretDropdown, usernameField, passwordField)
		})
	}
	
	// Only reset Secret Key dropdown, not Secret dropdown (as Pod's Secret list is already set)
	secretKeyDropdown.SetOptions([]string{"Please select Secret first"}, nil)
	
	// If containers exist, select the first one by default
	if len(v1Pod.Spec.Containers) > 0 {
		containerDropdown.SetCurrentOption(0)
		eh.HandleContainerSelection(0, portDropdown, hostField, portField)
	}
}

// HandleContainerSelection handles container selection event
func (eh *EventHandlers) HandleContainerSelection(index int, portDropdown *tview.DropDown, hostField, portField *tview.InputField) {

	
	// Reset port information
	portDropdown.SetOptions([]string{MsgLoadingPorts}, nil)
	portField.SetText("")
	hostField.SetText("")
	
	// Validate Pod selection
	selectedPod := eh.app.stateManager.GetSelectedPod()
	if selectedPod == nil {
		portDropdown.SetOptions([]string{MsgSelectPodFirst}, nil)
		return
	}
	
	// Validate container index
	if index < 0 || index >= len(selectedPod.Spec.Containers) {

		portDropdown.SetOptions([]string{MsgInvalidContainerSelection}, nil)
		return
	}
	
	// Get selected container
	selectedContainer := &selectedPod.Spec.Containers[index]
	eh.app.stateManager.SetSelectedContainer(selectedContainer)
	
	// Set Pod IP address to host field
	if selectedPod.Status.PodIP != "" {
		hostField.SetText(selectedPod.Status.PodIP)
	}
	
	// Update port information
	eh.updatePortInfoForContainer(selectedContainer, portDropdown, portField)
}

// HandlePortSelection handles port selection event
func (eh *EventHandlers) HandlePortSelection(text string, index int, containerDropdown *tview.DropDown, portField *tview.InputField) {
	// Validate selection state
	if err := eh.errorHandler.ValidateSelection(eh.app); err != nil {
		return
	}
	
	// Get selected container
	selectedContainer := eh.app.stateManager.GetSelectedContainer()
	if selectedContainer == nil {
		return
	}
	
	// Validate port index
	if index < 0 || index >= len(selectedContainer.Ports) {
		return
	}
	
	// Set port field
	portField.SetText(strconv.Itoa(int(selectedContainer.Ports[index].ContainerPort)))
}

// Helper methods

// setupPodDropdownHandler sets up Pod dropdown event handler
func (eh *EventHandlers) setupPodDropdownHandler(podDropdown, containerDropdown, portDropdown *tview.DropDown, hostField, portField *tview.InputField, namespaceDropdown *tview.DropDown, secretDropdown, secretKeyDropdown *tview.DropDown, usernameField, passwordField *tview.InputField) {
	podDropdown.SetSelectedFunc(func(text string, index int) {
		eh.HandlePodSelection(index, containerDropdown, portDropdown, hostField, portField, secretDropdown, secretKeyDropdown, usernameField, passwordField)
	})
}

// resetDropdowns resets all dropdowns
func (eh *EventHandlers) resetDropdowns(podDropdown, containerDropdown, portDropdown *tview.DropDown) {
	containerDropdown.SetOptions([]string{MsgSelectPodFirst}, nil)
	portDropdown.SetOptions([]string{MsgSelectContainerFirst}, nil)
}

// convertK8sPodToV1Pod converts k8s.Pod to v1.Pod
func (eh *EventHandlers) convertK8sPodToV1Pod(kubePod k8s.Pod) v1.Pod {
	v1Pod := v1.Pod{}
	v1Pod.Name = kubePod.Name
	v1Pod.Namespace = kubePod.Namespace
	v1Pod.Status.Phase = v1.PodPhase(kubePod.Status)
	v1Pod.Status.PodIP = kubePod.PodIP
	
	// Convert container info
	v1Pod.Spec.Containers = make([]v1.Container, len(kubePod.Containers))
	for i, container := range kubePod.Containers {
		v1Container := v1.Container{
			Name:  container.Name,
			Image: container.Image,
		}
		
		// Convert port info
		v1Container.Ports = make([]v1.ContainerPort, len(container.Ports))
		for j, port := range container.Ports {
			v1Container.Ports[j] = v1.ContainerPort{
				Name:          port.Name,
				ContainerPort: port.ContainerPort,
				Protocol:      v1.Protocol(port.Protocol),
			}
		}
		
		v1Pod.Spec.Containers[i] = v1Container
	}
	
	return v1Pod
}

// HandleSecretSelection handles Secret selection event
func (eh *EventHandlers) HandleSecretSelection(secretName string, index int, secretKeyDropdown *tview.DropDown, secretDropdown *tview.DropDown, usernameField, passwordField *tview.InputField) {

	
	// Reset Secret Key dropdown
	// secretKeyDropdown.SetOptions([]string{"Loading Secret Keys..."}, nil)
	
	// Get current namespace
	namespace := eh.app.stateManager.GetCurrentNamespace()
	if namespace == "" {

		secretKeyDropdown.SetOptions([]string{"Please select namespace first"}, nil)
		return
	}
	
	// Get specified Secret directly from K8s API
	secrets, err := eh.app.k8sClient.GetSecrets(namespace)
	if err != nil {
		eh.errorHandler.HandleError(err, "Get Secret details")
			secretKeyDropdown.SetOptions([]string{fmt.Sprintf("Failed to get Secret: %v", err)}, nil)
		return
	}
	
	// Find Secret with specified name
	var selectedSecret *k8s.Secret
	for i := range secrets {
		if secrets[i].Name == secretName {
			selectedSecret = &secrets[i]
			break
		}
	}
	
	if selectedSecret == nil {

		secretKeyDropdown.SetOptions([]string{"Specified Secret not found"}, nil)
		return
	}
	
	// Save selected Secret
	eh.app.stateManager.SetSelectedSecret(selectedSecret)


	// Update Secret Key dropdown
	if len(selectedSecret.Data) == 0 {

		secretKeyDropdown.SetOptions([]string{"No data in this Secret"}, nil)
		return
	}

	keyNames := make([]string, 0, len(selectedSecret.Data))
	for key := range selectedSecret.Data {
		keyNames = append(keyNames, key)

	}


	secretKeyDropdown.SetOptions(keyNames, nil)
	
	// If there are keys available, set the first one as default and trigger selection
	if len(keyNames) > 0 {
		secretKeyDropdown.SetCurrentOption(0)
		// Manually trigger the selection for the first key to ensure it's processed
		eh.HandleSecretKeySelection(keyNames[0], 0, secretDropdown, usernameField, passwordField)
	}
	
	// Rebind Secret Key selection event handler
	secretKeyDropdown.SetSelectedFunc(func(text string, index int) {
		eh.HandleSecretKeySelection(text, index, secretDropdown, usernameField, passwordField)
	})
}

// HandleSecretKeySelection handles Secret Key selection event
func (eh *EventHandlers) HandleSecretKeySelection(keyName string, index int, secretDropdown *tview.DropDown, usernameField, passwordField *tview.InputField) {
	// Get selected Secret
	selectedSecret := eh.app.stateManager.GetSelectedSecret()
	if selectedSecret == nil {

		return
	}
	
	// Get Secret Key value
	value, exists := selectedSecret.Data[keyName]
	if !exists {
	
		return
	}
	

	// Intelligently fill fields based on Key name
	if eh.isUsernameKey(keyName) {
		usernameField.SetText(value)
	} else if eh.isPasswordKey(keyName) {
		passwordField.SetText(value)
	}
	// Remove default fill logic to avoid overwriting user manual input
}

// resetSecretDropdowns resets Secret related dropdowns
func (eh *EventHandlers) resetSecretDropdowns(secretDropdown, secretKeyDropdown *tview.DropDown) {
	secretDropdown.SetOptions([]string{"Please select namespace first"}, nil)
		secretKeyDropdown.SetOptions([]string{"Please select Secret first"}, nil)
}

// resetAuthFields resets authentication fields
func (eh *EventHandlers) resetAuthFields(usernameField, passwordField *tview.InputField) {
	usernameField.SetText("")
	passwordField.SetText("")
}

// isUsernameKey determines if it's a username related Key
func (eh *EventHandlers) isUsernameKey(keyName string) bool {
	usernameKeys := []string{"username", "user", "POSTGRES_USER", "DB_USER", "DATABASE_USER"}
	for _, key := range usernameKeys {
		if keyName == key {
			return true
		}
	}
	return false
}

// getPodSecrets gets all Secret names referenced in Pod
func (eh *EventHandlers) getPodSecrets(pod *v1.Pod) []string {


	// Get complete Pod info directly from Kubernetes API
	namespace := eh.app.stateManager.GetCurrentNamespace()
	if namespace == "" {

		return []string{}
	}

	// Get complete Pod info from K8s API
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	fullPod, err := eh.app.k8sClient.GetClientset().CoreV1().Pods(namespace).Get(ctx, pod.Name, metav1.GetOptions{})
	if err != nil {

		return []string{}
	}

	secretNames := make(map[string]bool)
	
	// Check Secrets referenced by Volumes in Pod
	for _, volume := range fullPod.Spec.Volumes {
		if volume.Secret != nil {
			secretNames[volume.Secret.SecretName] = true

		}
	}
	
	// Check Secrets referenced by environment variables in containers
	for _, container := range fullPod.Spec.Containers {

		for _, env := range container.Env {
			if env.ValueFrom != nil && env.ValueFrom.SecretKeyRef != nil {
				secretNames[env.ValueFrom.SecretKeyRef.Name] = true

			}
		}
		
		// Check Secrets referenced by EnvFrom in containers
		for _, envFrom := range container.EnvFrom {
			if envFrom.SecretRef != nil {
				secretNames[envFrom.SecretRef.Name] = true

			}
		}
	}
	
	// Check Secrets referenced by environment variables in InitContainers
	for _, container := range fullPod.Spec.InitContainers {
		for _, env := range container.Env {
			if env.ValueFrom != nil && env.ValueFrom.SecretKeyRef != nil {
				secretNames[env.ValueFrom.SecretKeyRef.Name] = true

			}
		}
		
		// Check Secrets referenced by EnvFrom in InitContainers
		for _, envFrom := range container.EnvFrom {
			if envFrom.SecretRef != nil {
				secretNames[envFrom.SecretRef.Name] = true

			}
		}
	}
	
	// Check ImagePullSecrets
	for _, imagePullSecret := range fullPod.Spec.ImagePullSecrets {
		secretNames[imagePullSecret.Name] = true

	}
	
	// Convert to slice
	result := make([]string, 0, len(secretNames))
	for name := range secretNames {
		result = append(result, name)
	}
	

	return result
}

// getPodEnvCredentials gets authentication info directly from Pod environment variables
func (eh *EventHandlers) getPodEnvCredentials(pod *v1.Pod) (username, password string) {
	// Check environment variables of all containers
	for _, container := range pod.Spec.Containers {
		for _, env := range container.Env {
			// Get directly from environment variable value
			if env.Value != "" {
				if eh.isUsernameKey(env.Name) {
					username = env.Value
				} else if eh.isPasswordKey(env.Name) {
					password = env.Value
				}
			}
			
			// Get from Secret reference
			if env.ValueFrom != nil && env.ValueFrom.SecretKeyRef != nil {
				secretName := env.ValueFrom.SecretKeyRef.Name
				secretKey := env.ValueFrom.SecretKeyRef.Key
				
				// Get Secret value
				namespace := eh.app.stateManager.GetCurrentNamespace()
				if namespace != "" {
					secretValue := eh.getSecretValue(namespace, secretName, secretKey)
					if secretValue != "" {
						if eh.isUsernameKey(env.Name) || eh.isUsernameKey(secretKey) {
							username = secretValue
						} else if eh.isPasswordKey(env.Name) || eh.isPasswordKey(secretKey) {
							password = secretValue
						}
					}
				}
			}
		}
		
		// Check Secrets referenced by EnvFrom
		for _, envFrom := range container.EnvFrom {
			if envFrom.SecretRef != nil {
				secretName := envFrom.SecretRef.Name
				namespace := eh.app.stateManager.GetCurrentNamespace()
				if namespace != "" {
					// Get all key-value pairs of the entire Secret
					secretData := eh.getSecretData(namespace, secretName)
					for key, value := range secretData {
						if eh.isUsernameKey(key) && username == "" {
							username = value
						} else if eh.isPasswordKey(key) && password == "" {
							password = value
						}
					}
				}
			}
		}
	}
	
	return username, password
}

// getSecretValue gets the value of specified Key from specified Secret
func (eh *EventHandlers) getSecretValue(namespace, secretName, secretKey string) string {
	secrets, err := eh.app.k8sClient.GetSecrets(namespace)
	if err != nil {
		return ""
	}
	
	for _, secret := range secrets {
		if secret.Name == secretName {
			if value, exists := secret.Data[secretKey]; exists {
				return value
			}
			break
		}
	}
	
	return ""
}

// getSecretData gets all data from specified Secret
func (eh *EventHandlers) getSecretData(namespace, secretName string) map[string]string {
	secrets, err := eh.app.k8sClient.GetSecrets(namespace)
	if err != nil {
		return nil
	}
	
	for _, secret := range secrets {
		if secret.Name == secretName {
			return secret.Data
		}
	}
	
	return nil
}

// isPasswordKey determines if it's a password related Key
func (eh *EventHandlers) isPasswordKey(keyName string) bool {
	passwordKeys := []string{"password", "passwd", "POSTGRES_PASSWORD", "DB_PASSWORD", "DATABASE_PASSWORD"}
	for _, key := range passwordKeys {
		if keyName == key {
			return true
		}
	}
	return false
}

// resetFields resets input fields
func (eh *EventHandlers) resetFields(hostField, portField *tview.InputField) {
	hostField.SetText("")
	portField.SetText("")
}

// updateContainerDropdown updates container dropdown
func (eh *EventHandlers) updateContainerDropdown(containerDropdown *tview.DropDown, pod v1.Pod) {
	containerNames := make([]string, len(pod.Spec.Containers))
	for i, c := range pod.Spec.Containers {
		containerNames[i] = c.Name
	}
	containerDropdown.SetOptions(containerNames, nil)
}

// updatePortInfoForContainer updates port info for specified container
func (eh *EventHandlers) updatePortInfoForContainer(container *v1.Container, portDropdown *tview.DropDown, portField *tview.InputField) {
	if len(container.Ports) == 0 {
		portDropdown.SetOptions([]string{MsgNoExposedPorts}, nil)
		portField.SetText("")
		return
	}
	
	portOptions := make([]string, len(container.Ports))
	for i, port := range container.Ports {
		if port.Name != "" {
			portOptions[i] = fmt.Sprintf("%s: %d/%s", port.Name, port.ContainerPort, port.Protocol)
		} else {
			portOptions[i] = fmt.Sprintf("%d/%s", port.ContainerPort, port.Protocol)
		}
	}
	
	portDropdown.SetOptions(portOptions, nil)
	
	// Rebind port selection event handler
	portDropdown.SetSelectedFunc(func(text string, index int) {
		eh.HandlePortSelection(text, index, nil, portField)
	})
	
	if len(portOptions) > 0 {
		portDropdown.SetCurrentOption(0)
		portField.SetText(strconv.Itoa(int(container.Ports[0].ContainerPort)))
	}
}

// findContainerByName finds container by name
func (eh *EventHandlers) findContainerByName(name string) *v1.Container {
	selectedPod := eh.app.stateManager.GetSelectedPod()
	if selectedPod == nil {
		return nil
	}
	
	for i := range selectedPod.Spec.Containers {
		if selectedPod.Spec.Containers[i].Name == name {
			return &selectedPod.Spec.Containers[i]
		}
	}
	return nil
}