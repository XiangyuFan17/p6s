package app

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// showK8sSecretSelector shows K8s Secret selector
func (a *App) showK8sSecretSelector(passwordField *tview.InputField) {
	// Get all namespaces
	namespaces, err := a.k8sClient.GetNamespaces()
	if err != nil {
		a.ShowError(fmt.Sprintf("Failed to get namespaces: %v", err))
		return
	}

	if len(namespaces) == 0 {
		a.ShowError("No namespaces found")
		return
	}

	// Create namespace selection form
	form := tview.NewForm()

	// Create namespace dropdown
	nsDropdown := tview.NewDropDown().SetLabel("Select Namespace: ")
	nsDropdown.SetFieldTextColor(tcell.ColorWhite)
	nsDropdown.SetFieldBackgroundColor(tcell.ColorBlack)
	
	// Set styles for unselected and selected items in dropdown list
	unselectedStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)
	selectedStyle := tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorWhite)
	nsDropdown.SetListStyles(unselectedStyle, selectedStyle)
	
	// Add namespace options
	nsDropdown.SetOptions(namespaces, nil)
	
	// If there are namespaces, set the first one as default selection
	if len(namespaces) > 0 {
		nsDropdown.SetCurrentOption(0)
	}

	// Add dropdown to form
	form.AddFormItem(nsDropdown)

	// Add confirm button
	form.AddButton("Next", func() {
		// Get selected namespace
		nsIndex, nsText := nsDropdown.GetCurrentOption()
		if nsIndex < 0 {
			a.ShowError("Please select a namespace")
			return
		}

		// Remove namespace selection page
		a.ui.Pages.RemovePage("k8s_ns_select")

		// Show Secret selection page
		a.showK8sSecretList(nsText, passwordField)
	})

	// Add cancel button
	form.AddButton("Cancel", func() {
		// Remove page
		a.ui.Pages.RemovePage("k8s_ns_select")
	})

	// Set form style
	form.SetBorder(true).SetTitle("Select Kubernetes Namespace").SetTitleAlign(tview.AlignCenter)
	form.SetTitleColor(tcell.ColorWhite)
	form.SetBorderColor(tcell.ColorWhite)

	// Create modal dialog container
	modal := tview.NewFlex().SetDirection(tview.FlexRow)
	modal.AddItem(nil, 0, 1, false) // Top margin

	// Create horizontal layout container
	modalRow := tview.NewFlex().SetDirection(tview.FlexColumn)
	modalRow.AddItem(nil, 0, 1, false) // Left margin
	modalRow.AddItem(form, 50, 1, true) // Form
	modalRow.AddItem(nil, 0, 1, false) // Right margin

	modal.AddItem(modalRow, 10, 1, true) // Form row
	modal.AddItem(nil, 0, 1, false) // Bottom margin

	// Create a centered container
	center := tview.NewFlex().SetDirection(tview.FlexRow)
	center.AddItem(nil, 0, 1, false)
	center.AddItem(modal, 10, 1, true)
	center.AddItem(nil, 0, 1, false)

	// Add modal dialog
	a.ui.Pages.AddPage("k8s_ns_select", center, true, true)

	// Add input capture for namespace dropdown, handle Enter and arrow keys
	var nsDropdownOpen bool = false
	nsDropdown.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			if !nsDropdownOpen {
				// If dropdown is not open, open it (simulate space key press)
				nsDropdownOpen = true
				return tcell.NewEventKey(tcell.KeyRune, ' ', tcell.ModNone)
			} else {
				// If dropdown is open, close it first then select current option
				nsDropdownOpen = false
				// First simulate space key to close dropdown
				a.ui.App.QueueEvent(tcell.NewEventKey(tcell.KeyRune, ' ', tcell.ModNone))
				// Then return Enter key event to select option
				return event
			}
		case tcell.KeyEscape:
			// If ESC key is pressed, close dropdown
			nsDropdownOpen = false
			return event
		case tcell.KeyUp, tcell.KeyDown:
			if !nsDropdownOpen {
				// If dropdown is not open, allow arrow keys to move focus
				if event.Key() == tcell.KeyUp {
					// If it's up key, move to previous form item
					return tcell.NewEventKey(tcell.KeyBacktab, 0, tcell.ModNone)
				} else {
					// If it's down key, move to next form item
					return tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone)
				}
			} else {
				// If dropdown is open, handle arrow keys normally
				return event
			}
		case tcell.KeyRune:
			if event.Rune() == ' ' {
				// Space key toggles dropdown state
				nsDropdownOpen = !nsDropdownOpen
			}
			return event
		}
		return event
	})

	// Set focus
	a.ui.App.SetFocus(nsDropdown)
}

