package runtime

import (
	"context"
	"fmt"
	"slices"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/runtime/trace"
	"go.opentelemetry.io/otel/attribute"
	oteltrace "go.opentelemetry.io/otel/trace"
)

func evalDistinct(ctx context.Context, ec *ExecutionContext, exec *executorImpl, p *index.Policy, d *ast.DistinctExpression) (any, *trace.Node, error) {
	node, done := trace.New("distinct", "", d, map[string]any{
		"collection": d.Collection.String(),
		"left_iter":  d.LeftIterator,
		"right_iter": d.RightIterator,
	})
	defer done()

	// Create OpenTelemetry span for JavaScript calls if tracing is enabled
	var span oteltrace.Span
	if cfg := ec.executor.OTelConfig(); cfg.Enabled && cfg.TraceExecution {
		ctx, span = ec.executor.Tracer().Start(ctx, "distinct")
		defer span.End()

		span.SetAttributes(
			attribute.String("sentrie.ast.node.kind", d.Kind()),
			attribute.String("sentrie.ast.node.range", d.Span().String()),
		)
	}

	col, colNode, err := eval(ctx, ec, exec, p, d.Collection)
	node.Attach(colNode)
	if err != nil {
		return nil, node.SetErr(err), err
	}

	list, ok := col.([]any)
	if !ok {
		return nil, node.SetErr(fmt.Errorf("distinct expects list source")), fmt.Errorf("distinct expects list source")
	}

	if len(list) < 2 {
		// nothing to do here
		return list, node, nil
	}

	// clone the list
	list = slices.Clone(list)

	theDistinct := make([]any, 0, len(list))
	theDistinct = append(theDistinct, list[0]) // start with the first item
	list = list[1:]

	// start with a distinct list of 1
	// for every item in the distinct list, iterate through the list with:
	// - the distinct item as left iterator
	// - the item as right iterator
	// - the predicate
	// if the predicate is truthy, the distinct item is the same as the item - continue to next item
	// if all predicates are falsey, add the item to the distinct list

	// iterate through the list
	for len(list) > 0 {
		// get the next item
		item := list[0]
		list = list[1:]
		foundMatch := false

		// now, iterate through the current known distinct items
		for _, distinctItem := range theDistinct {
			childContext := ec.AttachedChildContext()
			childContext.SetLocal(d.LeftIterator, distinctItem, true)
			childContext.SetLocal(d.RightIterator, item, true)
			res, resNode, err := eval(ctx, childContext, exec, p, d.Predicate)
			node.Attach(resNode)
			childContext.Dispose()
			if err != nil {
				return nil, node.SetErr(err), err
			}
			if IsTruthy(res) {
				foundMatch = true
				break
			}
		}

		// if no match was found, add the item to the distinct list
		if !foundMatch {
			theDistinct = append(theDistinct, item)
		}
	}

	theDistinct = slices.Clip(theDistinct)

	return theDistinct, node, nil
}
