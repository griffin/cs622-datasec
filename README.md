# CS622 Data Security and Privacy Project

## Running

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

## Organization of Code

The Go programs in this repository implement the Authentication, Authorization
and Proxy modules of the system. Go will be the frontend for any queries coming
from external users and will be highly scalable. Postgres connections will be
pooled here and will be the most sensative application in the infrastructure as
it will have permissions to create any role in the Postges server, so it will
have pretty wide permissions.

*Upcoming:*
The Auditing and Policy enforement modules will be implemented in NodeJS and be
consumed by the Go API server.

### cmd

Store any executables/main files in this folder to seperate library code from
specific frontends for the proccess.

### Routes

Define HTTP routes to interact with the datastore and models as APIs for
consumption by external modules. No model code should be defined here and should
get it's own package. Implement common functions as middleware and pass as Gin
context.

Adding a new module:
1. Create <ModuleName>Manager with the needed Datastore interfaces with interactions
   with persistant storage. This will store any other shared fields between
   other routes of this type such as configuration.
1. Create each route in this format <HTTP Verb><ModuleName><Function>Route to
   keep it consistant, it must also be implementing the function for the
   <ModuleName>Manger type.
1. Register each route in `cmd/api-server/main.go`, add middleware such as
   authentication to recieve user information

### Datastore

This package contains code to interact with databases. Code such as connecting
to databases and relevant abstractions should live here. Currently this package
supports connecting to Postgres databases.

### Utility

Store any common Utility functions here

### Modules

#### Users

Modules to represent users that can query the database along with relevant
metadata such as the Postgres user to associate the account with and login
credentials. User passwords are hashed and salted using bcrypt. Outside the
system the user is identified by a selector which is a random string that is
unique to the user, this is to prevent people from infering potential
permissions of a user.

#### Session

Implements user sessions so that a we don't need to store the user password, is
refreshed periodotically and can track the amount of logins and length of usage by a user.

#### Policy

Validates policy against queries and the results of queries. Policys will be
defined in a generic format with access to a varity of parameters.

#### Audit

Sends query information and status to the Audit module to log information to the
Postgres audit table for later querying.

#### Validation

- Check for SQL injection
- Common SQL antipatterns such as `SELECT *`
- To be expanded
