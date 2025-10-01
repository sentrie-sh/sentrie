package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/binaek/cling"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/loader"
	"github.com/sentrie-sh/sentrie/runtime"
	"github.com/sentrie-sh/sentrie/trinary"
)

func addExecCmd(cli *cling.CLI) {
	cli.WithCommand(
		cling.NewCommand("exec", execCmd).
			WithArgument(cling.NewStringCmdInput("rule").
				WithDescription("Rule to execute").
				AsArgument(),
			).
			WithFlag(cling.
				NewStringCmdInput("pack-location").
				WithDefault(".").
				WithDescription("Pack directory to load").
				AsFlag(),
			).
			WithFlag(cling.
				NewStringCmdInput("facts").
				WithDefault("{}").
				WithDescription("Facts to execute the rule with").
				AsFlag(),
			),
	)
}

type execCmdArgs struct {
	PackLocation string `cling-name:"pack-location"`
	Rule         string `cling-name:"rule"`
	Facts        string `cling-name:"facts"`
}

func execCmd(ctx context.Context, args []string) error {
	input := execCmdArgs{}
	if err := cling.Hydrate(ctx, args, &input); err != nil {
		return err
	}

	pack, err := loader.LoadPack(ctx, input.PackLocation)
	if err != nil {
		return err
	}

	idx := index.CreateIndex()

	if err := idx.SetPack(ctx, pack); err != nil {
		return err
	}

	programs, err := loader.LoadPrograms(ctx, pack)
	if err != nil {
		return err
	}

	for _, program := range programs {
		if err := idx.AddProgram(ctx, program); err != nil {
			return err
		}
	}

	if err := idx.Validate(ctx); err != nil {
		return err
	}

	exec, err := runtime.NewExecutor(idx)
	if err != nil {
		return err
	}

	facts := make(map[string]any)
	if err := json.Unmarshal([]byte(input.Facts), &facts); err != nil {
		return err
	}

	namespace, policy, rule, err := exec.Index().ResolveSegments(input.Rule)
	if err != nil {
		return err
	}

	var outputs []*runtime.ExecutorOutput
	var runErr error
	if len(rule) == 0 {
		outputs, runErr = exec.ExecPolicy(ctx, namespace, policy, facts)
	} else {
		output, err := exec.ExecRule(ctx, namespace, policy, rule, facts)
		outputs = []*runtime.ExecutorOutput{output}
		runErr = err
	}

	// now that we have the outputs, lets map it by namespace and policy

	if runErr != nil {
		return runErr
	}

	m := sortOutputs(outputs)
	formatOutput(m)

	return nil
}

type ExecutorOutputMap map[string]map[string]map[string]*runtime.ExecutorOutput

func (m ExecutorOutputMap) Format() {
	formatOutput(m)
}

func sortOutputs(outputs []*runtime.ExecutorOutput) ExecutorOutputMap {
	m := ExecutorOutputMap{}

	for _, output := range outputs {
		if _, ok := m[output.Namespace]; !ok {
			m[output.Namespace] = map[string]map[string]*runtime.ExecutorOutput{}
		}

		if _, ok := m[output.Namespace][output.PolicyName]; !ok {
			m[output.Namespace][output.PolicyName] = map[string]*runtime.ExecutorOutput{}
		}

		if _, ok := m[output.Namespace][output.PolicyName][output.RuleName]; !ok {
			m[output.Namespace][output.PolicyName][output.RuleName] = output
		}
	}

	return m
}

// formatOutput formats the decision output in the specified format
//
// Examples:
//
// Namespace: my_first_policy
// Policy:    user_access
// Rules:
//
//	✓ allow_admin: true
//	✓ allow_user: true
//
// Values:
//
//	✓ allow_admin: true
//	✓ allow_user: true
//
// Attachments:
//
//	 allow_admin:
//		  NAME: VALUE
//	 allow_user:
//		  NAME: VALUE
//		  NAME:
//		    listvalue1
//		    listvalue2
//		  NAME:
//		    mapKey1: mapValue1
//		    mapKey2: mapValue2
func formatOutput(m ExecutorOutputMap) {
	for namespace, policies := range m {
		fmt.Printf("Namespace: %s\n", namespace)
		for policyName, policyData := range policies {
			fmt.Printf("Policy:    %s\n", policyName)
			fmt.Println()
			fmt.Printf("Rules:     \n")
			for ruleName, ruleData := range policyData {
				fmt.Printf("  ✓ %s: %s\n", ruleName, formatDecision(ruleData.Decision))
			}
			fmt.Println()
			fmt.Printf("Values:    \n")
			for ruleName, ruleData := range policyData {
				fmt.Printf("  ✓ %s: %v\n", ruleName, ruleData.Decision.Value)
			}
			fmt.Println()

			numAttachments := 0
			for _, ruleData := range policyData {
				numAttachments += len(ruleData.Attachments)
			}

			if numAttachments > 0 {
				fmt.Printf("Attachments: \n")
				for ruleName, ruleData := range policyData {
					if len(ruleData.Attachments) == 0 {
						continue
					}
					fmt.Printf("  ✓ %s:\n", ruleName)
					for name, value := range ruleData.Attachments {
						formatAttachment(name, value)
					}
				}
				fmt.Println()
			}

		}
	}
}

// formatDecision formats the decision state with appropriate symbols
func formatDecision(decision any) string {
	if _, ok := decision.(trinary.HasTrinary); ok {
		return formatTrinaryState(decision)
	}

	if _, ok := decision.(trinary.Value); ok {
		return formatTrinaryState(decision)
	}

	return "• Unknown"
}

// formatTrinaryState formats trinary values with symbols
func formatTrinaryState(state any) string {
	switch state {
	case trinary.True:
		return "✓ True"
	case trinary.False:
		return "⨯ False"
	case trinary.Unknown:
		return "• Unknown"
	default:
		return formatTrinaryState(trinary.From(state))
	}
}

// formatAttachment formats attachment values with proper indentation
func formatAttachment(name string, value any) {
	if list, ok := value.([]any); ok {
		fmt.Printf("       %s:\n", name)
		for _, item := range list {
			fmt.Printf("        %s\n", item)
		}
		return
	}
	if m, ok := value.(map[string]any); ok {
		fmt.Printf("       %s:\n", name)
		for key, val := range m {
			fmt.Printf("        %s: %s\n", key, val)
		}
		return
	}
	fmt.Printf("     %s: %s\n", name, value)
}
