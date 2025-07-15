package app

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// UIFactory UI component factory
type UIFactory struct{}

// NewUIFactory creates a new UI factory
func NewUIFactory() *UIFactory {
	return &UIFactory{}
}

// CreateInputField creates a standard input field
func (uf *UIFactory) CreateInputField(label, text string, disabled bool) *tview.InputField {
	field := tview.NewInputField().SetLabel(label).SetText(text)
	field.SetFieldTextColor(FieldTextColor)
	field.SetFieldBackgroundColor(FieldBackgroundColor)
	if disabled {
		field.SetDisabled(true)
	}
	return field
}

// CreateDropDown creates a standard dropdown
func (uf *UIFactory) CreateDropDown(label string) *tview.DropDown {
	dropdown := tview.NewDropDown().SetLabel(label)
	dropdown.SetFieldTextColor(FieldTextColor)
	dropdown.SetFieldBackgroundColor(FieldBackgroundColor)
	dropdown.SetListStyles(UnselectedStyle, SelectedStyle)
	return dropdown
}

// CreateForm creates a standard form
func (uf *UIFactory) CreateForm(title string) *tview.Form {
	form := tview.NewForm()
	form.SetBorder(true).SetTitle(title).SetTitleAlign(tview.AlignCenter)
	form.SetTitleColor(TitleColor)
	form.SetBorderColor(BorderColor)
	return form
}

// CreateButton creates a standard button
func (uf *UIFactory) CreateButton(label string, color tcell.Color) *tview.Button {
	button := tview.NewButton(label)
	button.SetLabelColor(ButtonTextColor)
	button.SetBackgroundColor(color)
	button.SetStyle(tcell.StyleDefault.Foreground(ButtonTextColor).Background(color))
	return button
}

// CreateSaveButton creates a save button
func (uf *UIFactory) CreateSaveButton(label string) *tview.Button {
	return uf.CreateButton("[::b]"+label+"[::-]", SaveButtonColor)
}

// CreateCancelButton creates a cancel button
func (uf *UIFactory) CreateCancelButton(label string) *tview.Button {
	return uf.CreateButton("[::b]"+label+"[::-]", CancelButtonColor)
}

// CreateModalContainer creates a modal dialog container
func (uf *UIFactory) CreateModalContainer(content tview.Primitive) *tview.Flex {
	// Create modal dialog container
	modal := tview.NewFlex().SetDirection(tview.FlexRow)
	modal.AddItem(nil, 0, 1, false) // Top margin

	// Create horizontal layout container
	modalRow := tview.NewFlex().SetDirection(tview.FlexColumn)
	modalRow.AddItem(nil, 0, 1, false)        // Left margin
	modalRow.AddItem(content, FormWidth, 1, true) // Content
	modalRow.AddItem(nil, 0, 1, false)        // Right margin

	modal.AddItem(modalRow, FormHeight, 1, true) // Content row
	modal.AddItem(nil, 0, 1, false)             // Bottom margin

	// Create centered container
	center := tview.NewFlex().SetDirection(tview.FlexRow)
	center.AddItem(nil, 0, 1, false)
	center.AddItem(modal, FormHeight, 1, true)
	center.AddItem(nil, 0, 1, false)

	return center
}

// SetupFormInputCapture sets up form input capture
func (uf *UIFactory) SetupFormInputCapture(form *tview.Form, onEscape func()) {
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			if onEscape != nil {
				onEscape()
			}
			return nil
		case tcell.KeyTab:
			return event
		case tcell.KeyUp, tcell.KeyDown:
			if event.Key() == tcell.KeyUp {
				return tcell.NewEventKey(tcell.KeyBacktab, 0, tcell.ModNone)
			} else {
				return tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone)
			}
		default:
			return event
		}
	})
}

// SetupDropdownInputCapture sets up dropdown input capture
func (uf *UIFactory) SetupDropdownInputCapture(dropdown *tview.DropDown, isOpen *bool) {
	dropdown.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			if !*isOpen {
				*isOpen = true
				return event
			} else {
				*isOpen = false
				return event
			}
		case tcell.KeyEscape:
			if *isOpen {
				*isOpen = false
				return nil
			}
			return event
		case tcell.KeyUp, tcell.KeyDown:
			if *isOpen {
				return event
			}
			return event
		default:
			return event
		}
	})
}