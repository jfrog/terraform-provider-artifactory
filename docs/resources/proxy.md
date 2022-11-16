---
subcategory: "Configuration"
---
# Artifactory Proxy Resource

Provides an Artifactory Proxy resource.

This resource configuration corresponds to 'proxies' config block in system configuration XML
(REST endpoint: [artifactory/api/system/configuration](https://www.jfrog.com/confluence/display/JFROG/Artifactory+REST+API#ArtifactoryRESTAPI-GeneralConfiguration)).

~>The `artifactory_proxy` resource utilizes endpoints which are blocked/removed in SaaS environments (i.e. in Artifactory online), rendering this resource incompatible with Artifactory SaaS environments.

## Example Usage

```hcl
resource "artifactory_proxy" "my-proxy" {
  key               = "my-proxy"
  host              = "my-proxy.mycompany.com"
  port              = 8888
  username          = "user1"
  password          = "password"
  nt_host           = "MYCOMPANY.COM"
  nt_domain         = "MYCOMPANY"
  platform_default  = false
  redirect_to_hosts = ["redirec-host.mycompany.com"]
  services          = ["jfrt", "jfxr"]
}
```

## Argument Reference

The following arguments are supported:

* `key` - (Required) The unique ID of the proxy.
* `host` - (Required) The name of the proxy host.
* `port` - (Required) The proxy port number.
* `username` - (Optional) The proxy username when authentication credentials are required.
* `password` - (Optional) The proxy password when authentication credentials are required.
* `nt_host` - (Optional) The computer name of the machine (the machine connecting to the NTLM proxy).
* `nt_domain` - (Optional) The proxy domain/realm name.
* `platform_default` - (Optional) When set, this proxy will be the default proxy for new remote repositories and for internal HTTP requests issued by Artifactory. Will also be used as proxy for all other services in the platform (for example: Xray, Distribution, etc).
* `redirect_to_hosts` - (Optional) An optional list of host names to which this proxy may redirect requests. The credentials defined for the proxy are reused by requests redirected to all of these hosts.
* `services` - (Optional) An optional list of services names to which this proxy be the default of. The options are `jfrt`, `jfmc`, `jfxr`, `jfds`.

## Import

Current Proxy can be imported using `proxy-key` from Artifactory as the `ID`, e.g.

```
$ terraform import artifactory_proxy.my-proxy proxy-key
```
