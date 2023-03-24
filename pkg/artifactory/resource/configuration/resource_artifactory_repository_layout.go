package configuration

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/util"
	"gopkg.in/yaml.v3"
)

type Layout struct {
	Name                             string `hcl:"name" xml:"name" yaml:"name"`
	ArtifactPathPattern              string `hcl:"artifact_path_pattern" xml:"artifactPathPattern" yaml:"artifactPathPattern"`
	DistinctiveDescriptorPathPattern bool   `hcl:"distinctive_descriptor_path_pattern" xml:"distinctiveDescriptorPathPattern" yaml:"distinctiveDescriptorPathPattern"`
	DescriptorPathPattern            string `hcl:"descriptor_path_pattern" xml:"descriptorPathPattern" yaml:"descriptorPathPattern"`
	FolderIntegrationRevisionRegExp  string `hcl:"folder_integration_revision_regexp" xml:"folderIntegrationRevisionRegExp" yaml:"folderIntegrationRevisionRegExp"`
	FileIntegrationRevisionRegExp    string `hcl:"file_integration_revision_regexp" xml:"fileIntegrationRevisionRegExp" yaml:"fileIntegrationRevisionRegExp"`
}

func (l Layout) Id() string {
	return l.Name
}

type Layouts struct {
	Layouts []Layout `xml:"repoLayouts>repoLayout" yaml:"repoLayout"`
}

func ResourceArtifactoryRepositoryLayout() *schema.Resource {
	var layoutSchema = map[string]*schema.Schema{
		"name": {
			Type:             schema.TypeString,
			Required:         true,
			ForceNew:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
			Description:      "Layout name",
		},
		"artifact_path_pattern": {
			Type:             schema.TypeString,
			Required:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
			Description:      "Please refer to: [Path Patterns](https://www.jfrog.com/confluence/display/JFROG/Repository+Layouts#RepositoryLayouts-ModulesandPathPatternsusedbyRepositoryLayouts) in the Artifactory Wiki documentation.",
		},
		"distinctive_descriptor_path_pattern": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "When set, 'descriptor_path_pattern' will be used. Default to 'false'.",
		},
		"descriptor_path_pattern": {
			Type:             schema.TypeString,
			Optional:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
			Description:      "Please refer to: [Descriptor Path Patterns](https://www.jfrog.com/confluence/display/JFROG/Repository+Layouts#RepositoryLayouts-DescriptorPathPatterns) in the Artifactory Wiki documentation.",
		},
		"folder_integration_revision_regexp": {
			Type:             schema.TypeString,
			Required:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
			Description:      "A regular expression matching the integration revision string appearing in a folder name as part of the artifact's path. For example, 'SNAPSHOT', in Maven. Note! Take care not to introduce any regexp capturing groups within this expression. If not applicable use '.*'",
		},
		"file_integration_revision_regexp": {
			Type:             schema.TypeString,
			Required:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
			Description:      "A regular expression matching the integration revision string appearing in a file name as part of the artifact's path. For example, 'SNAPSHOT|(?:(?:[0-9]{8}.[0-9]{6})-(?:[0-9]+))', in Maven. Note! Take care not to introduce any regexp capturing groups within this expression. If not applicable use '.*'",
		},
	}

	var unpackLayout = func(s *schema.ResourceData) Layout {
		d := &util.ResourceData{ResourceData: s}
		return Layout{
			Name:                             d.GetString("name", false),
			ArtifactPathPattern:              d.GetString("artifact_path_pattern", false),
			DistinctiveDescriptorPathPattern: d.GetBool("distinctive_descriptor_path_pattern", false),
			DescriptorPathPattern:            d.GetString("descriptor_path_pattern", false),
			FolderIntegrationRevisionRegExp:  d.GetString("folder_integration_revision_regexp", false),
			FileIntegrationRevisionRegExp:    d.GetString("file_integration_revision_regexp", false),
		}
	}

	var resourceLayoutRead = func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		data := &util.ResourceData{ResourceData: d}
		name := data.GetString("name", false)

		layouts := Layouts{}
		_, err := m.(util.ProvderMetadata).Client.R().SetResult(&layouts).Get("artifactory/api/system/configuration")
		if err != nil {
			return diag.Errorf("failed to retrieve data from API: /artifactory/api/system/configuration during Read")
		}

		matchedLayout := FindConfigurationById[Layout](layouts.Layouts, name)
		if matchedLayout == nil {
			d.SetId("")
			return nil
		}

		pkr := packer.Default(layoutSchema)

		return diag.FromErr(pkr(matchedLayout, d))
	}

	var resourceLayoutUpdate = func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		unpackedLayout := unpackLayout(d)

		/* EXPLANATION FOR BELOW CONSTRUCTION USAGE.
		There is a difference in xml structure usage between GET and PATCH calls of API: /artifactory/api/system/configuration.
		GET call structure has "backups -> backup -> Array of backup config blocks".
		PATCH call structure has "backups -> Name/Key of backup that is being patched -> config block of the backup being patched".
		Since the Name/Key is dynamic string, following nested map of string structs are constructed to match the usage of PATCH call.
		*/
		constructBody := map[string]map[string]Layout{
			"repoLayouts": map[string]Layout{
				unpackedLayout.Name: unpackedLayout,
			},
		}
		content, err := yaml.Marshal(&constructBody)

		if err != nil {
			return diag.FromErr(err)
		}

		err = SendConfigurationPatch(content, m)
		if err != nil {
			return diag.FromErr(err)
		}

		d.SetId(unpackedLayout.Name)

		return resourceLayoutRead(ctx, d, m)
	}

	var resourceLayoutDelete = func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		unpackedLayout := unpackLayout(d)

		deleteLayoutConfig := fmt.Sprintf(`
repoLayouts:
  %s: ~
`, unpackedLayout.Name)

		err := SendConfigurationPatch([]byte(deleteLayoutConfig), m)
		if err != nil {
			return diag.FromErr(err)
		}

		d.SetId("")
		return nil
	}

	var distinctiveDescriptorPathPatternDiff = func(ctx context.Context, diff *schema.ResourceDiff, v interface{}) error {
		distinctiveDescriptorPathPattern := diff.Get("distinctive_descriptor_path_pattern").(bool)
		descriptorPathPattern := diff.Get("descriptor_path_pattern").(string)

		if distinctiveDescriptorPathPattern && len(descriptorPathPattern) == 0 {
			return fmt.Errorf("descriptor_path_pattern must be set when distinctive_descriptor_path_pattern is true")
		}

		return nil
	}

	return &schema.Resource{
		UpdateContext: resourceLayoutUpdate,
		CreateContext: resourceLayoutUpdate,
		DeleteContext: resourceLayoutDelete,
		ReadContext:   resourceLayoutRead,

		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
				d.Set("name", d.Id())
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema:        layoutSchema,
		CustomizeDiff: distinctiveDescriptorPathPatternDiff,
		Description:   "Provides an Artifactory repository layout resource. See [Repository Layout documentation](https://www.jfrog.com/confluence/display/JFROG/Repository+Layouts) for more details.",
	}
}
