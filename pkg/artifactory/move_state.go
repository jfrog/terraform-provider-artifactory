// Copyright (c) JFrog Ltd. (2025)
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

package artifactory

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// MoveStateMixin provides a generic ResourceWithMoveState implementation that
// can be embedded into any Framework resource to enable cross-provider state
// moves for the same resource type.
//
// This is needed because OpenTofu 1.12.4+ correctly passes the source provider
// address during resource moves, which triggers the MoveResourceState RPC
// whenever the provider namespace changes (e.g., from jfrog/artifactory to
// hashicorp/artifactory in acceptance tests, or when a user switches registry
// host or migrates between Terraform and OpenTofu). Without an implementation
// of the ResourceWithMoveState interface, the framework returns an "Unable to
// Move Resource State" error.
//
// Because the source and target are the same resource type (only the provider
// address differs), no data transformation is required: the source state is
// passed through unchanged. The framework pre-populates the target schema in
// resp.TargetState.Schema, so the mover only needs to unmarshal the source raw
// state against that schema and set it as the target state.
type MoveStateMixin struct{}

func (m MoveStateMixin) MoveState(ctx context.Context) []resource.StateMover {
	return []resource.StateMover{
		{
			StateMover: func(ctx context.Context, req resource.MoveStateRequest, resp *resource.MoveStateResponse) {
				// IgnoreUndefinedAttributes tolerates attributes present in the
				// source state that no longer exist in the target schema, which
				// allows additive/removal schema changes between provider
				// versions without breaking the move.
				rawStateValue, err := req.SourceRawState.UnmarshalWithOpts(
					resp.TargetState.Schema.Type().TerraformType(ctx),
					tfprotov6.UnmarshalOpts{
						ValueFromJSONOpts: tftypes.ValueFromJSONOpts{
							IgnoreUndefinedAttributes: true,
						},
					},
				)
				// Returning without setting TargetState signals the framework to
				// skip this mover (e.g., when the source schema is genuinely
				// incompatible) rather than corrupting state.
				if err != nil {
					return
				}

				resp.TargetState.Raw = rawStateValue
			},
		},
	}
}
