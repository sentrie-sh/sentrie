// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2025 Binaek Sarkar
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"strings"

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
				NewStringCmdInput("output").
				WithDefault("table").
				WithValidator(cling.NewEnumValidator("table", "json")).
				WithDescription("Output format to use. One of: table, json").
				AsFlag(),
			).
			WithFlag(cling.
				NewStringCmdInput("fact-file").
				WithDefault("").
				WithDescription("File to load facts from").
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
	FactFile     string `cling-name:"fact-file"`
	Output       string `cling-name:"output"`
}

func execCmd(ctx context.Context, args []string) error {
	input := execCmdArgs{}
	if err := cling.Hydrate(ctx, args, &input); err != nil {
		return err
	}

	factFileMap := make(map[string]any)
	// if the fact file is provided, load the facts from the file
	if input.FactFile != "" {
		content, err := os.ReadFile(input.FactFile)
		if err != nil {
			return err
		}
		decoder := json.NewDecoder(bytes.NewReader(content))
		if err := decoder.Decode(&factFileMap); err != nil {
			return err
		}
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

	var factFlagMap map[string]any
	decoder := json.NewDecoder(bytes.NewReader([]byte(input.Facts)))
	if err := decoder.Decode(&factFlagMap); err != nil {
		return err
	}

	facts := make(map[string]any)

	// merge in the values from the different sources
	maps.Copy(facts, factFileMap)
	maps.Copy(facts, factFlagMap)

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

	if input.Output == "json" {
		formatOutputJSON(outputs)
	} else {
		formatOutputTable(outputs)
	}

	return nil
}

type ExecutorOutputMap map[string]map[string]map[string]*runtime.ExecutorOutput

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

func formatOutputJSON(m []*runtime.ExecutorOutput) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(m)
}

// formatOutputTable formats the decision output in the specified format
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
func formatOutputTable(x []*runtime.ExecutorOutput) {
	m := sortOutputs(x)
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
						formatAttachment(name, value, 0)
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
func formatAttachment(name string, value any, indent int) {
	indentStr := strings.Repeat(" ", indent)
	if list, ok := value.([]any); ok {
		fmt.Printf("%s     %s:\n", indentStr, name)
		for _, item := range list {
			fmt.Printf("%s      - %v\n", indentStr, item)
		}
		return
	}

	if m, ok := value.(map[string]any); ok {
		fmt.Printf("%s     %s:\n", indentStr, name)
		for key, val := range m {
			formatAttachment(key, val, indent+1)
		}
		return
	}

	fmt.Printf("%s     %s: %v\n", indentStr, name, value)
}
