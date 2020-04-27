# CS622 Data Security and Privacy Project

## Running

Download the latest release: https://golang.org/dl/

Show help menu:
`go run cmd/api-server/main.go --help`

Normal Usage:
```
go mod vendor
go run cmd/api-server/main.go \
  --sql_username="postgres" \
  --sql_password=$SECRET" \
  --sql_database="postgres" \
  --sql_host="example.com" \
  --sql_port=5432 \
  --port=8080
```

## Organization of Code

The Go program in this repository implements the Authentication, Authorization
and Proxy modules of the system. Go will be the frontend for any queries coming
from external users and will be available in highly scalable manor as the
application persists no state. Root Postgres connections will be pooled here and
will be the most sensitive application in the infrastructure as it will have
permissions to assume all roles in the Postgres datastore, giving it very wide
permissions.

### cmd

Store any "main" files in this folder to separate library code from
specific front-ends for the process.

### Adding a New Route

Define HTTP routes to interact with the datastore and models as APIs for
consumption by external modules. No model code should be defined here and should
get it's own package.

Anything common to all routes such as authentication or authorization of routes
should be done as HTTP middleware. An example is the `authMiddeware` function
in the `cmd/api-server` package.

Adding a new module:

1. Create `New<ModuleName>Manager` with the needed datastore interfaces to interact
   with persistent storage. Pass any additional configuration in this function
   such as timeouts.
1. For each route you would like to create, make a function with the format
   `<HTTP Verb><ModuleName><Function>Route`, implementing the Gin handler
   interface. Group similar routes that require the same handlers.
1. Register each route in `cmd/api-server/main.go`, registering any required
   middleware such as authentication to be passed down via `context.Context`.

### Datastore Access

This package contains code to interact with databases. Code such as connecting
to databases and relevant abstractions should live here. Currently this package
supports connecting to Postgres databases.

### Modules

#### Routes

The routes package implement basic HTTP routes that we use as an interface to
interact with the API. A more human usable interface could be written on top
or a developer API for a specific language could be used to integrate with their
current codebase.

For test cases we've been interacting with this API using the `curl` utility for
querying HTTP routes and `jq` for making JSON responses more readable.

#### Users

This module manages the metadata that needs to be associated with a user of the
API. This includes a variety of information such as a username, authentication
information and database metadata.

Authentication is done using a user chosen password, that is hashed and salted
in our database using bcrypt.

We store the user's associated postgres role in this module as well. This field
cannot be edited by a user, but must be set by an administrative user. If this
field is blank they will be prompted to contact an administrator to do so.
Associated roles can be changed over time, leaving room for a system that
would dynamically scope credentials based on query behavior.

Lastly a user selector is associated with each user, so that we don't expose an
auto incrementing user id where an attacker can guess other users. An auto
incrementing id is still used for joins.

A user can register for an account using the `/v1/register` route, providing
a `name`, `email` and `passoword` encoded in a JSON object. An example of how
this could be done is shown below, assuming the service is running locally.

```bash
curl -H "Content-Type: application/json" \
--data '{"name": "test", "email": "test@example.com", "password": "qwerty123"}' \
localhost:8080/v1/register
```

A `200 OK` response indicates that the user was created successfully, if an
error was encountered a JSON object is returned with the error message. The
HTTP response code will be adjusted based on the exact error, but will be above
400.

#### Session

This module manages the activity a user does with the system, providint a with a
token which they would store securely. Using a token instead of a password each time reduces the
surface area of attack as you can easily expire tokens and associate a token
with specific login information like geospatial data.

Tokens are checked using HTTP middleware and transformed into the user domain
model for use by HTTP routes. Tokens are also invalidated after a period to
encourage credential rotation, it is currently set at 2 hours.

To login a user can send their `email` and `password` encoded as a JSON object
to the `/v1/login` route, which will return a JSON object with a `token` if
successful. An example can be shown below:

```bash
curl -H "Content-Type: application/json" \
--data '{"email": "test@example.com", "password": "qwerty123"}' \
localhost:8080/v1/login
```

When a user wants to authenticate a route they must provided the given token
using the `Authorization` HTTP header like so `Authorization: Bearer $TOKEN`.

When you would like to retire a token you can also logout like so:

```bash
curl -H "Content-Type: application/json" \
-H "Authorization: Bearer $TOKEN" \
--data '{"sql": "SELECT * FROM farms;"}' \
localhost:8080/v1/query
```

#### Query

The query module is the many API service offered by this application. It offers
the user the ability to query the database using the `/v1/query` HTTP route,
encoding the query in JSON. A user must provide a token obtained using the
session module in order to perform this request, along with having a valid
Postgres role associated with the user.

There are a few restrictions of the queries that are done through the API. First
you are not allowed to run any statements that modify roles, but instead must
be the basic SQL option such as `SELECT`, `INSERT` or `DELETE`. This is because
we assume the roles of each user to restrict the queries that can be done by
running `SET ROLE $ROLE;` then doing the user's query and lastly reset all in a
single SQL transaction. We enforce these constraints by parsing the AST and
verifying it against our criteria.

Results are returned to the user as a list of rows that match the query with the
value associated with the value:

```json
[
  [
    {
      "column": "farm_id",
      "value": 30
    },
    {
      "column": "name",
      "value": "Farm-30"
    },
    {
      "column": "bank_id",
      "value": "903415342571"
    }
  ]
]
```

This is achieved by running this example query:

```bash
curl -H "Content-Type: application/json" \
-H "Authorization: Bearer $TOKEN" \
--data '{"sql": "SELECT * FROM farms;"}' \
localhost:8080/v1/query
```

If the query is unauthorized a message will be returned with a human readable
error message and the appropriate HTTP status code.

#### Audit

The audit module is implemented using a two prong approach. We log every query
that passes through the API module in an audit table that includes the status
of the query, who ran it, what Postgres user was used and the query contents.

To provide additional inspection of what was done we have a trigger installed
in the database to gain additional information about the query on the database
side. This was done to save some time, but an optimal solution would do this at
the same location to better correlate this data together.

The results of the audit table can be queries by audit users through the same
`/v1/query` route. Additional modules could be built that use this api to
provide automatic notification os suspicious audit entries.

#### Policy

The policy module validates query metadata against a policy written in the rego
policy language. Policies can be written to guard a large number of things such
as disabling querying password columns across the board. Most of this metadata
comes from parsing the query itself, but additional metadata could easily be
added to the policy engine to extend functionality.

An example policy is shown below that doesn't allow you to make a query that
uses the `*` to select columns and doesn't allow you to select `bank_id`.

```rego
package sql

default allow = false

contains(arr, elem) {
    arr[_] = elem
}

allow {
    not contains(input.cols, "bank_id")
    input.star = false
}
```

This can be useful to prevent bad SQL practices like using the `*`, or to
prevent access to columns across the board. Currently the policies only have
access to the column names, tables names and if the query is using the `*`
operator. This could easily be extented to provide additional metadata such as
the amount of recent queries by the user or additional data from the SQL AST.

The policy is evaluated using [Open Policy Agent](https://www.openpolicyagent.org/),
which can also be deployed as a separate component, but it
is currently integrated into the api using the Golang API.
