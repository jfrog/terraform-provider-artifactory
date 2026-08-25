# CI: Build Gate (JTFPR-262)

This branch (`feature/JTFPR-262`) changes how GitHub Actions runs on **customer / fork pull requests**.

Ticket: [JTFPR-262](https://jfrog-int.atlassian.net/browse/JTFPR-262)  
Pattern: [jfrog/jfrog-cli build-gate](https://github.com/jfrog/jfrog-cli/blob/master/.github/workflows/build-gate.yml)

## Problem we are fixing

This repo is public. Customer PRs come from **forks**. The old workflow used `on: pull_request` and `environment: development`.

On a public fork that combination:

1. Never injects Artifactory license / Slack secrets
2. Often sits on “Approve and run workflows” or environment rejection
3. Never reaches `make acceptance`
4. Even if tests ran, `update-changelog` could not `git push` to the customer’s branch

The workaround was to copy the customer’s commits onto a same-repo branch and open an internal PR.

## What this branch does

Same idea as jfrog-cli: **one maintainer approval**, then tests run in the **base repo** context so secrets exist.

```
Customer PR (fork)
        │
        ▼
pull_request_target  ── workflow YAML always from our `master`, never from the fork
        │
        ▼
gate job  ── waits on GitHub environment `build-gate` (required reviewers)
        │
        ▼  (only after approval)
acceptance tests  ── checkout PR SHA, secrets: inherit, persist-credentials: false
        │
        ▼
build-gate-success  ── single required status check
```

| Change | File |
|---|---|
| New orchestrator | [`build-gate.yml`](build-gate.yml) |
| Tests are reusable only (`workflow_call`) | [`acceptance-tests.yml`](acceptance-tests.yml) |
| Checkout customer SHA after the gate; no git write token on that tree | `persist-credentials: false` |
| Changelog `git push` only for same-repo PRs | `update-changelog` job |
| Opt in to fetching fork PR code under `pull_request_target` | `allow-unsafe-pr-checkout: true` |

### Why `allow-unsafe-pr-checkout: true` is set

Since July 2026, `actions/checkout` refuses by default to fetch fork pull request code in a `pull_request_target` workflow ([changelog](https://github.blog/changelog/2026-06-18-safer-pull_request_target-defaults-for-github-actions-checkout/)). Without the opt-in, the suite fails at the Checkout step with *"Refusing to check out fork pull request code"* — this is the same protection whether or not a gate exists, because the action cannot see our gate.

GitHub's guidance is to opt in only when the fork code is never executed. That does not hold here: this suite exists to `go build` and `go test` the customer's provider. The risk is instead carried by the `build-gate` environment — no fork code is fetched at all until a maintainer approves, so approval is the review step the flag assumes. Keep treating approval as a trust decision, and do not move this checkout to a job that runs before `gate`.

Do **not** add `pull_request`, ungated `pull_request_target`, or `workflow_dispatch` back onto `acceptance-tests.yml`. That either withholds secrets or runs untrusted code with secrets and no approval.

## Secrets are not given to the customer PR automatically

| Layer | What it does |
|---|---|
| `pull_request_target` | Uses the workflow on `master`. A fork cannot change the gate. |
| Environment `build-gate` | Test jobs do not start (and do not receive secrets) until a maintainer approves. |
| `secrets: inherit` | Passed only to jobs that `need: gate`. |
| `persist-credentials: false` | Customer tree does not keep a `GITHUB_TOKEN` for `git push`. |
| No changelog push to forks | We never write to the customer’s repository. |

After you approve, tests **do** run the customer SHA with Artifactory license and password secrets. That is required to test their code. Treat the approval as a trust decision: do not approve a PR that looks like it will dump env or exfiltrate credentials.

`workflow_dispatch` still requires `build-gate` approval and must be run from the default branch so the workflow YAML is the trusted copy. Write access is not a substitute for that approval. Re-run a failed suite with **Re-run failed jobs** instead of dispatching from a feature branch.

## Changelog

There are two different changelog items.

### 1. Customer-written entry (keep it)

Example: a PR that already adds a `CHANGELOG.md` bullet for a resource change. That is a normal PR file. **Merging the PR lands that bullet.** CI does not write it.

### 2. CI “Tested on Artifactory …” heading (we cannot push to a fork)

The old job rewrote the version heading and `git push`ed to the PR branch. `GITHUB_TOKEN` cannot write to `customer/terraform-provider-artifactory`.

| PR type | After tests pass |
|---|---|
| Same-repo (`jfrog/terraform-provider-artifactory`) | CI still updates `CHANGELOG.md` and pushes, as before |
| Fork (customer) | Push is skipped. CI comments on the PR. Slack still asks for review. |

**While merging a customer PR:** resolve `CHANGELOG.md` if it conflicts (move their bullet under the current version), and add the “Tested on Artifactory …” suffix yourself if you want it. You can do that in the merge conflict UI, as a maintainer edit on their branch (if they allow edits), or as a small commit on `master` after merge.

You do not need to copy the PR onto a new branch just to get CI or changelog.

This matches jfrog-cli: they never push changelog onto a PR. Notes are handled at merge/release time.

## Maintainer flow for a customer PR

1. Review the diff (including their `CHANGELOG.md` bullet).
2. In Actions, approve the **`build-gate`** environment deployment **once**.
3. Terraform and OpenTofu acceptance tests run (local Artifactory container).
4. When `Build Gate / build-gate-success` is green, merge.
5. Update the changelog heading / conflict while merging if needed.

To recover a failed test suite: **Re-run failed jobs**. That does not re-ask for `build-gate` approval and does not need a new commit.

## One-time GitHub settings (not in git)

These must be done on `jfrog/terraform-provider-artifactory` after this lands on `master`:

1. **Settings → Environments → New environment: `build-gate`**
   - Add **Required reviewers** (the maintainers who today copy customer branches).
   - Limit **Deployment branches** to the default branch so a modified workflow on another ref cannot use this environment.
2. Leave test secrets/vars on **`development`** (or repo/org secrets).
3. **Do not** add required reviewers on `development`, or you will approve twice.
4. **Branch protection:** require `Build Gate / build-gate-success`. Stop requiring the old matrix job names (`terraform` / `tofu`).

Until `build-gate.yml` is on `master`, `pull_request_target` still uses the old workflow from the default branch.

## Files

| Workflow | Trigger | Role |
|---|---|---|
| `build-gate.yml` | `pull_request_target` on `master` (Go / these workflow files), `workflow_dispatch` from the default branch | Approval + call tests + aggregator |
| `acceptance-tests.yml` | `workflow_call` only | Artifactory container + Terraform/OpenTofu tests + changelog |
| `cla.yml` | `pull_request_target` | CLA (unchanged) |
| `slack-notify-pr.yml` | `pull_request_target` | Slack notify only (unchanged) |
| `slack-notify-issues.yml` | issues | Slack notify (unchanged) |
| `release.yml` | tag `v*` | Release (unchanged) |
| `changelog.yml` | pull request | Changelog reminder (unchanged) |
