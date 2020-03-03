# CS622 Data Security and Privacy Project

# Running

Download the latest release: https://golang.org/dl/

Show help menu:
`go run cmd/api-server/main.go --help`

Normal Usage:
```
go mod vendor
go run cmd/api-server/main.go \
  --sql_username="root" \
  --sql_password="SECRET" \
  --sql_database="cashew" \
  --sql_host="example.com" \
  --sql_port=5432 \
  --port=8080
```

# Organization

### Routes

Define HTTP routes to interact with the datastore and models. Register new
rouets in `cmd/api-server/main.go`.

### Datastore

Anything releated to connecting to different types of databases and infrastructure.
