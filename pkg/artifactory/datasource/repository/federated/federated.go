package federated

import "github.com/jfrog/terraform-provider-artifactory/v8/pkg/artifactory/resource/repository/federated"

const rclass = "federated"

var memberSchema = federated.MemberSchemaGenerator(false)
