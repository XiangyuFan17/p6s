package app

import (
	"fmt"
	"p6s/internal/config"
)

// ConfigManager configuration manager
type ConfigManager struct {
	app          *App
	errorHandler *ErrorHandler
}

// NewConfigManager creates new configuration manager
func NewConfigManager(app *App) *ConfigManager {
	return &ConfigManager{
		app:          app,
		errorHandler: NewErrorHandler(app),
	}
}

// ConnectionConfig connection configuration
type ConnectionConfig struct {
	Host     string
	Port     string
	Database string
	Username string
	Password string
	SSLMode  string
	// K8s related configuration
	Namespace string
	Pod       string
	Container string
	PortName  string // Save selected port name
	Secret    string
	SecretKey string
}

// NewConnectionConfig creates new connection configuration
func NewConnectionConfig(host, port, database string) *ConnectionConfig {
	return &ConnectionConfig{
		Host:     host,
		Port:     port,
		Database: database,
		Username: DefaultUsername,
		Password: DefaultPassword,
		SSLMode:  DefaultSSLMode,
	}
}

// NewConnectionConfigWithAuth creates connection configuration with auth info
func NewConnectionConfigWithAuth(host, port, database, username, password string) *ConnectionConfig {
	return &ConnectionConfig{
		Host:     host,
		Port:     port,
		Database: database,
		Username: username,
		Password: password,
		SSLMode:  DefaultSSLMode,
	}
}

// Validate validates configuration
func (cc *ConnectionConfig) Validate() error {
	if cc.Host == "" || cc.Port == "" {
		return fmt.Errorf(ErrIncompleteHostPort)
	}
	
	if cc.Database == "" {
		return fmt.Errorf(ErrEnterDatabaseName)
	}
	
	return nil
}

// SaveAndConnect saves configuration and connects to database
func (cm *ConfigManager) SaveAndConnect(connConfig *ConnectionConfig, onComplete func(error)) {
	// Validate configuration
	if err := connConfig.Validate(); err != nil {
		cm.errorHandler.HandleValidationError(err.Error())
		return
	}
	
	// Use goroutine to handle time-consuming operations
	go func() {
		// If K8s Secret is used, need to get actual password value
		actualPassword := connConfig.Password
		if connConfig.Secret != "" && connConfig.SecretKey != "" && connConfig.Namespace != "" {
			// Get actual password from K8s Secret
			if secretPassword := cm.getSecretPassword(connConfig.Namespace, connConfig.Secret, connConfig.SecretKey); secretPassword != "" {
				actualPassword = secretPassword
			}
		}

		// Update app connection parameters (using actual password)
		cm.updateAppConfigWithPassword(connConfig, actualPassword)
		
		// Create configuration object
		cfg := &config.Config{
			Host:     connConfig.Host,
			Port:     connConfig.Port,
			Username: connConfig.Username,
			Password: connConfig.Password, // Save original password or form input password
			Database: connConfig.Database,
			SSLMode:  connConfig.SSLMode,
			// K8s related configuration
			Namespace: connConfig.Namespace,
			Pod:       connConfig.Pod,
			Container: connConfig.Container,
			PortName:  connConfig.PortName,
			Secret:    connConfig.Secret,
			SecretKey: connConfig.SecretKey,
		}
		
		var finalError error
		
		// Save configuration to file
		if err := config.SaveConfig(cfg); err != nil {
			cm.errorHandler.HandleError(err, "Save configuration")
			finalError = err
		} else {
			// Try to connect to database
			if err := cm.app.Connect(); err != nil {
				cm.errorHandler.HandleError(err, "Connect to database")
				finalError = err
			} else {
				// Refresh data
				if err := cm.app.refreshData(); err != nil {
					cm.errorHandler.HandleError(err, "Refresh data")
					finalError = err
				} else {
			
				}
			}
		}
		
		// Use QueueUpdateDraw to update UI
		cm.app.ui.App.QueueUpdateDraw(func() {
			if onComplete != nil {
				onComplete(finalError)
			}
			
			// Show result message
			cm.showResultMessage(finalError)
		})
	}()
}

// updateAppConfig updates app configuration
func (cm *ConfigManager) updateAppConfig(connConfig *ConnectionConfig) {
	cm.updateAppConfigWithPassword(connConfig, connConfig.Password)
}

// updateAppConfigWithPassword updates app configuration with specified password
func (cm *ConfigManager) updateAppConfigWithPassword(connConfig *ConnectionConfig, password string) {
	cm.app.host = connConfig.Host
	cm.app.port = connConfig.Port
	cm.app.username = connConfig.Username
	cm.app.password = password // Use actual password
	cm.app.database = connConfig.Database
	cm.app.sslmode = connConfig.SSLMode
	cm.app.connStr = config.BuildConnStr(
		connConfig.Host,
		connConfig.Port,
		connConfig.Username,
		password, // Use actual password to build connection string
		connConfig.Database,
		connConfig.SSLMode,
	)
}

// getSecretPassword gets password from K8s Secret
func (cm *ConfigManager) getSecretPassword(namespace, secretName, secretKey string) string {
	if cm.app.k8sClient == nil || !cm.app.k8sClient.IsConnected() {
		return ""
	}
	
	secrets, err := cm.app.k8sClient.GetSecrets(namespace)
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

// showResultMessage shows result message
func (cm *ConfigManager) showResultMessage(err error) {
	// Get current connection info text
	currentText := cm.app.ui.ConnInfo.GetText(true)
	
	// Show operation result
	var resultMsg string
	if err != nil {
		resultMsg = fmt.Sprintf("\n[red]Operation failed: %v[white]", err)
	} else {
		resultMsg = fmt.Sprintf("\n[green]%s[white]", MsgConfigSaved)
	}
	
	// Append result message to existing text
	cm.app.ui.ConnInfo.SetText(currentText + resultMsg)
}