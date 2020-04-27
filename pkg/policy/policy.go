package policy

import (
	"context"
	"fmt"

	"github.com/griffin/cs622-datasec/pkg/user"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/xwb1989/sqlparser"
)

type PolicyHandler interface {
	// ApplyQueryPolicy checks to make sure that the query is following the
	// defined policies
	CheckPolicy(usr user.User, query string) error
}

type httpPolicyHandler struct {
	policyFileContent string
}

func NewHttpPolicyHandler(policyFileContent string) PolicyHandler {
	return &httpPolicyHandler{
		policyFileContent: policyFileContent,
	}
}

func (h *httpPolicyHandler) CheckPolicy(usr user.User, query string) error {
	res, err := sqlparser.Parse(query)
	if err != nil {
		return err
	}

	isStar := false
	var cols []string
	var tables []string

	switch stmt := res.(type) {
	case *sqlparser.Select:

		for _, table := range stmt.From {
			switch tableTy := table.(type) {
			case *sqlparser.AliasedTableExpr:
				tables = append(tables, sqlparser.GetTableName(tableTy.Expr).CompliantName())
			}
		}

		for _, selectCol := range stmt.SelectExprs {
			switch sel := selectCol.(type) {
			case *sqlparser.StarExpr:
				isStar = true
			case *sqlparser.AliasedExpr:
				switch name := sel.Expr.(type) {
				case *sqlparser.ColName:
					cols = append(cols, name.Name.CompliantName())
				default:
					return fmt.Errorf("not handled case")
				}
			}
		}

		break
	}

	metadata := map[string]interface{}{
		"star":   isStar,
		"cols":   cols,
		"tables": tables,
	}

	compiler, err := ast.CompileModules(map[string]string{
		"sql.rego": h.policyFileContent,
	})
	if err != nil {
		return err
	}

	r, err := rego.New(
		rego.Input(metadata),
		rego.Query("data.sql.allow"),
		rego.Compiler(compiler),
	).PrepareForEval(context.Background())

	rs, err := r.Eval(context.Background())
	if err != nil {
		return err
	}

	resBool := rs[0].Expressions[0].Value.(bool)
	if !resBool {
		return fmt.Errorf("failed policy")
	}

	return nil
}
