/*
Copyright 2020 The Skaffold Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package tags

import (
	"testing"

	latest_v1 "github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/latest/v1"
	"github.com/GoogleContainerTools/skaffold/testutil"
)

func TestSetAbsFilePaths(t *testing.T) {
	tests := []struct {
		description string
		config      *latest_v1.SkaffoldConfig
		base        string
		expected    *latest_v1.SkaffoldConfig
	}{
		{
			description: "relative path",
			config: &latest_v1.SkaffoldConfig{
				Pipeline: latest_v1.Pipeline{
					Build: latest_v1.BuildConfig{
						Artifacts: []*latest_v1.Artifact{
							{ImageName: "foo1", Workspace: "foo"},
							{ImageName: "foo2", Workspace: `C:\a\foo`},
						},
					},
					Deploy: latest_v1.DeployConfig{
						DeployType: latest_v1.DeployType{
							KptDeploy:     &latest_v1.KptDeploy{Dir: "."},
							KubectlDeploy: &latest_v1.KubectlDeploy{Manifests: []string{`foo\*`, `C:\a\foo\*`}},
							HelmDeploy: &latest_v1.HelmDeploy{Releases: []latest_v1.HelmRelease{
								{ChartPath: `..\charts`, ValuesFiles: []string{"values1.yaml", "values2.yaml"}, SetFiles: map[string]string{"envFile": "values3.yaml", "configFile": "values4.yaml", "anotherFile": `C:\c\values5.yaml`}},
								{RemoteChart: "foo/bar", ValuesFiles: []string{"values1.yaml", "values2.yaml"}, SetFiles: map[string]string{"envFile": "values3.yaml", "configFile": "values4.yaml", "anotherFile": `C:\c\values5.yaml`}},
							}},
						},
					},
				},
			},
			base: `C:\a\b`,
			expected: &latest_v1.SkaffoldConfig{
				Pipeline: latest_v1.Pipeline{
					Build: latest_v1.BuildConfig{
						Artifacts: []*latest_v1.Artifact{
							{ImageName: "foo1", Workspace: `C:\a\b\foo`},
							{ImageName: "foo2", Workspace: `C:\a\foo`},
						},
					},
					Deploy: latest_v1.DeployConfig{
						DeployType: latest_v1.DeployType{
							KptDeploy:     &latest_v1.KptDeploy{Dir: `C:\a\b`},
							KubectlDeploy: &latest_v1.KubectlDeploy{Manifests: []string{`C:\a\b\foo\*`, `C:\a\foo\*`}},
							HelmDeploy: &latest_v1.HelmDeploy{Releases: []latest_v1.HelmRelease{
								{ChartPath: `C:\a\charts`, ValuesFiles: []string{`C:\a\b\values1.yaml`, `C:\a\b\values2.yaml`}, SetFiles: map[string]string{"envFile": `C:\a\b\values3.yaml`, "configFile": `C:\a\b\values4.yaml`, "anotherFile": `C:\c\values5.yaml`}},
								{RemoteChart: "foo/bar", ValuesFiles: []string{`C:\a\b\values1.yaml`, `C:\a\b\values2.yaml`}, SetFiles: map[string]string{"envFile": `C:\a\b\values3.yaml`, "configFile": `C:\a\b\values4.yaml`, "anotherFile": `C:\c\values5.yaml`}},
							}},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		testutil.Run(t, test.description, func(t *testutil.T) {
			err := MakeFilePathsAbsolute(test.config, test.base)
			t.CheckNoError(err)
			t.CheckDeepEqual(test.expected, test.config)
		})
	}
}
