package app

import (
	"database/sql"
	"fmt"
	"p6s/internal/config"
	"p6s/internal/db"
	"p6s/internal/k8s"
	"p6s/internal/model"
	"p6s/internal/ui"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// App represents the application
type App struct {
	ui         *ui.Components
	db         *db.PostgresDB
	filterType string
	tableHeaders []string
	connStr    string
	host       string
	port       string
	username   string
	password   string
	database   string
	sslmode    string
	cmdMode    bool
	mouseDisabled bool
	

	k8sClient  *k8s.K8sClient
	k8sConnected bool
	stateManager *StateManager
}

// NewApp creates a new application instance
func NewApp() *App {
	app := &App{
		ui:         ui.NewComponents(),
		db:         db.NewPostgresDB(),
		filterType: "all",
		tableHeaders: []string{"PID", "User", "Database", "Client Address", "Application Name", "Start Time", "Status", "Query"},
		cmdMode:   false,

		host:     "",
		port:     "",
		username: "postgres",
		password: "",
		database: "",
		sslmode:  "disable",
	
		k8sClient: k8s.NewK8sClient(),
		k8sConnected: false,
	
		stateManager: NewStateManager(),
	}


	if err := app.k8sClient.Connect(); err == nil {
		app.k8sConnected = true
	}


	app.setupEventHandlers()

	return app
}

// SetConnectionParams sets connection parameters
func (a *App) SetConnectionParams(host, port, username, password, database, sslmode string) {
	a.host = host
	a.port = port
	a.username = username
	a.password = password
	a.database = database
	a.sslmode = sslmode
	a.connStr = config.BuildConnStr(host, port, username, password, database, sslmode)
}

// Connect connects to database
func (a *App) Connect() error {

	if a.db != nil {
		a.db.Close()
	}
	

	a.db = db.NewPostgresDB()
	

	if err := a.db.Connect(a.connStr); err != nil {

		a.ui.ConnInfo.SetText(fmt.Sprintf("[red]Connection failed: %v[white]\n", err))

		a.ui.ConnTable.Clear()

		for i, header := range a.tableHeaders {
			cell := tview.NewTableCell(header).SetTextColor(tcell.ColorYellow).SetSelectable(false)
			a.ui.ConnTable.SetCell(0, i, cell)
		}

		a.ui.ConnTable.SetCell(1, 0, tview.NewTableCell("Connection failed, please check connection parameters").SetSelectable(false).SetExpansion(1))
		return err
	}


	a.updateInstanceInfo()


	if err := a.refreshData(); err != nil {
		return err
	}


	a.ui.App.SetFocus(a.ui.ConnTable)
	a.ui.UpdateFocusStyle()

	return nil
}

// Run runs the application
func (a *App) Run() error {

	a.ui.App.SetFocus(a.ui.ConnTable)
	a.ui.UpdateFocusStyle()
	

	return a.ui.App.SetRoot(a.ui.Pages, true).EnableMouse(true).Run()
}

// updateInstanceInfo updates instance information
func (a *App) updateInstanceInfo() {

	version, err := a.db.GetDatabaseVersion()
	if err != nil {
		a.ui.ConnInfo.SetText(fmt.Sprintf("[red]Failed to get database version: %v[white]\n", err))
		return
	}


	k8sContext := ""
	if a.k8sConnected {
		k8sContext = a.k8sClient.GetCurrentContext()
	} else {
		k8sContext = "Not connected"
	}


	connInfo := fmt.Sprintf(
		"[yellow]Connection Info:[white]\n" +
		"Host: %s\n" +
		"Port: %s\n" +
		"Username: %s\n" +
		"Database: %s\n" +
		"SSL Mode: %s\n\n" +
		"[yellow]Database Version:[white]\n%s\n\n" +
		"[yellow]Kubernetes Context:[white]\n%s\n",
		a.host, a.port, a.username, a.database, a.sslmode, version, k8sContext)

	a.ui.ConnInfo.SetText(connInfo)
}

// refreshData refreshes data
func (a *App) refreshData() error {

	currentDB, err := a.db.GetCurrentDatabase()
	if err != nil {
		a.ui.ConnInfo.SetText(fmt.Sprintf("[red]Failed to get current database: %v[white]\n", err))
		return err
	}


	a.database = currentDB


	switch a.filterType {
	case "all", "active", "blocked":

		a.ui.TableHeaders = a.tableHeaders


		connections, err := a.db.GetConnections(a.filterType)
		if err != nil {
			a.ui.ConnInfo.SetText(fmt.Sprintf("[red]Failed to get connection info: %v[white]\n", err))
			return err
		}


		a.ui.DisplayConnections(connections)

	case "table_size":

		a.ui.TableHeaders = a.tableHeaders


		tableStats, err := a.db.GetTableStats()
		if err != nil {
			a.ui.ConnInfo.SetText(fmt.Sprintf("[red]Failed to get table statistics: %v[white]\n", err))
			return err
		}


		a.ui.DisplayTableStats(tableStats)

	case "custom":

		a.ui.TableHeaders = a.tableHeaders


		customData := []model.Connection{{
			PID:             0,
			User:            sql.NullString{String: "Press ':' key", Valid: true},
		Database:        sql.NullString{String: "Enter command mode", Valid: true},
		ClientAddr:      sql.NullString{String: "Execute custom SQL", Valid: true},
		ApplicationName: sql.NullString{String: "Query", Valid: true},
			BackendStart:    time.Now(),
			State:           sql.NullString{String: "Tip", Valid: true},
		Query:           sql.NullString{String: "Custom query mode", Valid: true},
		}}
		a.ui.DisplayConnections(customData)
	}

	return nil
}

// setupEventHandlers sets up event handlers
func (a *App) setupEventHandlers() {

	a.ui.App.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {

		if event.Key() == tcell.KeyRune && event.Rune() == ':' {

			a.cmdMode = true
			a.ui.CmdInput.SetText("")
			a.ui.FlexBox.ResizeItem(a.ui.CmdInput, 1, 0)
			a.ui.App.SetFocus(a.ui.CmdInput)
			return nil
		}
		

		if pageName, _ := a.ui.Pages.GetFrontPage(); pageName == "sql_query" {

			if event.Key() == tcell.KeyEnter {

				if sqlTextArea, ok := a.ui.App.GetFocus().(*tview.TextArea); ok {
					sqlQuery := sqlTextArea.GetText()
					if sqlQuery != "" {
		
						a.ui.Pages.RemovePage("sql_query")
						a.ui.App.SetFocus(a.ui.ConnTable)
		
						a.executeCustomSQL(sqlQuery)
					}
				}
				return nil
			}
			return event
		}
		

		if event.Key() == tcell.KeyRune {
			
			switch event.Rune() {
			case '1':
				a.filterType = "all"
				a.tableHeaders = []string{"PID", "User", "Database", "Client Address", "Application Name", "Start Time", "Status", "Query"}
				a.ui.TableHeaders = a.tableHeaders
				if err := a.refreshData(); err != nil {
	
				}

				a.ui.App.SetFocus(a.ui.ConnTable)
				a.ui.UpdateFocusStyle()
				return nil
			case '2':
				a.filterType = "active"
				a.tableHeaders = []string{"PID", "User", "Database", "Client Address", "Application Name", "Start Time", "Status", "Query"}
				a.ui.TableHeaders = a.tableHeaders
				if err := a.refreshData(); err != nil {
	
				}

				a.ui.App.SetFocus(a.ui.ConnTable)
				a.ui.UpdateFocusStyle()
				return nil
			case '3':
				a.filterType = "blocked"
				a.tableHeaders = []string{"PID", "User", "Database", "Client Address", "Application Name", "Start Time", "Status", "Query"}
				a.ui.TableHeaders = a.tableHeaders
				if err := a.refreshData(); err != nil {
	
				}

				a.ui.App.SetFocus(a.ui.ConnTable)
				a.ui.UpdateFocusStyle()
				return nil
			case '4':
				a.filterType = "table_size"
				a.tableHeaders = []string{"Schema", "Table Name", "Total Size", "Table Size", "Index Size", "Total Rows"}
				a.ui.TableHeaders = a.tableHeaders
				if err := a.refreshData(); err != nil {
	
				}

				a.ui.App.SetFocus(a.ui.ConnTable)
				a.ui.UpdateFocusStyle()
				return nil
			case '5':

				a.showSQLQueryForm()

				a.ui.App.SetFocus(a.ui.ConnTable)
				a.ui.UpdateFocusStyle()
				return nil
			}
		}
		return event
	})



	a.ui.CmdInput.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter {
	
			cmd := a.ui.CmdInput.GetText()
	
			a.handleCommand(cmd)
	
			a.cmdMode = false
			a.ui.CmdInput.SetText("")
			a.ui.FlexBox.ResizeItem(a.ui.CmdInput, 0, 0)
			a.ui.App.SetFocus(a.ui.ConnTable)
			return nil
		} else if event.Key() == tcell.KeyBackspace || event.Key() == tcell.KeyBackspace2 {
			text := a.ui.CmdInput.GetText()
	
			if len(text) == 0 {
	
				a.cmdMode = false
				a.ui.CmdInput.SetText("")
				a.ui.FlexBox.ResizeItem(a.ui.CmdInput, 0, 0)
				a.ui.App.SetFocus(a.ui.ConnTable)
				return nil
			}
		}
		return event
	})
}

