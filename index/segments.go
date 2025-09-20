package index

import (
	"strings"

	"github.com/binaek/sentra/ast"
	"github.com/binaek/sentra/xerr"
	"github.com/pkg/errors"
)

func (idx *Index) ResolveSegments(path string) (ns, policy, rule string, err error) {
	// split by .
	parts := strings.Split(path, "/")
	// start joining the parts, until we have a namespace, or we run out of parts

	nsName := ""
	for {
		nextPart := parts[0]
		parts = parts[1:]

		if nextPart == "" {
			continue
		}
		if len(nsName) == 0 {
			nsName = nextPart
		} else {
			nsName = strings.Join([]string{nsName, nextPart}, ast.FQNSeparator)
		}
		n, err := idx.ResolveNamespace(nsName)
		if errors.Is(err, xerr.NotFoundError{}) {
			if len(parts) == 0 {
				// if we have no more parts, and we still don't have a namespace, return an error
				return "", "", "", xerr.ErrNamespaceNotFound(path)
			}
			continue
		}

		// if we have an error, and it's not a namespace not found error, return the error
		if err != nil {
			return "", "", "", err
		}

		if n != nil {
			nsName = n.FQN.String()
			break
		}
		if len(parts) == 0 {
			return "", "", "", xerr.ErrNamespaceNotFound(path)
		}
	}

	// if we do not have at least 1 part left, return an error - it's a problem - we MUST have a policy name
	if len(parts) == 0 {
		return "", "", "", xerr.ErrPolicyNotFound(path)
	}

	// we have a namespace, the next segment is the policy name
	policyName, parts := parts[0], parts[1:]
	_, err = idx.ResolvePolicy(nsName, policyName)
	if err != nil {
		return "", "", "", err
	}

	// we have a policy, the next segment is the rule name
	ruleName := ""

	if len(parts) > 0 {
		ruleName = parts[0]
	}

	return nsName, policyName, ruleName, nil
}
