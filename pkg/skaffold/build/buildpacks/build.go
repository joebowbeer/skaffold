/*
Copyright 2019 The Skaffold Authors

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

package buildpacks

import (
	"context"
	"fmt"
	"io"

	latest_v1 "github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/latest/v1"
)

// Build builds an artifact with Cloud Native Buildpacks:
// https://buildpacks.io/
func (b *Builder) Build(ctx context.Context, out io.Writer, artifact *latest_v1.Artifact, tag string) (string, error) {
	built, err := b.build(ctx, out, artifact, tag)
	if err != nil {
		return "", err
	}

	if err := b.localDocker.Tag(ctx, built, tag); err != nil {
		return "", fmt.Errorf("tagging %s->%q: %w", built, tag, err)
	}

	if b.pushImages {
		return b.localDocker.Push(ctx, out, tag)
	}
	return b.localDocker.ImageID(ctx, tag)
}
