package ui

import (
	"database/sql"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"p6s/internal/model"
	"time"
)

// Components contains all UI components
type Components struct {
	App          *tview.Application
	Pages        *tview.Pages
	FlexBox      *tview.Flex
	MenuList     *tview.List
	MenuList2    *tview.List
	SwitchDBList *tview.List
	ConnTable    *tview.Table
	ConnInfo     *tview.TextView
	CmdInput     *tview.InputField
	TableHeaders []string
}

// NewComponents creates and initializes UI components
func NewComponents() *Components {
	components := &Components{
		App:          tview.NewApplication(),
		Pages:        tview.NewPages(),
		ConnTable:    tview.NewTable().SetBorders(true).SetFixed(1, 0),
		MenuList:     tview.NewList().ShowSecondaryText(false),
		CmdInput:     tview.NewInputField().SetLabel(":").SetFieldWidth(30).SetFieldBackgroundColor(tcell.ColorBlack),
	}


	components.MenuList.SetBorder(false).SetTitle("Options").SetTitleAlign(tview.AlignLeft)
	components.MenuList.AddItem("All Connections", "", '1', nil)
	components.MenuList.AddItem("Active Connections", "", '2', nil)
	components.MenuList.AddItem("Block Connections", "", '3', nil)
	components.MenuList.AddItem("Show Table Statics", "", '4', nil)
	components.MenuList.SetMainTextColor(tcell.ColorWhite)
	components.MenuList.SetSelectedTextColor(tcell.ColorWhite)
	components.MenuList.SetSelectedBackgroundColor(tcell.ColorBlack)
	

	components.MenuList2 = tview.NewList().ShowSecondaryText(false)
	components.MenuList2.SetBorder(false).SetTitle("Custom").SetTitleAlign(tview.AlignLeft)
	components.MenuList2.AddItem("Custom Query", "", '5', nil)
	components.MenuList2.SetMainTextColor(tcell.ColorWhite)
	components.MenuList2.SetSelectedTextColor(tcell.ColorWhite)
	components.MenuList2.SetSelectedBackgroundColor(tcell.ColorBlack)
	

	components.MenuList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		return nil
	})

	components.MenuList.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
		return action, nil
	})
	

	components.MenuList2.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		return nil
	})

	components.MenuList2.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
		return action, nil
	})
	

	components.SwitchDBList = tview.NewList().ShowSecondaryText(false)
	components.SwitchDBList.SetBorder(true).SetTitle("Help").SetTitleAlign(tview.AlignLeft)
	components.SwitchDBList.AddItem("[::d] :            Enter Command Line[-]", "",0, nil)
	components.SwitchDBList.AddItem("[::d] [\\c]         Switch Database[-]", "",0, nil)
	components.SwitchDBList.AddItem("[::d] [\\config]    Configure Connection[-]", "",0, nil)
	components.SwitchDBList.AddItem("[::d] [\\configk8s] Configure K8s Connection[-]", "",0, nil)
	components.SwitchDBList.SetMainTextColor(tcell.ColorWhite)
	components.SwitchDBList.SetSelectedTextColor(tcell.ColorDarkGrey)
	components.SwitchDBList.SetSelectedBackgroundColor(tcell.ColorBlack)
	

	components.SwitchDBList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		return nil
	})

	components.SwitchDBList.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
		return action, nil
	})


	headers := []string{"PID", "User", "Database", "Client Address", "Application Name", "Start Time", "Status", "Query"}
	for i, header := range headers {
		cell := tview.NewTableCell(header).SetTextColor(tcell.ColorYellow).SetSelectable(false)
		components.ConnTable.SetCell(0, i, cell)
	}


	components.ConnTable.SetBorder(true).SetTitle("Result Table").SetTitleAlign(tview.AlignLeft)
	components.ConnTable.SetSelectable(true, false)


	components.ConnInfo = tview.NewTextView().SetDynamicColors(true)
	components.ConnInfo.SetBorder(true).SetTitle("Instance Info").SetTitleAlign(tview.AlignLeft)
	

	components.ConnInfo.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		return nil
	})

	components.ConnInfo.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
		return action, nil
	})


	components.FlexBox = createLayout(components)


	components.Pages.AddPage("main", components.FlexBox, true, true)

	return components
}

