package app

import (
	"fmt"
	"time"
	"p6s/internal/config"
	"github.com/rivo/tview"
)

// showK8sConfigForm shows form for configuring database connection from K8s Pod
func (a *App) showK8sConfigForm() {
	// Create UI factory and event handlers
	uiFactory := NewUIFactory()
	eventHandlers := NewEventHandlers(a)
	configManager := NewConfigManager(a)
	
	// Try to load saved configuration
	var savedConfig *config.Config
	if cfg, err := config.LoadConfig(); err == nil {
		savedConfig = cfg
	}
	
	// Create configuration form
	form := uiFactory.CreateForm("Configure PostgreSQL Connection from Kubernetes Pod")

	// First create K8s related fields
	// Create namespace selection dropdown
	namespaceDropdown := uiFactory.CreateDropDown("Namespace: ")
	var namespaceDropdownOpen bool = false
	uiFactory.SetupDropdownInputCapture(namespaceDropdown, &namespaceDropdownOpen)
	form.AddFormItem(namespaceDropdown)
	
	// Create Pod selection dropdown
	podDropdown := uiFactory.CreateDropDown("Pod: ")
	var podDropdownOpen bool = false
	uiFactory.SetupDropdownInputCapture(podDropdown, &podDropdownOpen)
	form.AddFormItem(podDropdown)

	// Create container selection dropdown
	containerDropdown := uiFactory.CreateDropDown("Container: ")
	var containerDropdownOpen bool = false
	uiFactory.SetupDropdownInputCapture(containerDropdown, &containerDropdownOpen)
	form.AddFormItem(containerDropdown)

	// Create port selection dropdown
	portDropdown := uiFactory.CreateDropDown("Port: ")
	var portDropdownOpen bool = false
	uiFactory.SetupDropdownInputCapture(portDropdown, &portDropdownOpen)
	form.AddFormItem(portDropdown)

	// Create Secret selection dropdown
	secretDropdown := uiFactory.CreateDropDown("Secret: ")
	var secretDropdownOpen bool = false
	uiFactory.SetupDropdownInputCapture(secretDropdown, &secretDropdownOpen)
	form.AddFormItem(secretDropdown)

	// Create Secret Key selection dropdown
	secretKeyDropdown := uiFactory.CreateDropDown("Secret Key: ")
	var secretKeyDropdownOpen bool = false
	uiFactory.SetupDropdownInputCapture(secretKeyDropdown, &secretKeyDropdownOpen)
	form.AddFormItem(secretKeyDropdown)

	// Then create database connection fields, place at bottom
	// Create read-only connection info fields, use saved values if config exists
	hostValue := ""
	if savedConfig != nil {
		hostValue = savedConfig.Host
	}
	hostField := uiFactory.CreateInputField("Host: ", hostValue, true)
	form.AddFormItem(hostField)
	
	portValue := ""
	if savedConfig != nil {
		portValue = savedConfig.Port
	}
	portField := uiFactory.CreateInputField("Port: ", portValue, true)
	form.AddFormItem(portField)
	
	databaseValue := "postgres"
	if savedConfig != nil && savedConfig.Database != "" {
		databaseValue = savedConfig.Database
	}
	databaseField := uiFactory.CreateInputField("Database: ", databaseValue, false)
	form.AddFormItem(databaseField)

	// Create username field, use saved value if config exists and not empty, otherwise default to postgres
	usernameValue := "postgres"
	if savedConfig != nil && savedConfig.Username != "" {
		usernameValue = savedConfig.Username
	}
	usernameField := uiFactory.CreateInputField("Username: ", usernameValue, false)
	form.AddFormItem(usernameField)

	// Create password field, use saved value if config exists
	passwordValue := ""
	if savedConfig != nil {
		passwordValue = savedConfig.Password
	}
	passwordField := uiFactory.CreateInputField("Password: ", passwordValue, false)
	passwordField.SetMaskCharacter('*')
	form.AddFormItem(passwordField)

	// // First set default empty options for all dropdowns to ensure they are always visible
	// podDropdown.SetOptions([]string{"Please select namespace first"}, nil)
	// containerDropdown.SetOptions([]string{"Please select Pod first"}, nil)
	// portDropdown.SetOptions([]string{"Please select container first"}, nil)
	
	// Get all namespaces
	namespaces, err := a.k8sClient.GetNamespaces()
	if err != nil {
		// Even if getting namespaces fails, show a default option
		namespaceDropdown.SetOptions([]string{fmt.Sprintf("Failed to get namespaces: %v", err)}, nil)
	} else if len(namespaces) == 0 {
		// If no namespaces found, show hint message
		namespaceDropdown.SetOptions([]string{"No namespaces found"}, nil)
	} else {
		// Add namespace options
		namespaceDropdown.SetOptions(namespaces, nil)
		
		// If saved config exists, try to set to saved value
		selectedNamespace := ""
		if savedConfig != nil && savedConfig.Namespace != "" {
			// Check if saved namespace exists in current list
			for i, ns := range namespaces {
				if ns == savedConfig.Namespace {
					namespaceDropdown.SetCurrentOption(i)
					selectedNamespace = ns
					break
				}
			}
		}
		
		// If saved namespace not found, select the first one
		if selectedNamespace == "" {
			namespaceDropdown.SetCurrentOption(0)
			selectedNamespace = namespaces[0]
		}
		
		// Initialize other dropdowns
		eventHandlers.HandleNamespaceSelection(selectedNamespace, namespaceDropdown, podDropdown, containerDropdown, portDropdown, secretDropdown, secretKeyDropdown, hostField, portField, usernameField, passwordField)
		
		// If saved config exists, try to restore other K8s field selections
		if savedConfig != nil {
			// Delayed execution, wait for dropdown options to load
			go func() {
				time.Sleep(500 * time.Millisecond) // Wait 500ms for options to load
				a.ui.App.QueueUpdateDraw(func() {
					a.restoreK8sSelections(savedConfig, podDropdown, containerDropdown, portDropdown, secretDropdown, secretKeyDropdown)
				})
			}()
		}
	}

	// Use StateManager to manage selected Pod state

	// Set namespace dropdown selection event
	namespaceDropdown.SetSelectedFunc(func(text string, index int) {
		eventHandlers.HandleNamespaceSelection(text, namespaceDropdown, podDropdown, containerDropdown, portDropdown, secretDropdown, secretKeyDropdown, hostField, portField, usernameField, passwordField)
	})

	// Set Pod dropdown selection event
	podDropdown.SetSelectedFunc(func(text string, index int) {
		eventHandlers.HandlePodSelection(index, containerDropdown, portDropdown, hostField, portField, secretDropdown, secretKeyDropdown, usernameField, passwordField)
	})

	// Set container dropdown selection event
	containerDropdown.SetSelectedFunc(func(text string, index int) {
		eventHandlers.HandleContainerSelection(index, portDropdown, hostField, portField)
	})

	// Set port dropdown selection event
	portDropdown.SetSelectedFunc(func(text string, index int) {
		eventHandlers.HandlePortSelection(text, index, containerDropdown, portField)
	})

	// Set Secret dropdown selection event
	secretDropdown.SetSelectedFunc(func(text string, index int) {
		eventHandlers.HandleSecretSelection(text, index, secretKeyDropdown, secretDropdown, usernameField, passwordField)
	})

	// Set Secret Key dropdown selection event
	secretKeyDropdown.SetSelectedFunc(func(text string, index int) {
		eventHandlers.HandleSecretKeySelection(text, index, secretDropdown, usernameField, passwordField)
	})

	// Add save button
	form.AddButton("    Save    ", func() {
		// Validate selection state
		if a.stateManager.GetSelectedPod() == nil {
			eventHandlers.errorHandler.HandleValidationError(ErrSelectPodFirst)
			return
		}

		// Check if container is selected
		_, containerName := containerDropdown.GetCurrentOption()
		if containerName == "" {
			eventHandlers.errorHandler.HandleValidationError(ErrSelectContainerFirst)
			return
		}

		// Check if port is selected
		_, portText := portDropdown.GetCurrentOption()
		if portText == "" {
			eventHandlers.errorHandler.HandleValidationError(ErrSelectPortFirst)
			return
		}

		// Get field values
		host := hostField.GetText()
		port := portField.GetText()
		database := databaseField.GetText()
		username := usernameField.GetText()
		password := passwordField.GetText()

		// Get K8s related field values
		namespace := ""
		pod := ""
		container := ""
		portName := ""
		secret := ""
		secretKey := ""
		
		if nsIndex, nsText := namespaceDropdown.GetCurrentOption(); nsIndex >= 0 {
			namespace = nsText
		}
		if podIndex, podText := podDropdown.GetCurrentOption(); podIndex >= 0 {
			pod = podText
		}
		if containerIndex, containerText := containerDropdown.GetCurrentOption(); containerIndex >= 0 {
			container = containerText
		}
		if portIndex, portText := portDropdown.GetCurrentOption(); portIndex >= 0 {
			portName = portText
		}
		if secretIndex, secretText := secretDropdown.GetCurrentOption(); secretIndex >= 0 {
			secret = secretText
		}
		if secretKeyIndex, secretKeyText := secretKeyDropdown.GetCurrentOption(); secretKeyIndex >= 0 {
			secretKey = secretKeyText
		}

		// Create connection configuration
		connConfig := &ConnectionConfig{
			Host:      host,
			Port:      port,
			Database:  database,
			Username:  username,
			Password:  password,
			SSLMode:   DefaultSSLMode,
			Namespace: namespace,
			Pod:       pod,
			Container: container,
			PortName:  portName,
			Secret:    secret,
			SecretKey: secretKey,
		}

		// Immediately remove page and set focus to avoid UI freeze
		a.ui.Pages.RemovePage(K8sConfigPageName)
		a.ui.App.SetFocus(a.ui.ConnTable)

		// Save configuration and connect
		configManager.SaveAndConnect(connConfig, nil)
	})

	// Add cancel button
	form.AddButton("   Cancel   ", func() {
		// Immediately remove page and set focus to avoid UI freeze
		a.ui.Pages.RemovePage(K8sConfigPageName)
		a.ui.App.SetFocus(a.ui.ConnTable)
	})

	// Get buttons and set styles
	saveButton := form.GetButton(0) // Save button
	cancelButton := form.GetButton(1) // Cancel button
	
	// Set button styles
	saveButton.SetLabelColor(ButtonTextColor)
	saveButton.SetBackgroundColor(SaveButtonColor)
	saveButton.SetLabel("[::b]    Save    [::-]")
	
	cancelButton.SetLabelColor(ButtonTextColor)
	cancelButton.SetBackgroundColor(CancelButtonColor)
	cancelButton.SetLabel("[::b]   Cancel   [::-]")

	// Set form input capture
	uiFactory.SetupFormInputCapture(form, func() {
		a.ui.Pages.RemovePage(K8sConfigPageName)
		a.ui.App.SetFocus(a.ui.ConnTable)
	})

	// Create modal dialog container
	center := uiFactory.CreateModalContainer(form)

	// Add modal dialog
	a.ui.Pages.AddPage(K8sConfigPageName, center, true, true)

	// Set focus
	a.ui.App.SetFocus(form)

	// Use goroutine to delay setting focus to namespace dropdown
	go func() {
		time.Sleep(UIUpdateDelay)

		a.ui.App.QueueUpdate(func() {
			a.ui.App.SetFocus(namespaceDropdown)
		})
	}()
}

