package app

import (
	"fmt"
	"strings"
)

// handleK8sCommand handles Kubernetes related commands
func (a *App) handleK8sCommand(cmd string) {
	// Check if connected to Kubernetes
	if !a.k8sConnected {
		a.ShowError("Not connected to Kubernetes cluster, please ensure your kubeconfig is configured correctly")
		return
	}

	// Parse command
	cmdParts := strings.Fields(cmd)
	if len(cmdParts) == 1 {
		// Only \k8s, show help info
		a.showK8sHelp()
		return
	}

	// Handle subcommands
	switch cmdParts[1] {
	case "ns", "namespace":
		// List all namespaces
		a.listNamespaces()
	
	case "secrets":
		// Get secrets in specified namespace
		if len(cmdParts) < 3 {
			a.ShowError("Please specify namespace, e.g.: \\k8s secrets default")
			return
		}
		namespace := cmdParts[2]
		a.listSecrets(namespace)
	
	case "secret":
		// Get specific secret in specified namespace
		if len(cmdParts) < 4 {
			a.ShowError("Please specify namespace and Secret name, e.g.: \\k8s secret default my-secret")
			return
		}
		namespace := cmdParts[2]
		secretName := cmdParts[3]
		a.showSecret(namespace, secretName)
	
	case "context":
		// Show current context
		context := a.k8sClient.GetCurrentContext()
		a.ShowInfo(fmt.Sprintf("Current Kubernetes context: %s", context))
	
	default:
		a.ShowError(fmt.Sprintf("Unknown K8s command: %s", cmdParts[1]))
		a.showK8sHelp()
	}
}

// showK8sHelp shows Kubernetes command help information
func (a *App) showK8sHelp() {
	helpText := `Kubernetes command help:
\k8s ns/namespace - List all namespaces
\k8s secrets <namespace> - List all Secrets in specified namespace
\k8s secret <namespace> <name> - Show specific Secret in specified namespace
\k8s context - Show current Kubernetes context`

	a.ShowInfo(helpText)
}

// listNamespaces lists all namespaces
func (a *App) listNamespaces() {
	namespaces, err := a.k8sClient.GetNamespaces()
	if err != nil {
		a.ShowError(fmt.Sprintf("Failed to get namespaces: %v", err))
		return
	}

	// Clear table
	a.ClearTable()

	// Set table headers
	a.SetTableHeaders([]string{"Namespace"})

	// Add data rows
	for _, ns := range namespaces {
		a.AddTableRow([]string{ns})
	}

	// Show result info
	a.ShowInfo(fmt.Sprintf("Found %d namespaces", len(namespaces)))
}

// listSecrets lists all Secrets in specified namespace
func (a *App) listSecrets(namespace string) {
	secrets, err := a.k8sClient.GetSecrets(namespace)
	if err != nil {
		a.ShowError(fmt.Sprintf("Failed to get Secret list: %v", err))
		return
	}

	// Clear table
	a.ClearTable()

	// Set table headers
	a.SetTableHeaders([]string{"Name", "Type", "Creation Time", "Data Key Count"})

	// Add data rows
	for _, s := range secrets {
		a.AddTableRow([]string{
			s.Name,
			s.Type,
			s.CreateTime.Format("2006-01-02 15:04:05"),
			fmt.Sprintf("%d", len(s.Data)),
		})
	}

	// Show result info
	a.ShowInfo(fmt.Sprintf("Found %d Secrets in namespace %s", len(secrets), namespace))
}

// showSecret shows specific Secret in specified namespace
func (a *App) showSecret(namespace, name string) {
	secret, err := a.k8sClient.GetSecret(namespace, name)
	if err != nil {
		a.ShowError(fmt.Sprintf("Failed to get Secret: %v", err))
		return
	}

	// Clear table
	a.ClearTable()

	// Set table headers
	a.SetTableHeaders([]string{"Key", "Value"})

	// Add data rows
	for key, value := range secret.Data {
		a.AddTableRow([]string{key, value})
	}

	// Show result info
	a.ShowInfo(fmt.Sprintf("Secret: %s (Type: %s, Namespace: %s)", 
		secret.Name, secret.Type, secret.Namespace))
}