// createLayout creates application layout
func createLayout(c *Components) *tview.Flex {

	p6sHeader := tview.NewTextView()
	bannerText := "[yellow::b]╔══════════════════════════════════════════════════════════════════════════════╗[-:-:-]\n" +
				 "[yellow::b][white::b]                           🐘 p6s - Postgres TUI 💻                           [yellow::b][-:-:-]\n" +
				 "[yellow::b]╚══════════════════════════════════════════════════════════════════════════════╝[-:-:-]"
	p6sHeader.SetText(bannerText)
	p6sHeader.SetDynamicColors(true)
	p6sHeader.SetTextAlign(tview.AlignCenter)
	p6sHeader.SetBorder(false)
	
	

	optionsFlex := tview.NewFlex().SetDirection(tview.FlexColumn)
	optionsFlex.AddItem(c.MenuList, 0, 1, false)
	optionsFlex.AddItem(c.MenuList2, 0, 1, false)
	

	optionsAreaFlex := tview.NewFlex().SetDirection(tview.FlexRow)
	optionsAreaFlex.AddItem(optionsFlex, 0, 1, false)
	

	topMenuFlex := tview.NewFlex().SetDirection(tview.FlexColumn)
	topMenuFlex.AddItem(optionsAreaFlex, 0, 3, false)
	topMenuFlex.AddItem(c.SwitchDBList, 0, 1, false)


	bottomContentFlex := tview.NewFlex().SetDirection(tview.FlexColumn)
	bottomContentFlex.AddItem(c.ConnTable, 0, 3, true)
	bottomContentFlex.AddItem(c.ConnInfo, 0, 1, false)


	mainFlex := tview.NewFlex().SetDirection(tview.FlexRow)
	mainFlex.AddItem(p6sHeader, 4, 0, false)
	mainFlex.AddItem(topMenuFlex, 6, 0, false)
	mainFlex.AddItem(bottomContentFlex, 0, 1, true)


	finalFlex := tview.NewFlex().SetDirection(tview.FlexRow)
	finalFlex.AddItem(mainFlex, 0, 1, true)
	finalFlex.AddItem(c.CmdInput, 0, 0, false)

	return finalFlex
}

// UpdateFocusStyle updates focus style
func (c *Components) UpdateFocusStyle() {

	c.MenuList.SetBorder(false)
	c.MenuList.SetTitle("[::b]Options[-]")
	c.MenuList2.SetBorder(false)
	c.MenuList2.SetTitle("[::b] [-]")
	c.ConnTable.SetBorderColor(tcell.ColorWhite)
	c.ConnTable.SetTitle("[::b]Result Table[-]")
	c.ConnInfo.SetBorderColor(tcell.ColorWhite)
	c.ConnInfo.SetTitle("[::b]Instance Info[-]")


	switch c.App.GetFocus() {
	case c.MenuList:
		c.MenuList.SetBorder(false)
		c.MenuList.SetTitle("[::b]Options[-]")
	case c.MenuList2:
		c.MenuList2.SetBorder(false)
		c.MenuList2.SetTitle("[::b] [-]")
	case c.ConnTable:
		c.ConnTable.SetBorderColor(tcell.ColorWhite)
		c.ConnTable.SetTitle("[::b]Result Table[-]")
	case c.ConnInfo:
		c.ConnInfo.SetTitle("[::b]Instance Info[-]")
	}
}

// DisplayConnections displays connection information in table
func (c *Components) DisplayConnections(connections []model.Connection) {

	c.ConnTable.Clear()

	c.ConnTable.SetFixed(1, 0)


	if len(c.TableHeaders) == 0 {
	
		c.TableHeaders = []string{"PID", "User", "Database", "Client Address", "Application Name", "Start Time", "Status", "Query"}
	}
	
	for i, header := range c.TableHeaders {
		cell := tview.NewTableCell(header).SetTextColor(tcell.ColorYellow).SetSelectable(false)
		c.ConnTable.SetCell(0, i, cell)
	}


	if len(connections) == 0 {

		cell := tview.NewTableCell("No connection information").SetSelectable(true).SetExpansion(1)
		cell.SetAlign(tview.AlignCenter)
		c.ConnTable.SetCell(1, 0, cell)
		

		for i := 1; i < len(c.TableHeaders); i++ {
			c.ConnTable.SetCell(1, i, tview.NewTableCell("").SetSelectable(true))
		}
		

		c.ConnTable.Select(1, 0)
		return
	}


	for i, conn := range connections {
		row := i + 1

		// PID
		c.ConnTable.SetCell(row, 0, tview.NewTableCell(formatInt(conn.PID)))


		c.ConnTable.SetCell(row, 1, tview.NewTableCell(formatNullString(conn.User)))


		c.ConnTable.SetCell(row, 2, tview.NewTableCell(formatNullString(conn.Database)))


		c.ConnTable.SetCell(row, 3, tview.NewTableCell(formatNullString(conn.ClientAddr)))


		c.ConnTable.SetCell(row, 4, tview.NewTableCell(formatNullString(conn.ApplicationName)))


		c.ConnTable.SetCell(row, 5, tview.NewTableCell(formatTime(conn.BackendStart)))


		c.ConnTable.SetCell(row, 6, tview.NewTableCell(formatNullString(conn.State)))


		c.ConnTable.SetCell(row, 7, tview.NewTableCell(formatNullString(conn.Query)))
	}
	

	if len(connections) > 0 {
		c.ConnTable.Select(1, 0)
	}
	

	c.ConnTable.ScrollToBeginning()
}

