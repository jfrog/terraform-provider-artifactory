package security

import (
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func unableToCreateResourceError(resp *resource.CreateResponse, err error) {
	resp.Diagnostics.AddError(
		"Unable to Create Resource",
		"An unexpected error occurred while creating the resource update request. "+
			"Please report this issue to the provider developers.\n\n"+
			"JSON Error: "+err.Error(),
	)
}

func unableToUpdateResourceError(resp *resource.UpdateResponse, err error) {
	resp.Diagnostics.AddError(
		"Unable to Update Resource",
		"An unexpected error occurred while updating the resource update request. "+
			"Please report this issue to the provider developers.\n\n"+
			"JSON Error: "+err.Error(),
	)
}

func unableToRefreshResourceError(resp *resource.ReadResponse, err error) {
	resp.Diagnostics.AddError(
		"Unable to Refresh Resource",
		"An unexpected error occurred while attempting to refresh resource state. "+
			"Please retry the operation or report this issue to the provider developers.\n\n"+
			"HTTP Error: "+err.Error(),
	)
}

func unableToDeleteResourceError(resp *resource.DeleteResponse, err error) {
	resp.Diagnostics.AddError(
		"Unable to Delete Resource",
		"An unexpected error occurred while attempting to delete the resource. "+
			"Please retry the operation or report this issue to the provider developers.\n\n"+
			"HTTP Error: "+err.Error(),
	)
}
