# InstantGate API

**InstantGate** turns any relational database (MySQL/PostgreSQL) into a fully functional REST API in seconds. Stop writing repetitive CRUD code for every table.

## Features

- **Auto CRUD**: GET, POST, PATCH, DELETE endpoints for all tables
- **Advanced Filtering**: Operators like `eq`, `gt`, `like`, `in`
- **Pagination**: `limit` and `offset` support
- **Security**: JWT authentication, table access control
- **Performance**: Written in Go with Redis caching
- **Docker Ready**: Single-command deployment

## Quick Start

```bash
# Download dependencies
go mod download

# Edit config/config.yaml with your database credentials

# Run the API
go run cmd/instantgate/main.go -config config/config.yaml

# Or build and run
go build -o bin/instantgate.exe cmd/instantgate/main.go
./bin/instantgate.exe -config config/config.yaml
```

API runs at `http://localhost:8080`

### Docker

```bash
docker-compose -f deployments/docker-compose.yml up
```

## API Endpoints

```bash
# Health check
curl http://localhost:8080/health

# List all tables
curl http://localhost:8080/api/schema

# Get table data (replace :table with actual table name)
curl http://localhost:8080/api/:table

# With filters
curl "http://localhost:8080/api/:table?status=active&age=gt.18"

# Get single record
curl http://localhost:8080/api/:table/:id

# Create record
curl -X POST http://localhost:8080/api/:table \
  -H "Content-Type: application/json" \
  -d '{"field":"value"}'

# Update record
curl -X PATCH http://localhost:8080/api/:table/:id \
  -H "Content-Type: application/json" \
  -d '{"field":"newvalue"}'

# Delete record
curl -X DELETE http://localhost:8080/api/:table/:id
```

## Filter Operators

| Operator | Description | Example |
|----------|-------------|---------|
| `eq` | Equal | `?status=active` |
| `ne` | Not equal | `?status=ne.inactive` |
| `gt` | Greater than | `?age=gt.18` |
| `gte` | Greater or equal | `?age=gte.18` |
| `lt` | Less than | `?price=lt.100` |
| `lte` | Less or equal | `?price=lte.100` |
| `like` | LIKE pattern | `?name=like.%john%` |
| `in` | IN list | `?status=in.active,pending` |
| `nin` | NOT IN list | `?status=nin.deleted` |

## Configuration

Edit `config/config.yaml`:

```yaml
server:
  port: 8080

database:
  driver: mysql
  host: localhost
  port: 3306
  name: mydb
  user: root
  password: ""

jwt:
  secret: change-me-in-production
  expiry: 24h

security:
  enabled: true
  require_auth: false
  whitelist: []  # Empty = allow all tables
  blacklist: []  # Tables to block

redis:
  host: localhost
  port: 6379
  cache_ttl: 5m
```

## Security

### Table Access Control

Allow only specific tables:
```yaml
security:
  whitelist: ["users", "products", "orders"]
```

Block specific tables:
```yaml
security:
  blacklist: ["admin_users", "secrets"]
```

### SQL Injection Protection

All queries use prepared statements. User input is never concatenated into SQL.

## Project Structure

```
cmd/instantgate/main.go          # Entry point
internal/
  api/                            # HTTP router, handlers, middleware
  database/mysql/                 # MySQL driver
  query/                          # SQL builder, filters
  schema/                         # Schema cache
  security/                       # JWT, access control
config/config.yaml                # Configuration
test.html                         # API test UI
```

## Test UI

Open `test.html` in your browser to explore and test all endpoints.

## License

MIT License
