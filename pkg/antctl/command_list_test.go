// Copyright 2019 Antrea Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package antctl

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	"antrea.io/antrea/pkg/agent/apiserver/handlers/multicast"
	fallbackversion "antrea.io/antrea/pkg/antctl/fallback/version"
	"antrea.io/antrea/pkg/antctl/raw/featuregates"
	"antrea.io/antrea/pkg/antctl/raw/proxy"
	"antrea.io/antrea/pkg/antctl/raw/supportbundle"
	"antrea.io/antrea/pkg/antctl/raw/traceflow"
	"antrea.io/antrea/pkg/antctl/runtime"
	"antrea.io/antrea/pkg/antctl/transform/version"
	systemv1beta1 "antrea.io/antrea/pkg/apis/system/v1beta1"
	controllerinforest "antrea.io/antrea/pkg/apiserver/registry/system/controllerinfo"
	"antrea.io/antrea/pkg/client/clientset/versioned/scheme"
)

type testResponse struct {
	Label string `json:"label" antctl:"key"`
	Value uint64 `json:"value"`
}

var testCommandList = &commandList{
	definitions: []commandDefinition{
		{
			use:                 "test",
			short:               "test short description ${component}",
			long:                "test description ${component}",
			transformedResponse: reflect.TypeOf(testResponse{}),
		},
	},
	codec: scheme.Codecs,
}

func TestCommandListApplyToCommand(t *testing.T) {
	testRoot := new(cobra.Command)
	testRoot.Short = "The component is ${component}"
	testRoot.Long = "The component is ${component}"
	testCommandList.ApplyToRootCommand(testRoot)
	// sub-commands should be attached
	assert.True(t, testRoot.HasSubCommands())
	// render should work as expected
	assert.Contains(t, testRoot.Short, fmt.Sprintf("The component is %s", runtime.Mode))
	assert.Contains(t, testRoot.Long, fmt.Sprintf("The component is %s", runtime.Mode))
}

var testCommandList2 = &commandList{
	definitions: []commandDefinition{
		{
			use:          "version",
			short:        "Print version information",
			long:         "Print version information of antctl and ${component}",
			commandGroup: flat,
			controllerEndpoint: &endpoint{
				resourceEndpoint: &resourceEndpoint{
					resourceName:         controllerinforest.ControllerInfoResourceName,
					groupVersionResource: &systemv1beta1.ControllerInfoVersionResource,
				},
				addonTransform: version.ControllerTransform,
				// print the antctl client version even if request to Controller fails
				requestErrorFallback: fallbackversion.RequestErrorFallback,
			},
			agentEndpoint: &endpoint{
				nonResourceEndpoint: &nonResourceEndpoint{
					path: "/version",
				},
				addonTransform: version.AgentTransform,
				// print the antctl client version even if request to Agent fails
				requestErrorFallback: fallbackversion.RequestErrorFallback,
			},
			flowAggregatorEndpoint: &endpoint{
				nonResourceEndpoint: &nonResourceEndpoint{
					path: "/version",
				},
				addonTransform: version.FlowAggregatorTransform,
				// print the antctl client version even if request to Flow Aggregator fails
				requestErrorFallback: fallbackversion.RequestErrorFallback,
			},
			transformedResponse: reflect.TypeOf(version.Response{}),
		},
		{
			use:   "podmulticaststats",
			short: "Show multicast statistics",
			long:  "Show multicast traffic statistics of Pods",
			example: `  Show multicast traffic statistics of all local Pods on the Node
$ antctl get podmulticaststats
Show multicast traffic statistics of a given Pod
$ antctl get podmulticaststats pod -n namespace`,
			commandGroup: get,
			agentEndpoint: &endpoint{
				nonResourceEndpoint: &nonResourceEndpoint{
					path:       "/podmulticaststats",
					outputType: multiple,
					params: []flagInfo{
						{
							name:  "name",
							usage: "Retrieve Pod Multicast Statistics by name. If present, Namespace must be provided.",
							arg:   true,
						},
						{
							name:      "namespace",
							usage:     "Get Pod Multicast Statistics from specific Namespace.",
							shorthand: "n",
						},
					},
				},
			},

			transformedResponse: reflect.TypeOf(multicast.Response{}),
		},
		{
			use:   "log-level",
			short: "Show or set log verbosity level",
			long:  "Show or set the log verbosity level of ${component}",
			example: `  Show the current log verbosity level
  $ antctl log-level
  Set the log verbosity level to 2
  $ antctl log-level 2`,
			commandGroup: flat,
			controllerEndpoint: &endpoint{
				nonResourceEndpoint: &nonResourceEndpoint{
					path: "/loglevel",
					params: []flagInfo{
						{
							name:  "level",
							usage: "The integer log verbosity level to set",
							arg:   true,
						},
					},
					outputType: single,
				},
			},
			agentEndpoint: &endpoint{
				nonResourceEndpoint: &nonResourceEndpoint{
					path: "/loglevel",
					params: []flagInfo{
						{
							name:  "level",
							usage: "The integer log verbosity level to set",
							arg:   true,
						},
					},
					outputType: single,
				},
			},
			flowAggregatorEndpoint: &endpoint{
				nonResourceEndpoint: &nonResourceEndpoint{
					path: "/loglevel",
					params: []flagInfo{
						{
							name:  "level",
							usage: "The integer log verbosity level to set",
							arg:   true,
						},
					},
					outputType: single,
				},
			},
			transformedResponse: reflect.TypeOf(0),
		},
	},
	rawCommands: []rawCommand{
		{
			cobraCommand:      supportbundle.Command,
			supportAgent:      true,
			supportController: true,
		},
		{
			cobraCommand:      traceflow.Command,
			supportAgent:      true,
			supportController: true,
		},
		{
			cobraCommand:      proxy.Command,
			supportAgent:      false,
			supportController: true,
		},
		{
			cobraCommand:      featuregates.Command,
			supportAgent:      true,
			supportController: true,
			commandGroup:      get,
		},
	},
	codec: scheme.Codecs,
}

func TestGetDebugCommands(t *testing.T) {

	tc := []struct {
		mode     string
		expected [][]string
	}{
		{
			mode:     "controller",
			expected: [][]string{{"version"}, {"supportbundle"}, {"traceflow"}, {"get", "featuregates"}},
		},
		{
			mode:     "agent",
			expected: [][]string{{"version"}, {"get", "podmulticaststats"}, {"log-level"}, {"supportbundle"}, {"traceflow"}, {"get", "featuregates"}},
		},
		{
			mode:     "flowaggregator",
			expected: [][]string{{"version"}, {"log-level"}},
		},
	}
	for _, tt := range tc {
		t.Run("Naman", func(t *testing.T) {
			generated := testCommandList2.GetDebugCommands(tt.mode)
			assert.Equal(t, tt.expected, generated)
		})
	}

}
