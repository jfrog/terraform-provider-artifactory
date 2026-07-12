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

package user

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/samber/lo"
)

func NewManagedUserResource() resource.Resource {
	return &ArtifactoryManagedUserResource{
		ArtifactoryBaseUserResource: ArtifactoryBaseUserResource{
			TypeName: "artifactory_managed_user",
		},
	}
}

type ArtifactoryManagedUserResource struct {
	ArtifactoryBaseUserResource
}

func (r *ArtifactoryManagedUserResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	var managedUserSchemaFramework = map[string]schema.Attribute{
		"password": schema.StringAttribute{
			Required:            true,
			Sensitive:           true,
			MarkdownDescription: "Password for the user.",
		},
	}

	managedUserSchemaFramework = lo.Assign(baseUserSchemaFramework, managedUserSchemaFramework)

	resp.Schema = schema.Schema{
		MarkdownDescription: "Provides an Artifactory managed user resource. This can be used to create and manage Artifactory users. For example, service account where password is known and managed externally.",
		Attributes:          managedUserSchemaFramework,
		Version:             1,
	}
}
