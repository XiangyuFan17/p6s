package app

import (
	"time"

	"github.com/gdamore/tcell/v2"
)

// UI related constants
const (
	// Form dimensions
	FormWidth  = 60
	FormHeight = 30
	
	// Delay times
	UIUpdateDelay = 5 * time.Millisecond
	
	// Page names
	K8sConfigPageName = "k8s_config"
	SQLQueryPageName  = "sql_query"
)

// Color constants
var (
	// Field colors
	FieldTextColor       = tcell.ColorWhite
	FieldBackgroundColor = tcell.ColorBlack
	
	// Dropdown styles
	UnselectedStyle = tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)
	SelectedStyle   = tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorWhite)
	
	// Button colors
	SaveButtonColor   = tcell.ColorBlue
	CancelButtonColor = tcell.ColorRed
	ButtonTextColor   = tcell.ColorWhite
	
	// Border colors
	BorderColor = tcell.ColorWhite
	TitleColor  = tcell.ColorWhite
)

// Default value constants
const (
	DefaultUsername = "postgres"
	DefaultPassword = ""
	DefaultSSLMode  = "disable"
)

// Error message constants
const (
	ErrSelectPodFirst       = "Please select a Pod first"
	ErrSelectContainerFirst = "Please select a container"
	ErrSelectPortFirst      = "Please select a port"
	ErrIncompleteHostPort   = "Host or port information is incomplete, please reselect Pod, container and port"
	ErrEnterDatabaseName    = "Please enter database name"
)

// Info message constants
const (
	MsgSelectNamespaceFirst      = "Please select namespace first"
	MsgSelectPodFirst            = "Please select Pod first"
	MsgSelectContainerFirst      = "Please select container first"
	MsgInvalidPodSelection       = "Invalid Pod selection"
	MsgInvalidContainerSelection = "Invalid container selection"
	MsgNoPods                    = "No Pods in namespace %s"
	MsgNoContainers              = "No containers in this Pod"
	MsgNoExposedPorts            = "No exposed ports in this container"
	MsgContainerNotFound         = "Container %s not found"
	MsgGetPodsError              = "Failed to get Pods: %v"
	MsgLoadingPorts              = "Loading ports..."
)

// Success message constants
const (
	MsgConfigSaved = "Configuration saved and successfully connected to database"
)