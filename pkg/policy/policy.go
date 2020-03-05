package policy

import (
	"github.com/griffin/cs622-datasec/pkg/user"
)

type PolicyHandler interface {

	// ApplyQueryPolicy applies policy to a query before it's executed
	ApplyQueryPolicy(usr user.User, query string) (string, error)

	// ApplyResultPolicy applicy policy to the result of a query
	ApplyResultPolicy(usr user.User, data map[string]interface{}) (map[string]interface{}, error)
}
