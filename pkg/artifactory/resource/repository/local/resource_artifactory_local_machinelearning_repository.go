package local

import (
	"context"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
)

func NewMachineLearningLocalRepositoryResource() resource.Resource {
	return &MachineLearningLocalRepositoryResource{
		localResource: NewLocalRepositoryResource(
			repository.MachineLearningType,
			"Machine Learning",
			reflect.TypeFor[LocalResourceModel](),
			reflect.TypeFor[LocalAPIModel](),
		),
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