// handleCommand handles commands
func (a *App) handleCommand(cmd string) {

	if strings.HasPrefix(cmd, "\\c") && !strings.HasPrefix(cmd, "\\config") {

		parts := strings.Fields(cmd)
		if len(parts) > 1 {

			dbName := parts[1]
			
			a.database = dbName
			
			a.connStr = config.BuildConnStr(a.host, a.port, a.username, a.password, a.database, a.sslmode)

			if err := a.Connect(); err != nil {

				a.ui.ConnInfo.SetText(fmt.Sprintf("[red]Failed to connect to database %s: %v[white]\n", dbName, err))
			}
		} else {

			databases, err := a.db.GetDatabases()
			if err != nil {
				a.ui.ConnInfo.SetText(fmt.Sprintf("[red]Failed to get database list: %v[white]\n", err))
				return
			}

	
			a.showDatabaseSelectionForm(databases)
		}
	} else if cmd == "\\config" {

		a.showConfigForm()
	} else if cmd == "\\configk8s" {


		if !a.k8sConnected {
			a.ShowError("Not connected to Kubernetes cluster, please ensure your kubeconfig is configured correctly")
			return
		}
		
		
		a.showK8sConfigForm()
	}
}

// showDatabaseSelectionForm shows database selection form
func (a *App) showDatabaseSelectionForm(databases []string) {

	form := tview.NewForm()


	dropdown := tview.NewDropDown().SetLabel("Select database: ")
	dropdown.SetFieldTextColor(tcell.ColorWhite)
	dropdown.SetFieldBackgroundColor(tcell.ColorBlack)

	unselectedStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)
	selectedStyle := tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorWhite)
	dropdown.SetListStyles(unselectedStyle, selectedStyle)
	
	var selectedIndex int = 0
	var selectedDB string = ""
	

	dropdown.SetOptions(databases, func(text string, index int) {
		selectedIndex = index
		selectedDB = text
	})
	

	if len(databases) > 0 {
		dropdown.SetCurrentOption(0)
		selectedIndex = 0
		selectedDB = databases[0]
	}


	form.AddFormItem(dropdown)
	

	form.AddTextView("Operation Tip", "Y Confirm  N Cancel", 0, 1, false, true)
	

	form.SetFieldTextColor(tcell.ColorWhite)
	form.SetFieldBackgroundColor(tcell.ColorBlack)


	form.SetBorder(true).SetTitle("Switch Database").SetTitleAlign(tview.AlignCenter)
	form.SetTitleColor(tcell.ColorWhite)
	form.SetBorderColor(tcell.ColorWhite)


	modal := tview.NewFlex().SetDirection(tview.FlexRow)
	modal.AddItem(nil, 0, 1, false)


	modalRow := tview.NewFlex().SetDirection(tview.FlexColumn)
	modalRow.AddItem(nil, 0, 1, false)
	modalRow.AddItem(form, 40, 1, true)
	modalRow.AddItem(nil, 0, 1, false)

	modal.AddItem(modalRow, 10, 1, true)
	modal.AddItem(nil, 0, 1, false)


	center := tview.NewFlex().SetDirection(tview.FlexRow)
	center.AddItem(nil, 0, 1, false)
	center.AddItem(modal, 10, 1, true)
	center.AddItem(nil, 0, 1, false)


	center.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {

		if event.Key() == tcell.KeyEscape {
			a.ui.Pages.RemovePage("db_select")
			a.ui.App.SetFocus(a.ui.ConnTable)
			return nil
		}
		return event
	})


	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {

		if event.Key() == tcell.KeyTab {
			return nil
		}
		
		switch {
		case event.Rune() == 'y' || event.Rune() == 'Y':
	
			if selectedIndex < 0 || selectedIndex >= len(databases) || selectedDB == "" {
				return nil
			}
			a.database = selectedDB
			
			a.connStr = config.BuildConnStr(a.host, a.port, a.username, a.password, a.database, a.sslmode)
			
			if err := a.Connect(); err != nil {
				
				a.ui.ConnInfo.SetText(fmt.Sprintf("[red]Failed to connect to database %s: %v[white]\n", selectedDB, err))
			}
			
			a.refreshData()
			
			a.ui.Pages.RemovePage("db_select")
			a.ui.App.SetFocus(a.ui.ConnTable)
			return nil
		case event.Rune() == 'n' || event.Rune() == 'N':
	
			a.ui.Pages.RemovePage("db_select")
			a.ui.App.SetFocus(a.ui.ConnTable)
			return nil
		case event.Key() == tcell.KeyEscape:
	
			a.ui.Pages.RemovePage("db_select")
			a.ui.App.SetFocus(a.ui.ConnTable)
			return nil
		default:
			return event
		}
	})


	a.ui.Pages.RemovePage("db_select")
	a.ui.Pages.AddPage("db_select", center, true, true)

	a.ui.App.SetFocus(a.ui.ConnTable)


	go func() {
		time.Sleep(5 * time.Millisecond)

		a.ui.App.QueueUpdate(func() {
			if len(databases) > 0 {
				a.ui.App.SetFocus(dropdown)
			} else {
				a.ui.App.SetFocus(form)
			}
		})
	}()
}

