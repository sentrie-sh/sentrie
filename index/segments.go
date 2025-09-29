package index

import (
	"strings"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/xerr"
)

func (idx *Index) ResolveSegments(path string) (ns, policy, rule string, err error) {
	// split by .
	parts := strings.Split(path, "/")
	// start joining the parts, until we have a namespace, or we run out of parts

	// Handle empty path case - check if all parts are empty
	allEmpty := true
	for _, part := range parts {
		if part != "" {
			allEmpty = false
			break
		}
	}
	if allEmpty {
		return "", "", "", xerr.ErrNamespaceNotFound(path)
	}

	// Try to find the longest possible namespace by building it greedily
	nsName := ""
	foundNamespace := false
	lastValidNamespace := ""
	lastValidParts := parts

	for i := 0; i < len(parts); i++ {
		nextPart := parts[i]

		if nextPart == "" {
			continue
		}

		if len(nsName) == 0 {
			nsName = nextPart
		} else {
			nsName = strings.Join([]string{nsName, nextPart}, ast.FQNSeparator)
		}

		n, err := idx.ResolveNamespace(nsName)
		if err == nil && n != nil {
			// Found a namespace, remember it but continue to see if we can find a longer one
			foundNamespace = true
			lastValidNamespace = n.FQN.String()
			lastValidParts = parts[i+1:]
		}
	}

	if !foundNamespace {
		return "", "", "", xerr.ErrNamespaceNotFound(path)
	}

	// Use the longest namespace we found
	nsName = lastValidNamespace
	parts = lastValidParts

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
