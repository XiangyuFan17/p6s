package app

import (
	"fmt"
)

// ErrorHandler error handler
type ErrorHandler struct {
	app *App
}

// NewErrorHandler creates a new error handler
func NewErrorHandler(app *App) *ErrorHandler {
	return &ErrorHandler{app: app}
}

// HandleError unified error handling
func (eh *ErrorHandler) HandleError(err error, context string) {
	if err == nil {
		return
	}
	
	// Display user-friendly error message
	eh.app.ShowError(fmt.Sprintf("%s: %v", context, err))
}

// HandleValidationError handles validation errors
func (eh *ErrorHandler) HandleValidationError(message string) {
	eh.app.ShowError(message)
}







// ValidateSelection validates selection state
func (eh *ErrorHandler) ValidateSelection(app *App) error {
	if app.stateManager.GetSelectedPod() == nil {
		eh.HandleValidationError(ErrSelectPodFirst)
		return fmt.Errorf(ErrSelectPodFirst)
	}
	return nil
}

// ValidateFields validates field completeness
func (eh *ErrorHandler) ValidateFields(host, port, database string) error {
	if host == "" || port == "" {
		// eh.HandleValidationError(ErrIncompleteHostPort)
		return fmt.Errorf(ErrIncompleteHostPort)
	}
	
	if database == "" {
		eh.HandleValidationError(ErrEnterDatabaseName)
		return fmt.Errorf(ErrEnterDatabaseName)
	}
	
	return nil
}