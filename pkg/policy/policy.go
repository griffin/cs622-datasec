package policy

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/griffin/cs622-datasec/pkg/user"
)

type PolicyHandler interface {
	// ApplyQueryPolicy checks to make sure that the query is following the
	// defined policies
	CheckPolicy(usr user.User, query string) error
}

type httpPolicyHandler struct {
	baseURL *url.URL
}

type policyResponse struct {
	Allow   bool
	Message string
}

func NewHttpPolicyHandler(url *url.URL) PolicyHandler {
	return &httpPolicyHandler{
		baseURL: url,
	}
}

func (h *httpPolicyHandler) CheckPolicy(usr user.User, query string) error {
	resp, err := http.PostForm(
		fmt.Sprintf("%v/enforce", h.baseURL.Host),
		url.Values{"sql": []string{query}, "user": []string{usr.PostgreUser}},
	)
	if err != nil {
		return err
	}

	r := policyResponse{}
	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		return err
	}

	if !r.Allow {
		return fmt.Errorf(r.Message)
	}

	return nil
}
