# Xray Policy Resource

Provides an Xray policy resource. This can be used to create and manage Xray v1 policies.

## Example Usage

```hcl
# Create a new Xray license policy
resource "xray_policy" "example" {
  name  = "policy-name"
  description = "license policy description"
  type = "license"

  rules {
    name = "license rule"
    priority = 1
    criteria {
      allowed_licenses = ["0BSD", "AAL"]
    }
  }
}

# Create a new Xray watch for all repositories and assign the policy
resource "xray_watch" "example" {
  name  = "watch-name"
  description = "watching all repositories"
  resources {
    type = "all-repos"
    name = "All Repositories"
  }
  assigned_policies {
    name = xray_policy.example.name
    type = "license"
  }
}
```

## Attribute Reference

The following arguments are supported:

* `name` - (Required) Name of the policy (must be unique)
* `type` - (Required) Type of the policy
* `description` - (Optional) More verbose description of the policy
* `author` - (Optional) Name of the policy author
* `rules` - (Required) Nested block describing the policy rules. Described below.

### Rules

The top-level `rules` block is a list of one or more rules that each supports the following:

* `name` - (Required) Name of the rule
* `priority` - (Required) Integer describing the rule priority
* `criteria` - (Required) Nested block describing the criteria for the policy. Described below.
* `actions` - (Required) Nested block describing the actions to be applied by the policy. Described below.

#### criteria

~> **NOTE:** Only one of either security criteria (`min_severity` and `cvss_range`) or license criteria (`allow_unknown`,
`banned_licenses`, and `allowed_licenses`) may be specified. While all attributes are marked as optional, at least one
attribute from only one of these groups must be defined.

The nested `criteria` block is a list of one item, supporting the following:

##### Security criteria

* `min_severity` - (Optional) The minimum security vulnerability severity that will be impacted by the policy.
* `cvss_range` - (Optional) Nested block describing a CVS score range to be impacted. Defined below.

###### cvss_range

The nested `cvss_range` block is a list of one object that contains the following attributes:

* `to` - (Required) The beginning of the range of CVS scores (from 1-10) to flag.
* `from` - (Required) The end of the range of CVS scores (from 1-10) to flag.

##### License criteria

* `allow_unknown` - (Optional) Whether or not to allow components whose license cannot be determined (`true` or `false`).
* `banned_licenses` - (Optional) A list of OSS license names that may not be attached to a component.
* `allowed_licenses` - (Optional) A list of OSS license names that may be attached to a component.

#### actions

~> **NOTE:** While all of the actions attributes are marked as optional, at least one action must be specified.

The nested `actions` block is a list of exactly one object with the following attributes:

* `mails` - (Optional) A list of email addressed that will get emailed when a violation is triggered.
* `fail_build` - (Optional) Whether or not the related CI build should be marked as failed if a violation is triggered. This option is only available when the policy is applied to an `xray_watch` resource with a `type` of `builds`.
* `block_download` - (Optional) Nested block describing artifacts that should be blocked for download if a violation is triggered. Described below.
* `webhooks` - (Optional) A list of Xray-configured webhook URLs to be invoked if a violation is triggered.
* `custom_severity` - (Optional) The severity of violation to be triggered if the `criteria` are met.

###### block_download

~> **NOTE:** Only one of `unscanned` or `active` may be set to `true`.

The nested `block_download` block is a list of exactly one object with the following attributes:

* `unscanned` - Whether or not to block download of artifacts that meet the artifact `filters` for the associated `xray_watch` resource but have not been scanned yet.
* `active` - Whether or not to block download of artifacts that meet the artifact and severity `filters` for the associated `xray_watch` resource.


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `created` - Timestamp of when the policy was first created
* `modified` - Timestamp of when the policy was last modified

## Import

A policy can be imported by using the name, e.g.

```
$ terraform import xray_policy.example policy-name
```