// DisplayTableStats displays table statistics in table
func (c *Components) DisplayTableStats(tableStats []model.TableStat) {

	c.ConnTable.Clear()

	c.ConnTable.SetFixed(1, 0)


	if len(c.TableHeaders) == 0 {
	
		c.TableHeaders = []string{"Schema", "Table Name", "Total Size", "Table Size", "Index Size", "Total Rows"}
	}
	
	for i, header := range c.TableHeaders {
		cell := tview.NewTableCell(header).SetTextColor(tcell.ColorYellow).SetSelectable(false)
		c.ConnTable.SetCell(0, i, cell)
	}


	if len(tableStats) == 0 {

		cell := tview.NewTableCell("No table statistics").SetSelectable(true).SetExpansion(1)
		cell.SetAlign(tview.AlignCenter)
		c.ConnTable.SetCell(1, 0, cell)
		

		for i := 1; i < len(c.TableHeaders); i++ {
			c.ConnTable.SetCell(1, i, tview.NewTableCell("").SetSelectable(true))
		}
		

		c.ConnTable.Select(1, 0)
		return
	}


	for i, stat := range tableStats {
		row := i + 1


		c.ConnTable.SetCell(row, 0, tview.NewTableCell(stat.Schema))


		c.ConnTable.SetCell(row, 1, tview.NewTableCell(stat.Name))


		c.ConnTable.SetCell(row, 2, tview.NewTableCell(stat.TotalSize))


		c.ConnTable.SetCell(row, 3, tview.NewTableCell(stat.TableSize))


		c.ConnTable.SetCell(row, 4, tview.NewTableCell(stat.IndexSize))


		c.ConnTable.SetCell(row, 5, tview.NewTableCell(formatInt64(stat.RowCount)))
	}
	

	if len(tableStats) > 0 {
		c.ConnTable.Select(1, 0)
	}
	

	c.ConnTable.ScrollToBeginning()
}


func formatNullString(s sql.NullString) string {
	if s.Valid {
		return s.String
	}
	return ""
}

func formatInt(i int) string {
	return fmt.Sprintf("%d", i)
}

func formatInt64(i int64) string {
	return fmt.Sprintf("%d", i)
}

func formatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// DisplayCustomQueryResults displays custom query results
func (c *Components) DisplayCustomQueryResults(results [][]interface{}, headers []string) {

	c.ConnTable.Clear()


	for i, header := range headers {
		cell := tview.NewTableCell(header)
		cell.SetTextColor(tcell.ColorYellow)
		cell.SetAlign(tview.AlignCenter)
		cell.SetSelectable(false)
		c.ConnTable.SetCell(0, i, cell)
	}


	if len(results) == 0 {

		cell := tview.NewTableCell("No query results")
		cell.SetTextColor(tcell.ColorRed)
		cell.SetAlign(tview.AlignCenter)
		c.ConnTable.SetCell(1, 0, cell)
		

		for i := 1; i < len(headers); i++ {
			c.ConnTable.SetCell(1, i, tview.NewTableCell("").SetSelectable(true))
		}
		

		c.ConnTable.Select(1, 0)
		return
	}


	for i, result := range results {
		row := i + 1


		for j, value := range result {
			if j < len(headers) {
				cellValue := ""
				if value != nil {
					cellValue = fmt.Sprintf("%v", value)
				}
				c.ConnTable.SetCell(row, j, tview.NewTableCell(cellValue))
			}
		}
		

		for j := len(result); j < len(headers); j++ {
			c.ConnTable.SetCell(row, j, tview.NewTableCell(""))
		}
	}
	

	if len(results) > 0 {
		c.ConnTable.Select(1, 0)
	}
	

	c.ConnTable.ScrollToBeginning()
}