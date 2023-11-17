package federated

import "github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository/federated"

const rclass = "federated"

var federatedSchema = federated.SchemaGenerator(false)
