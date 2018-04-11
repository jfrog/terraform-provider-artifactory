# Contributing to go-artifactory

## Reporting Issues

This section guides you through submitting a bug report for go-artifactory. Following these guidelines helps us and the community understand your issue, reproduce the behavior, and find related issues.

When you are creating an issue, please include as many details as possible.

### Before submitting an issue

* **Perform a [cursory search][IssueTracker]** to see if the problem has already been reported. If it has, add a comment to the existing issue instead of opening a new one.

### How do I submit a (good) issue?

* **Use a clear and descriptive title** for the issue to identify the problem.
* **Describe the exact steps which reproduce the problem** in as many details as possible. When listing steps, **don't just say what you did, but explain how you did it**. For example, if you opened a inline dialog, explain if you used the mouse, or a keyboard shortcut.
* **If the problem wasn't triggered by a specific action**, describe what you were doing before the problem happened and share more information using the guidelines below.

Include details about your configuration and environment:

* **Which OS are you running on?**
* **What version of golang are you using**?

### Code Contributions

#### Why should I contribute?

1. While we strive to look at new issues as soon as we can, because of the many priorities we juggle and limited resources, issues raised often don't get looked into soon enough.
2. We want your contributions. We are always trying to improve our docs, processes and tools to make it easier to submit your own changes.
3. At Atlassian, "Play, As A Team" is one of our values. We encourage cross team contributions and collaborations.

Please raise a new issue [here][IssueTracker].

### Follow code style guidelines

It is recommended you use the git hooks found in the misc directory, this will include go-fmt

## Merge into master
All new feature code must be completed in a feature branch and have a corresponding Feature or Bug issue in the go-artifactory project.

Once you are happy with your changes, you must push your branch to Bitbucket and create a pull request. All pull requests must have at least 2 reviewers from the go-artifactory team. Once the pull request has been approved it may be merged into develop.

A separate pull request can be made to create a release and merge develop into master.

Each PR should consist of exactly one commit, use git rebase and squash, and should be as small as possible. If you feel multiple commits are warrented you should probably be filing them as multiple PRs.

**Attention!**: *Merging into master will automatically release a component. See below for more details*

## Release a component
Releasing components is completely automated. The process of releasing will begin when changes are made to the `master` branch:

* Pipelines will move the go branch forward after successful build on master. This will change the version acquired by go-get

## Root dependencies

go-artifactory endeavours to avoid external dependencies and be lightweight.

[IssueTracker]: https://github.com/atlassian/go-artifactory/issues