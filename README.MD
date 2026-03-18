# MoraLink GOst

A lightweight Windows/Linux background service written in Go that acts as the local data agent for the MoraLink / SharkBusiness ecosystem.

It runs at system startup — before any user login — connects to the MoraLink server via a persistent gRPC tunnel, and handles database queries, data sync, remote commands, and self-updates. No UI, no interaction required.

---

## How It Works

```
MoraLink Server (remote)
        │
        │  gRPC bidirectional stream
        ▼
MoraLink GOst (this service, running locally)
        │
        ├── DB Managers  →  MySQL / PostgreSQL / Firebird / MSSQL
        └── API Managers →  FrontSys / (other integrations)
```

On startup, GOst connects to the MoraLink server and sends a `HELLO` message. The server responds with an `ACK` containing the client's connection configuration (DB credentials or API config). From that point, the server can issue `QUERY` messages to pull data from the local database, and GOst streams results back in batches.

### Message Flow

| Type | Direction | Description |
|------|-----------|-------------|
| `HELLO` | GOst → Server | Initial handshake with agent ID and version |
| `ACK` | Server → GOst | Approved/denied, includes DB/API connection config |
| `HEARTBEAT` | Server → GOst | Keepalive ping |
| `QUERY` | Server → GOst | Request data for a specific table |
| `RESULT` | GOst → Server | Batch of rows in response to a query |
| `COMMAND` | Server → GOst | Remote control (restart, reconnect DB, update) |
| `TRACE` | GOst → Server | Step-by-step execution tracing |
| `ERROR` | GOst → Server | Error reporting |

If the connection drops, a guardian goroutine automatically retries every minute.

---

## Supported Tables

| ID | Table | Description |
|----|-------|-------------|
| 0 | `produtos` | Products |
| 1 | `clientes` | Customers |
| 2 | `categorias` | Categories |
| 3 | `vendas` | Sales |
| 4 | `vendedores` | Sellers |
| 5 | `financeiro` | Financial records |
| 6 | `generic` | Arbitrary queries |

---

## Supported Database Drivers

| Type | Driver |
|------|--------|
| MySQL | `go-sql-driver/mysql` |
| MySQL (legacy) | `go-sql-driver/mysql` (old auth mode) |
| PostgreSQL | `lib/pq` |
| Firebird | `nakagami/firebirdsql` |
| MSSQL | `microsoft/go-mssqldb` |

For clients using external APIs instead of a local DB, the **API Managers** handle fetching, field mapping, and transcription (see [API Managers](#api-managers) below).

---

## Project Structure

```
moralink-gost/
├── main.go                        # Entry point, service lifecycle
├── modules/
│   ├── service/                   # Windows/Linux service wrappers
│   ├── grpc/
│   │   └── grpc_client.go         # gRPC stream client + message handler
│   ├── db_managers/               # One file per DB driver
│   │   ├── selector.go            # Routes to the correct driver
│   │   ├── mysql.go
│   │   ├── postgresql.go
│   │   ├── firebird.go
│   │   └── mssql.go
│   ├── api_managers/              # External API integrations
│   │   ├── transcriptor.go        # Field mapper engine (see MAPPER_DSL.md)
│   │   ├── selector.go
|   |   ├── MAPPER_DSL.md          # Field Mapper DSL reference
│   │   └── frontsys.go            # FrontSys integration
│   ├── command_managers/
│   │   └── organizer.go           # Handles COMMAND messages from server
│   ├── updater/
│   │   └── github.go              # Self-update via GitHub Releases
│   ├── logger/
│   │   └── logger.go              # Leveled logger with log rotation
│   ├── proto/
│   │   └── agentpb/               # Generated protobuf types
│   └── utils/
│       ├── util.go                # Shared types (ProdutoRow, VendaRow, etc.)
│       └── config.go              # Viper config loader
├── build_assets/
│   └── installer.nsi              # NSIS Windows installer script
└── Makefile
```

---

## Configuration

On first run, GOst creates a default `config.yaml` in the appropriate system config directory:

| OS | Default path |
|----|-------------|
| Windows | `%PROGRAMDATA%\moralink-gost\config.yaml` |
| Linux | `/etc/moralink-gost/config.yaml` |
| Override | Set `MORALINK_CONFIG_DIR` environment variable |

**Default `config.yaml`:**

```yaml
api:
  token: <your-token>
  user: <your-user>
  mode: prod        # set to "dev" to enable debug logging

server:
  grpc_port: 50051
```

The token identifies this agent to the MoraLink server and is how the server looks up which DB to connect to.

---

## Service Commands

Run as Administrator (Windows) or with `sudo` (Linux):

```bash
moralink-gost install    # Register as a system service
moralink-gost uninstall  # Remove the service
moralink-gost start      # Start the service
moralink-gost stop       # Stop the service
moralink-gost restart    # Restart the service
moralink-gost status     # Print current status
moralink-gost            # Run interactively (no service, useful for debugging)
```

The service is configured with `DelayedAutoStart` and automatic restart on failure (5s delay).

---

## Building

Prerequisites: Go 1.24+, `gcc` (for CGO), `x86_64-w64-mingw32-gcc` (for Windows cross-compilation).

```bash
# Linux binary
make linux

# Windows binary + NSIS installer
make windows
make installer

# Both, then sign everything (requires CERT_PATH and CERT_PASSWORD in .env)
make release
```

Build variables are injected at compile time:

```makefile
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.ReleaseGH=$(RELEASE_GH) -s -w"
```

`RELEASE_GH` is the GitHub token used by the self-updater at runtime.

---

## Self-Update

GOst can update itself remotely via a `COMMAND_UPDATE_APP` message from the server. The update flow:

1. Server sends `UPDATE_APP` command with a target version tag (or `"latest"`)
2. GOst calls the GitHub Releases API and downloads the matching binary asset
3. A detached process is spawned that waits 3 seconds, swaps the binary, and restarts the service
4. GOst exits cleanly — the detached process handles the rest

Platform-specific asset selection is automatic: the updater picks the asset whose filename contains both the current OS and architecture.

---

## API Managers

For clients whose data source is an external API rather than a local database, GOst uses the **API Managers** + **Field Mapper** instead of the DB Managers.

Each integration (e.g. FrontSys) implements the same `QueriesFunctions` interface as the DB drivers, so the rest of the system treats them identically.

The field mapping between the external API's response format and MoraLink's internal schema is defined in a JSON config stored in the server — no code changes needed to onboard a new integration. See **[MAPPER_DSL.md](./MAPPER_DSL.md)** for the full reference.

---

## Logging

Logs are written to `<config-dir>/logs/moralink-gost.log` with automatic rotation:

- Max size: 10 MB per file
- Max backups: 5
- Max age: 30 days
- Compression: enabled

When running interactively (not as a service), logs are also printed to stdout.

Log levels: `[INFO]`, `[WARN]`, `[ERROR]`, `[DEBUG]` — debug output is suppressed unless `api.mode` is set to `dev`.

---

## Development

```bash
# Run interactively with debug mode
dev=1 go run .

# pprof profiling is available at http://localhost:6060/debug/pprof/ when dev=1
```

---

## Publisher

**ORBIS SOLUÇÕES TECNOLÓGICAS**