// showSQLQueryForm shows SQL query window
func (a *App) showSQLQueryForm() {

	form := tview.NewForm()


	sqlTextArea := tview.NewTextArea()
	sqlTextArea.SetLabel("SQL Query Statement: ")
	sqlTextArea.SetPlaceholder("Please enter your SQL query statement...")
	sqlTextArea.SetWrap(true)
	sqlTextArea.SetWordWrap(true)

	sqlTextArea.SetMaxLength(50000)

	sqlTextArea.SetChangedFunc(nil)


	sqlTextArea.SetBorderColor(tcell.ColorWhite)
	sqlTextArea.SetBorder(true)

	sqlTextArea.SetSize(20, 30)


	sqlTextArea.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {

		if event.Key() == tcell.KeyCtrlV || (event.Key() == tcell.KeyRune && event.Rune() == 'v' && event.Modifiers()&tcell.ModCtrl != 0) {

			go func() {
				time.Sleep(10 * time.Millisecond)
				a.ui.App.QueueUpdateDraw(func() {

				})
			}()
			return event
		}

		if event.Key() == tcell.KeyEnter {

			sqlQuery := sqlTextArea.GetText()
			if sqlQuery != "" {

				a.closeSQLQueryForm()

				a.executeCustomSQL(sqlQuery)
			}
			return nil
		}

		if event.Rune() >= '1' && event.Rune() <= '5' {
			return event
		}

		return event
	})


	form.AddFormItem(sqlTextArea)


	form.AddButton(" Execute [Enter] ", func() {

		sqlQuery := sqlTextArea.GetText()
		if sqlQuery == "" {
			a.ShowError("Please enter SQL query statement")
			return
		}


		a.closeSQLQueryForm()


		a.executeCustomSQL(sqlQuery)
	})


	executeButton := form.GetButton(0)


	executeButton.SetLabelColor(tcell.ColorWhite)
	executeButton.SetBackgroundColor(tcell.ColorBlue)
	executeButton.SetLabel("[::b]    Execute (Enter)    [::-]")


	form.SetBorder(true).SetTitle("Custom SQL Query").SetTitleAlign(tview.AlignCenter)
	form.SetTitleColor(tcell.ColorWhite)
	form.SetBorderColor(tcell.ColorWhite)
	form.SetFieldTextColor(tcell.ColorWhite)
	form.SetFieldBackgroundColor(tcell.ColorBlack)
	form.SetButtonsAlign(tview.AlignCenter)


	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:

			a.closeSQLQueryForm()
			return nil
		case tcell.KeyCtrlS:

			sqlQuery := sqlTextArea.GetText()
			if sqlQuery != "" {

				a.closeSQLQueryForm()

				a.executeCustomSQL(sqlQuery)
			}
			return nil
		default:

			if event.Rune() >= '1' && event.Rune() <= '5' {
				return event
			}
			return event
		}
	})


	modal := tview.NewFlex().SetDirection(tview.FlexRow)
	modal.AddItem(nil, 0, 1, false)


	modalRow := tview.NewFlex().SetDirection(tview.FlexColumn)
	modalRow.AddItem(nil, 0, 1, false)
	modalRow.AddItem(form, 60, 1, true)
	modalRow.AddItem(nil, 0, 1, false)

	modal.AddItem(modalRow, 30, 1, true)
	modal.AddItem(nil, 0, 1, false)


	center := tview.NewFlex().SetDirection(tview.FlexRow)
	center.AddItem(nil, 0, 1, false)
	center.AddItem(modal, 30, 1, true)
	center.AddItem(nil, 0, 1, false)


	a.mouseDisabled = true
	

	a.ui.App.SetMouseCapture(func(event *tcell.EventMouse, action tview.MouseAction) (*tcell.EventMouse, tview.MouseAction) {
		if a.mouseDisabled {

			return nil, action
		}
		return event, action
	})


	a.ui.Pages.AddPage("sql_query", center, true, true)
	a.ui.App.SetFocus(sqlTextArea)


	go func() {
		time.Sleep(100 * time.Millisecond)
		a.ui.App.QueueUpdateDraw(func() {
			a.ui.App.SetFocus(sqlTextArea)
		})
	}()
}