// showK8sSecretList shows Secret list for specified namespace
func (a *App) showK8sSecretList(namespace string, passwordField *tview.InputField) {
	// Get all Secrets in specified namespace
	secrets, err := a.k8sClient.GetSecrets(namespace)
	if err != nil {
		a.ShowError(fmt.Sprintf("Failed to get Secret list: %v", err))
		return
	}

	if len(secrets) == 0 {
		a.ShowError(fmt.Sprintf("No Secrets found in namespace %s", namespace))
		return
	}

	// Create Secret name list
	secretNames := make([]string, len(secrets))
	for i, s := range secrets {
		secretNames[i] = s.Name
	}

	// Create Secret selection form
	form := tview.NewForm()

	// Create Secret dropdown
	secretDropdown := tview.NewDropDown().SetLabel("Select Secret: ")
	secretDropdown.SetFieldTextColor(tcell.ColorWhite)
	secretDropdown.SetFieldBackgroundColor(tcell.ColorBlack)
	
	// Set styles for unselected and selected items in dropdown list
	unselectedStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)
	selectedStyle := tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorWhite)
	secretDropdown.SetListStyles(unselectedStyle, selectedStyle)
	
	// Add Secret options
	secretDropdown.SetOptions(secretNames, nil)
	
	// If there are Secrets, set the first one as default selection
	if len(secretNames) > 0 {
		secretDropdown.SetCurrentOption(0)
	}

	// Add dropdown to form
	form.AddFormItem(secretDropdown)

	// Create key input field
	keyField := tview.NewInputField().SetLabel("Key in Secret: ")
	keyField.SetFieldTextColor(tcell.ColorWhite)
	keyField.SetFieldBackgroundColor(tcell.ColorBlack)
	
	// Add key input field to form
	form.AddFormItem(keyField)

	// Add confirm button
	form.AddButton("Get Password", func() {
		// Get selected Secret
		secretIndex, secretName := secretDropdown.GetCurrentOption()
		if secretIndex < 0 {
			a.ShowError("Please select a Secret")
			return
		}

		// Get input key
		key := keyField.GetText()
		if key == "" {
			a.ShowError("Please enter key in Secret")
			return
		}

		// Get Secret
		secret, err := a.k8sClient.GetSecret(namespace, secretName)
		if err != nil {
			a.ShowError(fmt.Sprintf("Failed to get Secret: %v", err))
			return
		}

		// Get password
		password, ok := secret.Data[key]
		if !ok {
			a.ShowError(fmt.Sprintf("Key %s does not exist in Secret %s", key, secretName))
			return
		}

		// Set password
		passwordField.SetText(password)

		// Show success message
		a.ShowInfo(fmt.Sprintf("Password retrieved from Secret %s", secretName))

		// Remove page
		a.ui.Pages.RemovePage("k8s_secret_select")
	})

	// Add cancel button
	form.AddButton("Cancel", func() {
		// Remove page
		a.ui.Pages.RemovePage("k8s_secret_select")
	})

	// Set form style
	form.SetBorder(true).SetTitle(fmt.Sprintf("Select Secret in namespace %s", namespace)).SetTitleAlign(tview.AlignCenter)
	form.SetTitleColor(tcell.ColorWhite)
	form.SetBorderColor(tcell.ColorWhite)

	// Create modal dialog container
	modal := tview.NewFlex().SetDirection(tview.FlexRow)
	modal.AddItem(nil, 0, 1, false) // Top margin

	// Create horizontal layout container
	modalRow := tview.NewFlex().SetDirection(tview.FlexColumn)
	modalRow.AddItem(nil, 0, 1, false) // Left margin
	modalRow.AddItem(form, 50, 1, true) // Form
	modalRow.AddItem(nil, 0, 1, false) // Right margin

	modal.AddItem(modalRow, 12, 1, true) // Form row
	modal.AddItem(nil, 0, 1, false) // Bottom margin

	// Create a centered container
	center := tview.NewFlex().SetDirection(tview.FlexRow)
	center.AddItem(nil, 0, 1, false)
	center.AddItem(modal, 12, 1, true)
	center.AddItem(nil, 0, 1, false)

	// Add modal dialog
	a.ui.Pages.AddPage("k8s_secret_select", center, true, true)

	// Add input capture for Secret dropdown, handle Enter and arrow keys
	var secretDropdownOpen bool = false
	secretDropdown.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			if !secretDropdownOpen {
				// If dropdown is not open, open it (simulate space key press)
				secretDropdownOpen = true
				return tcell.NewEventKey(tcell.KeyRune, ' ', tcell.ModNone)
			} else {
				// If dropdown is open, close it first then select current option
				secretDropdownOpen = false
				// First simulate space key to close dropdown
				a.ui.App.QueueEvent(tcell.NewEventKey(tcell.KeyRune, ' ', tcell.ModNone))
				// Then return Enter key event to select option
				return event
			}
		case tcell.KeyEscape:
			// If ESC key is pressed, close dropdown
			secretDropdownOpen = false
			return event
		case tcell.KeyUp, tcell.KeyDown:
			if !secretDropdownOpen {
				// If dropdown is not open, allow arrow keys to move focus
				if event.Key() == tcell.KeyUp {
					// If it's up key, move to previous form item
					return tcell.NewEventKey(tcell.KeyBacktab, 0, tcell.ModNone)
				} else {
					// If it's down key, move to next form item
					return tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone)
				}
			} else {
				// If dropdown is open, handle arrow keys normally
				return event
			}
		case tcell.KeyRune:
			if event.Rune() == ' ' {
				// Space key toggles dropdown state
				secretDropdownOpen = !secretDropdownOpen
			}
			return event
		}
		return event
	})

	// Set focus
	a.ui.App.SetFocus(secretDropdown)
}