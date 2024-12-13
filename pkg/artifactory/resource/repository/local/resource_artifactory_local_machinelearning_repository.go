package local

import (
	"context"
	"reflect"

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
				Description:       "Provides a resource to creates a local Machine Learning repository.\n\nOfficial documentation can be found [here](https://jfrog.com/help/r/jfrog-artifactory-documentation/machine-learning-repositories).",
				PackageType:       repository.MachineLearningType,
				Rclass:            Rclass,
				ResourceModelType: reflect.TypeFor[LocalResourceModel](),
				APIModelType:      reflect.TypeFor[LocalAPIModel](),
			},
		},
	}
}

type MachineLearningLocalRepositoryResource struct {
	localResource
}

func (r *MachineLearningLocalRepositoryResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes:  LocalAttributes,
		Description: r.Description,
	}
}
