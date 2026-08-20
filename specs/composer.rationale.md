# Composer — spec rationale

Extracted from the shipped implementation as of the `update-repository` reconciliation
run on branch `JTFPR-196` (2026-08-21). One spec per class, per the extract-spec
granularity rule: [`composer-local.repository.yaml`](composer-local.repository.yaml),
[`composer-remote.repository.yaml`](composer-remote.repository.yaml),
[`composer-virtual.repository.yaml`](composer-virtual.repository.yaml),
[`composer-federated.repository.yaml`](composer-federated.repository.yaml).

## Why local + federated are now `dedicated` instead of generic-slice

Live REST verification (create + `GET /artifactory/api/repositories/{key}`) found
`enableComposerV1Indexing` present on both local and federated composer repos —
confirmed against `LocalRepositoryConfigurationImpl`'s
`@IncludeTypeSpecific(packageType = Composer, repoType = {LOCAL, FEDERATED})`
annotation — but the provider never exposed it: composer was carried entirely
through the fully-generic slice for both classes. No existing generic-slice
variant adds a single extra local/federated boolean, so both were promoted to
dedicated resources mirroring the `debian` pattern (Plugin Framework local
resource + legacy SDKv2 schema reused by the SDKv2 federated resource). This is
purely a Go-implementation change; the Terraform resource type names
(`artifactory_local_composer_repository`, `artifactory_federated_composer_repository`)
are unchanged, and a provider-upgrade test (apply on the last published version
v12.11.11, then switch to the local build and assert an empty plan) passed for
both classes.

## Why virtual gained `retrieval_cache_period_seconds`

Same story: the field is real and independently settable per a live REST
round-trip, even though the backend's `@IncludeTypeSpecific` annotation for this
exact field on `VirtualRepositoryConfigurationImpl` does not list Composer (only
Chef/Conda/CRAN/Conan/Debian/Helm/Npm/YUM are listed there). REST is the
authoritative source per the source-of-truth matrix, so it overrides the static
annotation read. Composer moved from `virtual.PackageTypesLikeGeneric` to
`virtual.PackageTypesLikeGenericWithRetrievalCachePeriodSecs` — no new Go file
needed, since this generic-slice variant already existed.

**Known, accepted tradeoff:** the shared schema for this slice uses a single
literal default (`7200`) across every package type on the list. This instance's
own live default before Terraform manages the field was observed as `600`
(likely an instance-level system default, not a per-type constant — an NPM
virtual repo probed the same way also returned `600`). Because the field was
never previously managed by Terraform for composer, existing customers will see
a one-time `update` plan on their first apply after upgrading, settling the
value to `7200`. This was confirmed with an upgrade test
(`TestAccVirtualComposerRepository_UpgradeFromSDKv2`, asserting the post-upgrade
value rather than an empty plan) and explicitly checkpointed with — and accepted
by — the contributor, since it is identical in nature to what every other type
on this shared list would have experienced when first added, and there is no
way to give composer a literal default that matches every real instance's
current live value.

## Why `composer_registry_url` (remote) needed no change

`artifactory-service` declares two different default constants depending on
code path — `ComposerConfiguration.java`'s `@XmlElement` default is
`"https://packagist.org"`, while `RepoConfigDefaultValues.DEFAULT_COMPOSER_REGISTRY`
is `"https://repo.packagist.org"`. Live REST create+GET confirmed the actual
server default is `"https://packagist.org"`, which already matched the shipped
provider default — no divergence, no code change.

## Why `vcs_type` is not exposed anywhere

Checked for remote (the only class where the VCS-group fields apply). The
backend's `VcsType` enum has exactly one value (`GIT`), so the field carries no
configurable information. This is consistent with how the provider already
treats every other VCS-group type (bower, cocoapods, go, vcs) — not a
composer-specific gap.

## Curation

Curation support (`curated` / `pass_through`) is resolved dynamically at
runtime from Xray's `getCuratedPackageTypes()` API response rather than a
static list in `artifactory-service` source, so it can't be captured as a fixed
spec fact the way REST-verified defaults can. Marked `unverified` in all four
specs — out of scope for this reconciliation run.
