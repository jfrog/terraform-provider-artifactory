package local

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/util"
)

func NewMachineLearningLocalRepositoryResource() resource.Resource {
	return &MachineLearningLocalRepositoryResource{
		localResource: localResource{
			BaseResource: repository.BaseResource{
				JFrogResource: util.JFrogResource{
					TypeName:                "artifactory_local_machinelearning_repository",
					ValidArtifactoryVersion: "7.102.0",
					CollectionEndpoint:      "artifactory/api/repositories",
					DocumentEndpoint:        "artifactory/api/repositories/{key}",
				},
				Description: "Provides a resource to creates a local Machine Learning repository.",
				PackageType: repository.MachineLearningType,
				Rclass:      Rclass,
			},
		},
	}
}

type MachineLearningLocalRepositoryResource struct {
	localResource
}

type MachineLearningLocalRepositoryResourceModel struct {
	LocalResourceModel
}

func (r *MachineLearningLocalRepositoryResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes:  LocalAttributes,
		Description: r.Description,
	}
}
