---
subcategory: "Configuration"
---
# Artifactory Property Set Resource

Provides an Artifactory Property Set resource.
This resource configuration corresponds to 'propertySets' config block in system configuration XML
(REST endpoint: artifactory/api/system/configuration).

~>The `artifactory_property_set` resource utilizes endpoints which are blocked/removed in SaaS environments (i.e. in Artifactory online), rendering this resource incompatible with Artifactory SaaS environments.

## Example Usage

```hcl
resource "artifactory_property_set" "foo" {
  name 		= "property-set1"
  visible 	= true

  property {
    name = "set1property1"

    predefined_value {
      name 			    = "passed-QA"
      default_value 	= true
    }

    predefined_value {
      name 			    = "failed-QA"
      default_value 	= false
    }

    closed_predefined_values 	= true
    multiple_choice 			= true
  }

  property {
    name = "set1property2"

    predefined_value {
      name 			    = "passed-QA"
      default_value 	= true
    }

    predefined_value {
      name 			    = "failed-QA"
      default_value 	= false
    }

    closed_predefined_values 	= false
    multiple_choice 			= false
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Property set name.
* `visible` - (Optional) Defines if the list visible and assignable to the repository or artifact. Default value is `true`.
* `property` - (Required) A list of properties that will be part of the property set.
  * `name` - (Required) The name pf the property.
  * `closed_predefined_values` - (Required) Disables `multiple_choice` if set to `false` at the same time with multiple_choice set to `true`. Default value is `false`
  * `multiple_choice` - (Optional) Defines if user can select multiple values. `closed_predefined_values` should be set to `true`. Default value is `false`.
    * `predefined_value` - (Required) Properties in the property set.  
      * `name` - (Required) Predefined property name.
      * `default_value` - (Required) Whether the value is selected by default in the UI.


## Import

Current Property Set can be imported using `property-set1` as the `ID`, e.g.

```
$ terraform import artifactory_property_set.foo property-set1
```
