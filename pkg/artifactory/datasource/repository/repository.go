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

package repository

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkv2_diag "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	sdkv2_schema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/util"
)

// Framework types and functions

// BaseRepositoryDataSourceModel contains common fields for all repository data sources
type BaseRepositoryDataSourceModel struct {
	Key                 types.String `tfsdk:"key"`
	ProjectKey          types.String `tfsdk:"project_key"`
	ProjectEnvironments types.Set    `tfsdk:"project_environments"`
	Description         types.String `tfsdk:"description"`
	Notes               types.String `tfsdk:"notes"`
	IncludesPattern     types.String `tfsdk:"includes_pattern"`
	ExcludesPattern     types.String `tfsdk:"excludes_pattern"`
	RepoLayoutRef       types.String `tfsdk:"repo_layout_ref"`
	PackageType         types.String `tfsdk:"package_type"`
}

// BaseRepositoryAPIModel contains common fields for all repository API models
// This is a helper type for conversion - actual API models embed resource repository.BaseAPIModel
type BaseRepositoryAPIModel struct {
	Key                 string
	ProjectKey          string
	ProjectEnvironments []string
	Description         string
	Notes               string
	IncludesPattern     string
	ExcludesPattern     string
	RepoLayoutRef       string
	PackageType         string
}

// BaseDataSourceAttributes defines the common attributes for all repository datasources
var BaseDataSourceAttributes = map[string]schema.Attribute{
	"key": schema.StringAttribute{
		Required:            true,
		MarkdownDescription: "A mandatory identifier for the repository that must be unique. Must be 1 - 64 alphanumeric and hyphen characters. It cannot contain spaces or special characters.",
	},
	"project_key": schema.StringAttribute{
		Computed:            true,
		MarkdownDescription: "Project key for assigning this repository to. Must be 2 - 32 lowercase alphanumeric and hyphen characters.",
	},
	"project_environments": schema.SetAttribute{
		ElementType:         types.StringType,
		Computed:            true,
		MarkdownDescription: "Project environments.",
	},
	"description": schema.StringAttribute{
		Computed:            true,
		MarkdownDescription: "Public description.",
	},
	"notes": schema.StringAttribute{
		Computed:            true,
		MarkdownDescription: "Internal description.",
	},
	"includes_pattern": schema.StringAttribute{
		Computed:            true,
		MarkdownDescription: "List of comma-separated artifact patterns to include when evaluating artifact requests.",
	},
	"excludes_pattern": schema.StringAttribute{
		Computed:            true,
		MarkdownDescription: "List of artifact patterns to exclude when evaluating artifact requests.",
	},
	"repo_layout_ref": schema.StringAttribute{
		Computed:            true,
		MarkdownDescription: "Sets the layout that the repository should use for storing and identifying modules.",
	},
	"package_type": schema.StringAttribute{
		Computed:            true,
		MarkdownDescription: "Package type.",
	},
}

// CommonFromAPIModel provides common conversion logic from API model to Terraform model for repositories
func CommonFromAPIModel(ctx context.Context, baseModel *BaseRepositoryDataSourceModel, apiModel BaseRepositoryAPIModel) diag.Diagnostics {
	diags := diag.Diagnostics{}

	// Base fields
	baseModel.Key = types.StringValue(apiModel.Key)
	baseModel.ProjectKey = types.StringValue(apiModel.ProjectKey)
	baseModel.Description = types.StringValue(apiModel.Description)
	baseModel.Notes = types.StringValue(apiModel.Notes)
	baseModel.IncludesPattern = types.StringValue(apiModel.IncludesPattern)
	baseModel.ExcludesPattern = types.StringValue(apiModel.ExcludesPattern)
	baseModel.RepoLayoutRef = types.StringValue(apiModel.RepoLayoutRef)
	baseModel.PackageType = types.StringValue(apiModel.PackageType)

	// Project environments
	var projectEnvironments []types.String
	for _, env := range apiModel.ProjectEnvironments {
		projectEnvironments = append(projectEnvironments, types.StringValue(env))
	}
	if len(projectEnvironments) > 0 {
		envSet, d := types.SetValueFrom(ctx, types.StringType, projectEnvironments)
		if d.HasError() {
			diags.Append(d...)
			return diags
		}
		baseModel.ProjectEnvironments = envSet
	} else {
		baseModel.ProjectEnvironments = types.SetNull(types.StringType)
	}

	return diags
}

// HexDataSourceAttributes defines the Hex-specific attributes for Hex repository datasources
var HexDataSourceAttributes = map[string]schema.Attribute{
	"hex_primary_keypair_ref": schema.StringAttribute{
		Computed:            true,
		MarkdownDescription: "Select the RSA key pair to sign and encrypt content for secure communication between Artifactory and the Mix client.",
	},
}

// HexRemoteDataSourceAttributes defines the Hex-specific attributes for remote Hex repository datasources
var HexRemoteDataSourceAttributes = map[string]schema.Attribute{
	"hex_primary_keypair_ref": schema.StringAttribute{
		Computed:            true,
		MarkdownDescription: "Select the RSA key pair to sign and encrypt content for secure communication between Artifactory and the Mix client.",
	},
	"public_key": schema.StringAttribute{
		Computed:            true,
		MarkdownDescription: "Contains the public key used when downloading packages from the Hex remote registry (public, private, or self-hosted Hex server).",
	},
}

// SDKv2 types and functions

var validRepositoryTypes = []string{"local", "remote", "virtual", "federated", "distribution"}
var validPackageTypes = []string{
	repository.AlpinePackageType,
	repository.BowerPackageType,
	repository.CargoPackageType,
	repository.ChefPackageType,
	repository.CocoapodsPackageType,
	repository.ComposerPackageType,
	repository.ConanPackageType,
	repository.CondaPackageType,
	repository.CranPackageType,
	repository.DebianPackageType,
	repository.DockerPackageType,
	repository.GemsPackageType,
	repository.GenericPackageType,
	repository.GitLFSPackageType,
	repository.GoPackageType,
	repository.GradlePackageType,
	repository.HelmPackageType,
	repository.HexPackageType,
	repository.HuggingFacePackageType,
	repository.IvyPackageType,
	repository.MavenPackageType,
	repository.NPMPackageType,
	repository.NugetPackageType,
	repository.OpkgPackageType,
	repository.P2PackageType,
	repository.PubPackageType,
	repository.PuppetPackageType,
	repository.PyPiPackageType,
	repository.RPMPackageType,
	repository.SBTPackageType,
	repository.SwiftPackageType,
	repository.TerraformPackageType,
	repository.TerraformBackendPackageType,
	repository.VagrantPackageType,
}

func MkRepoReadDataSource(pack packer.PackFunc, construct repository.Constructor) sdkv2_schema.ReadContextFunc {
	return func(ctx context.Context, d *sdkv2_schema.ResourceData, m interface{}) sdkv2_diag.Diagnostics {
		repo, err := construct()
		if err != nil {
			return sdkv2_diag.FromErr(err)
		}

		key := d.Get("key").(string)
		// repo must be a pointer
		resp, err := m.(util.ProviderMetadata).Client.R().
			SetResult(repo).
			SetPathParam("key", key).
			Get(repository.RepositoriesEndpoint)

		if err != nil {
			return sdkv2_diag.FromErr(err)
		}

		if resp.StatusCode() == http.StatusBadRequest || resp.StatusCode() == http.StatusNotFound {
			d.SetId("")
			return nil
		}

		if resp.IsError() {
			return sdkv2_diag.Errorf("%s", resp.String())
		}

		d.SetId(key)

		return sdkv2_diag.FromErr(pack(repo, d))
	}
}
