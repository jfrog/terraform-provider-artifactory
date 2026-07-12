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

package artifact

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jfrog/terraform-provider-shared/util"
)

func NewFileListDataSource() datasource.DataSource {
	return &FileListDataSource{
		TypeName: "artifactory_file_list",
	}
}

type FileListDataSource struct {
	ProviderData util.ProviderMetadata
	TypeName     string
}

type FileListDataSourceModel struct {
	RepositoryKey      types.String `tfsdk:"repository_key"`
	FolderPath         types.String `tfsdk:"folder_path"`
	DeepListing        types.Bool   `tfsdk:"deep_listing"`
	Depth              types.Int64  `tfsdk:"depth"`
	ListFolders        types.Bool   `tfsdk:"list_folders"`
	MetadataTimestamps types.Bool   `tfsdk:"metadata_timestamps"`
	IncludeRootPath    types.Bool   `tfsdk:"include_root_path"`
	Uri                types.String `tfsdk:"uri"`
	Created            types.String `tfsdk:"created"`
	Files              types.List   `tfsdk:"files"`
}

var filesAttrType = map[string]attr.Type{
	"uri":           types.StringType,
	"size":          types.Int64Type,
	"last_modified": types.StringType,
	"folder":        types.BoolType,
	"sha1":          types.StringType,
	"sha2":          types.StringType,
	"metadata_timestamps": types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"properties": types.StringType,
		},
	},
}

var metadataTimestampsAttType = map[string]attr.Type{
	"properties": types.StringType,
}

func (m *FileListDataSourceModel) FromAPIModel(ctx context.Context, data FileListAPIModel) (ds diag.Diagnostics) {
	m.Uri = types.StringValue(data.Uri)
	m.Created = types.StringValue(data.Created.Format(time.RFC3339))

	var files []attr.Value
	for _, file := range data.Files {
		metadataTimestamps := types.ObjectNull(metadataTimestampsAttType)
		if file.MetadataTimestamps != nil {
			m, d := types.ObjectValue(metadataTimestampsAttType, map[string]attr.Value{
				"properties": types.StringValue(file.MetadataTimestamps.Properties),
			})
			if d != nil {
				ds.Append(d...)
			}
			metadataTimestamps = m
		}

		f := types.ObjectValueMust(
			filesAttrType,
			map[string]attr.Value{
				"uri":                 types.StringValue(file.Uri),
				"size":                types.Int64Value(file.Size),
				"last_modified":       types.StringValue(file.LastModified.Format(time.RFC3339)),
				"folder":              types.BoolValue(file.IsFolder),
				"sha1":                types.StringValue(file.SHA1),
				"sha2":                types.StringValue(file.SHA2),
				"metadata_timestamps": metadataTimestamps,
			},
		)

		files = append(files, f)
	}

	filesList, d := types.ListValue(types.ObjectType{AttrTypes: filesAttrType}, files)
	if d != nil {
		ds.Append(d...)
	}

	m.Files = filesList

	return nil
}

type FileListAPIModel struct {
	Uri     string              `json:"uri"`
	Created time.Time           `json:"created"`
	Files   []FileListAttribute `json:"files"`
}

type FileListAttribute struct {
	Uri                string                      `json:"uri"`
	Size               int64                       `json:"size"`
	LastModified       time.Time                   `json:"lastModified"`
	IsFolder           bool                        `json:"folder"`
	SHA1               string                      `json:"sha1"`
	SHA2               string                      `json:"sha2"`
	MetadataTimestamps *FileListMetadataTimestamps `json:"mdTimestamps"`
}

type FileListMetadataTimestamps struct {
	Properties string `json:"properties"`
}

func (d *FileListDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = d.TypeName
}

func (d *FileListDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"repository_key": schema.StringAttribute{
				Description: "Repository key",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"folder_path": schema.StringAttribute{
				Description: "Path of the folder",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"deep_listing": schema.BoolAttribute{
				Description: "Get deep listing",
				Optional:    true,
			},
			"depth": schema.Int64Attribute{
				Description: "Depth of the deep listing",
				Optional:    true,
			},
			"list_folders": schema.BoolAttribute{
				Description: "Include folders",
				Optional:    true,
			},
			"metadata_timestamps": schema.BoolAttribute{
				Description: "Include metadata timestamps",
				Optional:    true,
			},
			"include_root_path": schema.BoolAttribute{
				Description: "Include root path",
				Optional:    true,
			},
			"uri": schema.StringAttribute{
				Description: "URL to file/path",
				Computed:    true,
			},
			"created": schema.StringAttribute{
				Description: "Creation time",
				Computed:    true,
			},
			"files": schema.ListNestedAttribute{
				Description: "A list of files.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"uri": schema.StringAttribute{
							Description: "URL to file",
							Computed:    true,
						},
						"size": schema.Int64Attribute{
							Description: "File size in bytes",
							Computed:    true,
						},
						"last_modified": schema.StringAttribute{
							Description: "Last modified time",
							Computed:    true,
						},
						"folder": schema.BoolAttribute{
							Description: "Is this a folder",
							Computed:    true,
						},
						"sha1": schema.StringAttribute{
							Description: "SHA-1 checksum",
							Computed:    true,
						},
						"sha2": schema.StringAttribute{
							Description: "SHA-256 checksum",
							Computed:    true,
						},
						"metadata_timestamps": schema.SingleNestedAttribute{
							Description: "File metadata",
							Computed:    true,
							Attributes: map[string]schema.Attribute{
								"properties": schema.StringAttribute{
									Description: "Properties timestamp",
									Computed:    true,
								},
							},
						},
					},
				},
			},
		},
		Description: "Get a flat (the default) or deep listing of the files and folders (not included by default) within a folder. For deep listing you can specify an optional depth to limit the results. Optionally include a map of metadata timestamp values as part of the result.",
	}
}

func (d *FileListDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	d.ProviderData = req.ProviderData.(util.ProviderMetadata)
}

func bool2String(v bool) string {
	if v {
		return "1"
	}

	return "0"
}

func (d *FileListDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data FileListDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var fileList FileListAPIModel
	folderPath := ""
	if data.FolderPath.ValueString() != "/" { // only use config folder path if it isn't just "/"
		folderPath = data.FolderPath.ValueString()
	}

	response, err := d.ProviderData.Client.R().
		SetQueryParams(map[string]string{
			"list":            "",
			"deep":            bool2String(data.DeepListing.ValueBool()),
			"depth":           fmt.Sprintf("%d", data.Depth.ValueInt64()),
			"listFolders":     bool2String(data.ListFolders.ValueBool()),
			"mdTimestamps":    bool2String(data.MetadataTimestamps.ValueBool()),
			"includeRootPath": bool2String(data.IncludeRootPath.ValueBool()),
		}).
		SetResult(&fileList).
		SetPathParams(map[string]string{
			"repoKey":    data.RepositoryKey.ValueString(),
			"folderPath": folderPath,
		}).
		Get("artifactory/api/storage/{repoKey}/{folderPath}")

	// Treat HTTP 404 Not Found status as a signal to recreate resource
	// and return early
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Data Source",
			"An unexpected error occurred while fetch the data source. "+
				"Please report this issue to the provider developers.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	if response.IsError() {
		resp.Diagnostics.AddError(
			"Unable to Read Data Source",
			"An unexpected error occurred while fetch the data source. "+
				"Please report this issue to the provider developers.\n\n"+
				"Error: "+response.String(),
		)
		return
	}

	// Convert from the API data model to the Terraform data model
	// and refresh any attribute values.
	resp.Diagnostics.Append(data.FromAPIModel(ctx, fileList)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
