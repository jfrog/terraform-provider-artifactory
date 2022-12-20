---
page_title: "Migrating from access token to scoped token"
---

Artifactory version 7.21.1 introduced scoped token which replaces the access token. In this provider, we have `artifactory_scoped_token` and `artifactory_access_token` resources supporting both accordingly.

While access token continues to be supported and can still be used to generate a token, we recommend only using scoped token for any new token resources.

Since Artifactory API does not allow retrieval of an existing token (for good security reason), importing an existing token (access or scoped) into Terraform state is not supported. Therefore, we recommend that when the time comes to rotate/replace an existing **access** token, you replace it with a new **scoped** token. This consists of creating a new scoped token resource and update any references to the existing access token with the new scoped token.

Once you've verified the new scoped token is working, you can safely remove the old access token resource from your Terraform configuration.