// closeSQLQueryForm closes SQL query window and restores mouse functionality
func (a *App) closeSQLQueryForm() {

	a.mouseDisabled = false

	a.ui.App.SetMouseCapture(nil)

	a.ui.Pages.RemovePage("sql_query")
	a.ui.App.SetFocus(a.ui.ConnTable)
}

// executeCustomSQL executes custom SQL query
func (a *App) executeCustomSQL(sqlQuery string) {

	if a.db == nil {

		errorHeaders := []string{"Error"}
		errorResults := [][]interface{}{{"Not connected to database"}}
		a.filterType = "custom"
		a.tableHeaders = errorHeaders
		a.ui.TableHeaders = errorHeaders
		a.ui.DisplayCustomQueryResults(errorResults, errorHeaders)
		return
	}


	results, headers, err := a.db.ExecuteCustomQuery(sqlQuery)
	if err != nil {

		errorHeaders := []string{"Error"}
		errorResults := [][]interface{}{{fmt.Sprintf("SQL query execution failed: %v", err)}}
		a.filterType = "custom"
		a.tableHeaders = errorHeaders
		a.ui.TableHeaders = errorHeaders
		a.ui.DisplayCustomQueryResults(errorResults, errorHeaders)
		return
	}


	a.filterType = "custom"
	a.tableHeaders = headers
	a.ui.TableHeaders = headers


	a.ui.DisplayCustomQueryResults(results, headers)



}

