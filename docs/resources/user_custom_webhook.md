---
subcategory: "Webhook"
---
# Artifactory "User" Custom Webhook Resource

Provides an Artifactory custom webhook resource. This can be used to register and manage Artifactory webhook subscription which enables you to be notified or notify other users when such events take place in Artifactory.

## Example Usage
.
```hcl
resource "artifactory_user_custom_webhook" "user-custom-webhook" {
  key         = "user-custom-webhook"
  event_types = ["locked"]

  handler {
    url       = "https://tempurl.org"
    secrets   = {
      secretName1 = "value1"
      secretName2 = "value2"
    }
    http_headers  = {
      headerName1    = "value1"
      headerName2    = "value2"
    }
    payload = "{ \"ref\": \"main\" , \"inputs\": { \"artifact_path\": \"test-repo/repo-path\" } }"
  }
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog Webhook API](https://www.jfrog.com/confluence/display/JFROG/Artifactory+REST+API). The following arguments are supported:

The following arguments are supported:

* `key` - (Required) The identity key of the webhook. Must be between 2 and 200 characters. Cannot contain spaces.
* `description` - (Optional) Webhook description. Max length 1000 characters.
* `enabled` - (Optional) Status of webhook. Default to `true`
* `event_types` - (Required) List of event triggers for the Webhook. Allow values: `locked`
* `handler` - (Required) At least one is required.
  * `url` - (Required) Specifies the URL that the Webhook invokes. This will be the URL that Artifactory will send an HTTP POST request to.
  * `secrets` - (Optional) Defines a set of sensitive values (such as, tokens and passwords) that can be injected in the headers and/or payload.Secrets’ values are encrypted. In the header/payload, the value can be invoked using the `{{.secrets.token}}` format, where token is the name provided for the secret value. Comprise key/value pair. **Note:** if multiple handlers are used, same secret name and different secret value for the same url won't work. Example:

```hcl
handler {
  url       = "https://tempurl.org" # same url in both handlers
  secrets   = {
    secretName1 = "value1"
    secretName2 = "value2"
  }
  http_headers  = {
    headerName1    = "value1"
    headerName2    = "value2"
  }
  payload = "{ \"ref\": \"main\" , \"inputs\": { \"artifact_path\": \"test-repo/repo-path\" } }"
}
handler {
  url       = "https://tempurl.org" # same url in both handlers
  secrets   = {
    secretName1 = "newValue1" # same secret name, but different value
    secretName2 = "newValue2" # same secret name, but different value
  }
  http_headers  = {
    headerName1    = "value1"
    headerName2    = "value2"
  }
  payload = "{ \"ref\": \"main\" , \"inputs\": { \"artifact_path\": \"test-repo/repo-path\" } }"
}
```

* `proxy` - (Optional) Proxy key from Artifactory UI (Administration -> Proxies -> Configuration).
* `http_headers` - (Optional) HTTP headers you wish to use to invoke the Webhook, comprise key/value pair.
