package app

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ShowInfo displays info message
func (a *App) ShowInfo(message string) {
	// Display info in connection info area
	a.ui.ConnInfo.SetText(fmt.Sprintf("[green]Info:[white]\n%s\n", message))
}



// ClearTable clears table data
func (a *App) ClearTable() {
	// Clear table data (keep headers)
	a.ui.ConnTable.Clear()
	// Ensure header row is fixed
	a.ui.ConnTable.SetFixed(1, 0)
}

// SetTableHeaders sets table headers
func (a *App) SetTableHeaders(headers []string) {
	// Save headers
	a.tableHeaders = headers
	// Set UI component headers
	a.ui.TableHeaders = headers
	
	// Set header row
	for i, header := range headers {
		cell := tview.NewTableCell(header).SetTextColor(tcell.ColorYellow).SetSelectable(false)
		a.ui.ConnTable.SetCell(0, i, cell)
	}
}

// AddTableRow adds table row
func (a *App) AddTableRow(rowData []string) {
	// Get current row count
	rowCount := a.ui.ConnTable.GetRowCount()
	
	// Add new row
	for i, cellData := range rowData {
		a.ui.ConnTable.SetCell(rowCount, i, tview.NewTableCell(cellData))
	}
}