// ShowError displays error information
func (a *App) ShowError(message string) {
	a.ui.ConnInfo.SetText(fmt.Sprintf("[red]%s[white]\n", message))
}

// showConfigForm shows configuration form
func (a *App) showConfigForm() {
	// Create configuration form
	form := tview.NewForm()

	// Add input fields
	hostField := form.AddInputField("Host", a.host, 30, nil, nil)
	hostField.SetFieldTextColor(tcell.ColorWhite)
	hostField.SetFieldBackgroundColor(tcell.ColorBlack)
	
	portField := form.AddInputField("Port", a.port, 30, nil, nil)
	portField.SetFieldTextColor(tcell.ColorWhite)
	portField.SetFieldBackgroundColor(tcell.ColorBlack)
	
	usernameField := form.AddInputField("Username", a.username, 30, nil, nil)
	usernameField.SetFieldTextColor(tcell.ColorWhite)
	usernameField.SetFieldBackgroundColor(tcell.ColorBlack)
	
	form.AddInputField("Password", a.password, 30, nil, nil)
	passwordField := form.GetFormItem(3).(*tview.InputField) // Get password field
	passwordField.SetFieldTextColor(tcell.ColorWhite)
	passwordField.SetFieldBackgroundColor(tcell.ColorBlack)
	passwordField.SetMaskCharacter('*') // Set password mask character to asterisk
	
	databaseField := form.AddInputField("Database", a.database, 30, nil, nil)
	databaseField.SetFieldTextColor(tcell.ColorWhite)
	databaseField.SetFieldBackgroundColor(tcell.ColorBlack)


	// Add buttons
	form.AddButton("Save", func() {
		// Immediately remove page and set focus to avoid UI freeze
		a.ui.Pages.RemovePage("config")
		a.ui.App.SetFocus(a.ui.ConnTable)
		
		// Use goroutine to handle time-consuming operations to avoid blocking UI
		go func() {
			// Get values from form
			host := form.GetFormItem(0).(*tview.InputField).GetText()
			port := form.GetFormItem(1).(*tview.InputField).GetText()
			username := form.GetFormItem(2).(*tview.InputField).GetText()
			password := form.GetFormItem(3).(*tview.InputField).GetText()
			database := form.GetFormItem(4).(*tview.InputField).GetText()
			sslmode := "disable" // Hardcode SSL Mode to disable

			// Update application connection parameters
			a.host = host
			a.port = port
			a.username = username
			a.password = password
			a.database = database
			a.sslmode = sslmode
			a.connStr = config.BuildConnStr(host, port, username, password, database, sslmode)

			// Load existing config to preserve K8s related fields
			existingConfig, _ := config.LoadConfig()
			
			// Create new config, preserve K8s fields
			cfg := &config.Config{
				Host:     host,
				Port:     port,
				Username: username,
				Password: password,
				Database: database,
				SSLMode:  sslmode,
			}
			
			// If K8s config exists, preserve these fields
			if existingConfig != nil {
				cfg.Namespace = existingConfig.Namespace
				cfg.Pod = existingConfig.Pod
				cfg.Container = existingConfig.Container
				cfg.PortName = existingConfig.PortName
				cfg.Secret = existingConfig.Secret
				cfg.SecretKey = existingConfig.SecretKey
			}

			var saveErr, connErr, refreshErr error
			
			// Save config to file
			saveErr = config.SaveConfig(cfg)
			
			// Try to connect to database
			if saveErr == nil {
				connErr = a.Connect()
			}
			
			// Refresh data
			if connErr == nil {
				refreshErr = a.refreshData()
			}
			
			// Use QueueUpdateDraw to update UI
			a.ui.App.QueueUpdateDraw(func() {
				// Get current connection info text
				currentText := a.ui.ConnInfo.GetText(true) // true means get raw text including color tags
				
				// Display operation result (append to existing info)
				var resultMsg string
				if saveErr != nil {
					resultMsg = fmt.Sprintf("\n[red]Failed to save config: %v[white]", saveErr)
				} else if connErr != nil {
					resultMsg = fmt.Sprintf("\n[red]Failed to connect to database: %v[white]", connErr)
				} else if refreshErr != nil {
					resultMsg = fmt.Sprintf("\n[red]Failed to refresh data: %v[white]", refreshErr)
				} else {
					resultMsg = "\n[green]Configuration saved and successfully connected to database[white]"
				}
				
				// Append result message to existing text
				a.ui.ConnInfo.SetText(currentText + resultMsg)
			})
		}()
	})

	// Add cancel button
	form.AddButton("Cancel", func() {
		// Immediately remove page and set focus to avoid UI freeze
		a.ui.Pages.RemovePage("config")
		a.ui.App.SetFocus(a.ui.ConnTable)
	})

	// Set form style
	form.SetBorder(true).SetTitle("Configure Connection").SetTitleAlign(tview.AlignCenter)
	form.SetTitleColor(tcell.ColorWhite)
	form.SetBorderColor(tcell.ColorWhite)

	// Get buttons and set styles
	saveButton := form.GetButton(0) // Save button
	cancelButton := form.GetButton(1) // Cancel button
	
	// Set button colors to ensure visibility in various terminal backgrounds
	saveButton.SetLabelColor(tcell.ColorWhite)
	saveButton.SetBackgroundColor(tcell.ColorBlue)
	saveButton.SetLabel("[::b]Save[::-]")
	cancelButton.SetLabelColor(tcell.ColorWhite)
	cancelButton.SetBackgroundColor(tcell.ColorRed)
	cancelButton.SetLabel("[::b]Cancel[::-]")
	
	// Set button styles
	saveButton.SetStyle(tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlue))
	cancelButton.SetStyle(tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorRed))
	
	// Set button center alignment
	form.SetButtonsAlign(tview.AlignCenter)

	// Set form input capture, handle Escape, Tab, and up/down keys
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			// Immediately remove page and set focus to avoid UI freeze
			a.ui.Pages.RemovePage("config")
			a.ui.App.SetFocus(a.ui.ConnTable)
			return nil
		case tcell.KeyTab:
			return nil
		case tcell.KeyUp:
			// Get current focus item
			currentFocus := a.ui.App.GetFocus()
			
			// If current focus is on Save button
			if currentFocus == saveButton {
				// Move to last form item
				a.ui.App.SetFocus(form.GetFormItem(form.GetFormItemCount()-1))
				return nil
			}
			
			// If current focus is on Cancel button
			if currentFocus == cancelButton {
				// Move to Save button
				a.ui.App.SetFocus(saveButton)
				return nil
			}
			
			// Move between form items
			for i := 0; i < form.GetFormItemCount(); i++ {
				if form.GetFormItem(i) == currentFocus {
					// If not the first item, move focus up
					if i > 0 {
						a.ui.App.SetFocus(form.GetFormItem(i-1))
						return nil
					}
					break
				}
			}
			return event
		case tcell.KeyDown:
			// Get current focus item
			currentFocus := a.ui.App.GetFocus()
			
			// Move between form items
			for i := 0; i < form.GetFormItemCount(); i++ {
				if form.GetFormItem(i) == currentFocus {
					// If it's the last item, move to Save button
					if i == form.GetFormItemCount()-1 {
						a.ui.App.SetFocus(saveButton)
						return nil
					}
					// Otherwise move to next form item
					if i < form.GetFormItemCount()-1 {
						a.ui.App.SetFocus(form.GetFormItem(i+1))
						return nil
					}
					break
				}
			}
			
			// If current focus is on Save button
			if currentFocus == saveButton {
				// Move to Cancel button
				a.ui.App.SetFocus(cancelButton)
				return nil
			}
			return event
		default:
			return event
		}
	})

	// Create modal dialog container
	modal := tview.NewFlex().SetDirection(tview.FlexRow)
	modal.AddItem(nil, 0, 1, false) // Top margin

	// Create horizontal layout container
	modalRow := tview.NewFlex().SetDirection(tview.FlexColumn)
	modalRow.AddItem(nil, 0, 1, false) // Left margin
	modalRow.AddItem(form, 50, 1, true) // Form
	modalRow.AddItem(nil, 0, 1, false) // Right margin

	modal.AddItem(modalRow, 18, 1, true) // Form row, increase height from 15 to 18
	modal.AddItem(nil, 0, 1, false) // Bottom margin

	// Create a centered container
	center := tview.NewFlex().SetDirection(tview.FlexRow)
	center.AddItem(nil, 0, 1, false)
	center.AddItem(modal, 18, 1, true) // Increase height from 15 to 18
	center.AddItem(nil, 0, 1, false)

	// First remove any existing old page, then add new modal dialog
	a.ui.Pages.RemovePage("config")
	a.ui.Pages.AddPage("config", center, true, true)

	// Set focus to the first input field of the form
	a.ui.App.SetFocus(form.GetFormItem(0))
	
	// Use delayed execution to ensure focus is set correctly
	go func() {
		time.Sleep(100 * time.Millisecond)
		a.ui.App.QueueUpdateDraw(func() {
			a.ui.App.SetFocus(form.GetFormItem(0))
		})
	}()
}