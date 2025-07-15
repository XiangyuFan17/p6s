package main

import (
	"strings"
	"p6s/internal/app"
	"p6s/internal/config"
)

func main() {
	// Create application instance
	app := app.NewApp()

	// Default connection string used when config file doesn't exist or fails to load (read-only mode)
	defaultConnStr := "postgres://postgres:password@localhost:5432/postgres?sslmode=disable&default_transaction_read_only=on&application_name=p6s-readonly"
	
	// Try to load connection info from config file
	cfg, err := config.LoadConfig()
	if err != nil {
		
		// Parse default connection string
		connStrNoPrefix := strings.TrimPrefix(defaultConnStr, "postgres://")
		userInfoAndHost := strings.SplitN(connStrNoPrefix, "@", 2)
		userInfo := userInfoAndHost[0]
		hostAndDb := userInfoAndHost[1]
		username := strings.SplitN(userInfo, ":", 2)[0]
		password := strings.SplitN(userInfo, ":", 2)[1]
		hostPortAndDb := strings.SplitN(hostAndDb, "/", 2)
		hostPort := hostPortAndDb[0]
		dbAndParams := hostPortAndDb[1]
		host := strings.SplitN(hostPort, ":", 2)[0]
		port := strings.SplitN(hostPort, ":", 2)[1]
		dbName := strings.SplitN(dbAndParams, "?", 2)[0]
		sslmode := "disable"
		if strings.Contains(dbAndParams, "sslmode=") {
			sslmode = strings.SplitN(strings.SplitN(dbAndParams, "sslmode=", 2)[1], "&", 2)[0]
		}
		
		// Set application connection parameters
		app.SetConnectionParams(host, port, username, password, dbName, sslmode)
		
		// Try to save default config to config file
		defaultConfig := &config.Config{
			Host:     host,
			Port:     port,
			Username: username,
			Password: password,
			Database: dbName,
			SSLMode:  sslmode,
		}
		config.SaveConfig(defaultConfig)
	} else {
		// Use connection info from config file
		app.SetConnectionParams(cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.Database, cfg.SSLMode)
	}

	// Try to connect to database, but continue running even if it fails
	app.Connect()

	// Run application
	app.Run()
}