package federated

import "github.com/jfrog/terraform-provider-artifactory/v10/pkg/artifactory/resource/repository/federated"

const rclass = "federated"

var federatedSchema = federated.SchemaGenerator(false)