// restoreK8sSelections restores K8s field selections
func (a *App) restoreK8sSelections(savedConfig *config.Config, podDropdown, containerDropdown, portDropdown, secretDropdown, secretKeyDropdown *tview.DropDown) {
	// Due to tview.DropDown API limitations and complex asynchronous event handling,
	// currently can only restore selection state without triggering related cascade update events
	// Users need to manually reselect to trigger complete update flow
	
	// Restore Pod selection
	if savedConfig.Pod != "" {
		pods := a.stateManager.GetCurrentPods()
		for i, pod := range pods {
			if pod.Name == savedConfig.Pod {
				podDropdown.SetCurrentOption(i)
				break
			}
		}
	}
	
	// Restore Container selection
	if savedConfig.Container != "" {
		containerCount := containerDropdown.GetOptionCount()
		for i := 0; i < containerCount; i++ {
			containerDropdown.SetCurrentOption(i)
			_, containerText := containerDropdown.GetCurrentOption()
			if containerText == savedConfig.Container {
				break
			}
		}
	}
	
	// Restore Port selection
	if savedConfig.PortName != "" {
		portCount := portDropdown.GetOptionCount()
		for i := 0; i < portCount; i++ {
			portDropdown.SetCurrentOption(i)
			_, portText := portDropdown.GetCurrentOption()
			if portText == savedConfig.PortName {
				break
			}
		}
	}
	
	// Restore Secret selection
	if savedConfig.Secret != "" {
		secretCount := secretDropdown.GetOptionCount()
		for i := 0; i < secretCount; i++ {
			secretDropdown.SetCurrentOption(i)
			_, secretText := secretDropdown.GetCurrentOption()
			if secretText == savedConfig.Secret {
				break
			}
		}
	}
	
	// Restore SecretKey selection
	if savedConfig.SecretKey != "" {
		secretKeyCount := secretKeyDropdown.GetOptionCount()
		for i := 0; i < secretKeyCount; i++ {
			secretKeyDropdown.SetCurrentOption(i)
			_, secretKeyText := secretKeyDropdown.GetCurrentOption()
			if secretKeyText == savedConfig.SecretKey {
				break
			}
		}
	}
	

}