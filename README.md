# p6s - PostgreSQL Terminal Management Tool

<div align="center">
  <img src="assets/p6s.png" alt="p6s" width="200" height="auto">
</div>

## Introduction

p6s is a terminal-based PostgreSQL database management tool developed in Go, providing an intuitive text user interface (TUI) for managing and monitoring PostgreSQL database connections.

## Key Features

- **Connection Management**: Connect to PostgreSQL database servers with support for configuring and saving connection information
- **Connection Monitoring**: Real-time viewing and monitoring of database connection status
- **Database Switching**: Quick switching between different databases
- **Connection Filtering**: Support for filtering connections by different conditions (all connections, active connections, blocked connections)
- **Command Mode**: Support for executing commands within the application
- **Kubernetes Integration**: Browse and manage Kubernetes Secrets for database credentials
- **Custom SQL Queries**: Execute custom SQL queries with results display
- **Table Statistics**: View database table size and statistics information

## Screenshots

### Kubernetes Configuration

<div align="center">
  <img src="assets/page_configk8s.png" alt="Kubernetes Configuration" width="1200" height="auto">
</div>

### Query Data Using SQL

<div align="center">
  <img src="assets/page_query.png" alt="Query Data" width="1200" height="auto">
</div>

## Usage

### Building From Source

p6s is currently using Go v1.21.X or above. In order to build p6s from source you must:

1. **Clone the repository**
   ```bash
   git clone https://github.com/your-username/p6s.git
   cd p6s
   ```

2. **Build and run the executable**
   ```bash
   # Build for current platform
   go build -o p6s cmd/p6s/main.go
   
   # Run the application
   ./p6s
   ```

3. **Cross-platform builds**
   ```bash
   # Build for Linux AMD64
   GOOS=linux GOARCH=amd64 go build -o p6s-amd64 cmd/p6s/main.go
   
   # Build for Windows
   GOOS=windows GOARCH=amd64 go build -o p6s.exe cmd/p6s/main.go
   
   # Build for macOS
   GOOS=darwin GOARCH=amd64 go build -o p6s-darwin cmd/p6s/main.go
   ```

### Basic Operations

- **Configure Connection**: After starting the application, configure database connection information (host, port, username, password, database name, SSL mode)
- **Switch Database**: Use `\c` command or menu options to switch to different databases
- **View Connections**: Main interface displays all connection information for the current database
- **Filter Connections**: Use menu to select different filter conditions (all connections, active connections, blocked connections)
- **Command Mode**: Press `:` to enter command mode for executing specific commands

### Keyboard Shortcuts

- `:` - Enter command line mode
- `\c` - Switch database
- `\config` - Configure connection information
- `\k8s` - Kubernetes commands

## Configuration File

Connection configuration information is saved in the `.p6s/config.json` file in the user's home directory, containing database connection settings and Kubernetes integration parameters:

```json
{
  "host": "",
  "port": "",
  "username": "",
  "password": "",
  "database": "",
  "sslmode": "",
  "namespace": "",
  "pod": "",
  "container": "",
  "port_name": "",
  "secret": "",
  "secret_key": ""
}
```

## Dependencies

- [github.com/gdamore/tcell/v2](https://github.com/gdamore/tcell) - Terminal interface library
- [github.com/rivo/tview](https://github.com/rivo/tview) - Terminal UI component library based on tcell
- [github.com/lib/pq](https://github.com/lib/pq) - PostgreSQL driver
- [k8s.io/client-go](https://github.com/kubernetes/client-go) - Kubernetes client library

## License

MIT License
