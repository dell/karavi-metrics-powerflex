<!--
Copyright (c) 2020 Dell Inc., or its subsidiaries. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0
-->
# How to Contribute

Become one of the contributors to this project! We thrive to build a welcoming and open community for anyone who wants to use the project or contribute to it. There are just a few small guidelines you need to follow. To help us create a safe and positive community experience for all, we require all participants to adhere to the [Code of Conduct](CODE_OF_CONDUCT.md).

# Table of Content:
* [Become a contributor](#Become-a-contributor)
* [Report bugs](#Report-bugs)
* [Feature request](#Feature-request)
* [Answering questions](#Answering-questions)
* [Triage issues](#Triage-issues)
* [Your first contribution](#Your-first-contribution)
* [Branching Strategy](#Branching-strategy)
* [Signing your commits](#Signing-your-commits)
* [Pull requests](#Pull-requests)
* [Code reviews](#Code-reviews)
* [Code Style](#Code-Style)
* [TODOs in the code](#TODOs-in-the-code)
* [Building, Deploying, and Testing](#Building-Deploying-and-Testing)

# Become a contributor

You can contribute to Karavi PowerFlex Metrics in several ways. Here are some examples:

- Contribute to the Karavi PowerFlex Metrics codebase.
- Report and triage bugs.
- Feature requests
- Write technical documentation and blog posts, for users and contributors.
- Help others by answering questions about Karavi PowerFlex Metrics.

# Report bugs

We aim to track and document everything related to this repo via the Issues page. The code and documentation are released with no warranties or SLAs and are intended to be supported through a community driven process.

Before submitting a new issue, try to make sure someone hasn't already reported the problem. Look through the [existing issues](https://github.com/dell/karavi-powerflex-metrics/issues) for similar issues.

Report a bug by submitting a [bug report](https://github.com/dell/karavi-powerflex-metrics/issues/new?template=bug_report.md). Make sure that you provide as much information as possible on how to reproduce the bug.

When opening a Bug please include the following information to help with debugging:

1. Version of relevant software: this software, Kubernetes, Dell Storage Platform, Helm, etc.
2. Details of the issue explaining the problem: what, when, where
3. The expected outcome that was not met (if any)
4. Supporting troubleshooting information. __Note: Do not provide private company information that could compromise your company's security.__

An Issue __must__ be created before submitting any pull request. Any pull request that is created should be linked to an Issue.

# Feature request

If you have an idea of how to improve Karavi PowerFlex Metrics, submit a [feature request](https://github.com/dell/karavi-powerflex-metrics/issues/new?template=feature_request.md).

# Answering questions

If you have a question and you can't find the answer in the documentation or issues, the next step is to submit a [question](https://github.com/dell/karavi-powerflex-metrics/issues/new?template=ask-a-question.md)

We'd love your help answering questions being asked by other Karavi users.

# Triage issues

Triage helps ensure that issues resolve quickly by:

- Ensuring the issue's intent and purpose is conveyed precisely. This is necessary because it can be difficult for an issue to explain how an end user experiences a problem and what actions they took.
- Giving a contributor the information they need before they commit to resolving an issue.
- Lowering the issue count by preventing duplicate issues.
- Streamlining the development process by preventing duplicate discussions.

If you don't have the knowledge or time to code, consider helping with _issue triage_. The Karavi community will thank you for saving them time by spending some of yours.

Read more about the ways you can [Triage issues](ISSUE_TRIAGE.md).

# Your first contribution

Unsure where to begin contributing to Karavi PowerFlex Metrics? Start by browsing issues labeled `beginner friendly` or `help wanted`.

- [Beginner-friendly](https://github.com/dell/karavi-powerflex-metrics/issues?q=is%3Aopen+is%3Aissue+label%3A%22beginner+friendly%22) issues are generally straightforward to complete.
- [Help wanted](https://github.com/dell/karavi-powerflex-metrics/issues?q=is%3Aopen+is%3Aissue+label%3A%22help+wanted%22) issues are problems we would like the community to help us with regardless of complexity.

When you're ready to contribute, it's time to create a pull request.

# Branching Strategy
We are following a scaled trunk branching strategy where short-lived branches are created off of the main branch. When coding is complete, the branch is merged back into main after being approved in a pull request code review.

## Branch Naming Convention

|  Branch Type |  Example                          |  Comment                                  |
|--------------|-----------------------------------|-------------------------------------------|
|  main        |  main                             |                                           |
|  Release     |  release-1.0                      |  hotfix: release-1.1 patch: release-1.0.1 |
|  Feature     |  feature-9-olp-support            |  "9" referring to GitHub issue ID         |
|  Bug Fix     |  bugfix-110-remove-docker-compose |  "110" referring to GitHub issue ID       |

#### Branch Types
- A Release branch is a branch created from main that will be solely used for release a Karavi version. Only critical bug fixes will be merged into this branch.
- Bug Fix branch is a branch which is created for the purpose of fixing the given defect/issue.
- Feature branch is created for a feature development purpose.

## Steps to create a branch for a bug fix or feature:
1. Fork the repository.
2. Create a branch off of the main branch. The branch name should follow [branch naming convention](#branch-naming-convention).
3. Write code, add tests, and commit to your branch. Optionally, add feature flags to disable any new features that are not yet ready for the release.
4. If other code changes have merged into the upstream main branch, perform a rebase of those changes into your branch.
5. Open a [pull request](#pull-requests) between your branch and the upstream main branch.
6. Once your pull request has merged, your branch can be deleted.

Release branches will be created from the main branch near the time of a planned release when all features are completed. Only critical bug fixes will be merged into the release branch at this time. All other bug fixes and features can continue to be merged into the main branch. When a release branch is stable, the branch will be tagged for release.

# Signing your commits

We require that developers sign off their commits to certify that they have permission to contribute the code in a pull request. This way of certifying is commonly known as the [Developer Certificate of Origin (DCO)](https://developercertificate.org/). We encourage all contributors to read the DCO text before signing a commit and making contributions.

GitHub will prevent a pull request from being merged if there are any unsigned commits.

## Signing a commit

GPG (GNU Privacy Guard) will be used to sign commits.  Follow the instructions [here](https://docs.github.com/en/free-pro-team@latest/github/authenticating-to-github/signing-commits) to create a GPG key and configure your GitHub account to use that key.

Make sure you have your user name and e-mail set.  This will be required for your signed commit to be properly verified.  Check the following references:

* Setting up your github user name [reference](https://help.github.com/articles/setting-your-username-in-git/)
* Setting up your e-mail address [reference](https://help.github.com/articles/setting-your-commit-email-address-in-git/)

Once Git and your GitHub account have been properly configured, you can add the -S flag to the git commits:
```console
$ git commit -S -m your commit message
# Creates a signed commit
```

## Commit message format

Karavi uses the guidelines for commit messages outlined in [How to Write a Git Commit Message](https://chris.beams.io/posts/git-commit/)

# Pull Requests

If this is your first time contributing to an open-source project on GitHub, make sure you read about [Creating a pull request](https://help.github.com/en/articles/creating-a-pull-request).

A pull request must always link to at least one GitHub issue. If that is not the case, create a GitHub issue and link it.

To increase the chance of having your pull request accepted, make sure your pull request follows these guidelines:

- Title and description matches the implementation.
- Commits within the pull request follow the formatting guidelines.
- The pull request closes one related issue.
- The pull request contains necessary tests that verify the intended behavior.
- If your pull request has conflicts, rebase your branch onto the master branch.

If the pull request fixes a bug:

- The pull request description must include `Fixes #<issue number>`.
- To avoid regressions, the pull request should include tests that replicate the fixed bug.

The Karavi team _squashes_ all commits into one when we accept a pull request. The title of the pull request becomes the subject line of the squashed commit message. We still encourage contributors to write informative commit messages, as they becomes a part of the Git commit body.

We use the pull request title when we generate change logs for releases. As such, we strive to make the title as informative as possible.

Make sure that the title for your pull request uses the same format as the subject line in the commit message.

## Quality Gates for pull request
Following GitHub Actions are used to enforce quality gates when a pull request is created or when any commit is made to the pull request. These GitHub Actions enforced our minimum code quality requirement for any code that get check into Karavi. If any of the gate fails, it is expected that contributor will look into the check log, understand the problem and resolve the issue.

#### Security scans
To check the security vulnerability, Karavi has following GitHub Actions that will be run when a pull request is open.
- [Golang Security Checker](https://github.com/securego/gosec) inspects source code for security vulnerability by scanning the Go AST.
- [Malwares](https://github.com/dell/common-github-actions/tree/main/malware-scanner) inspects source code for malwares.
- Container images are [scans](https://github.com/Azure/container-scan) for security vulnerability.

#### Code vet
This GitHub [action](https://github.com/dell/common-github-actions/tree/main/go-code-formatter-linter-vetter) analyzes source code to report suspicious constructs such as Printf calls whose arguments do not aligh with format string, abnormal or not used code in pull request. Refer Go [vet](https://golang.org/cmd/vet/)

#### Code lint
This Github [action](https://github.com/dell/common-github-actions/tree/main/go-code-formatter-linter-vetter) analyzes source code to flag programming errors, stylistics errors, and suspicious contructs. Refer Go [lint](https://github.com/golang/lint)

#### Code format
This GitHub [action](https://github.com/dell/common-github-actions/tree/main/go-code-formatter-linter-vetter) analyzes source code to flag formatting errors. Refer Go [format](https://golang.org/cmd/gofmt/)

#### Code sanitization
This GitHub [action](https://github.com/dell/common-github-actions/tree/main/code-sanitizer) analyzes source code for forbidden words and text. Forbidden words include such as non-inclusive language.

#### Code build/test/coverage
This GitHub [action](https://github.com/dell/common-github-actions/tree/main/go-code-tester) runs Go unit tests and check that code coverage per package meets a configured threshold. Flags error if given pull request do not meets the test coverage threshold and blocks the merge of pull request.

# Code Reviews

All submissions, including submissions by project members, require review. We use GitHub pull requests for this purpose. Consult [GitHub Help](https://help.github.com/articles/about-pull-requests/) for more information on using pull requests.

A pull request must satisfy following for it to be merged:
- All [quality check gates](#Quality-gates-for-pull-request) are passed.
- A pull request will require at least 2 maintainer approvals.
- Maintainer must perform a code review and ensure there is no malicious code.
- Maintainer must run a suite of tests that verify the quality of the code being submitted, and update the contributor if there are any failures.
- If any commits are made after the PR has been approved, the PR approval will automatically be removed and the above process must happen again

# Code Style

For the Go code in this repo, we expect the code styling outlined in [Effective Go](https://golang.org/doc/effective_go.html). In addition to this, we have the following supplements:

#### Handle Errors
See https://golang.org/doc/effective_go.html#errors.

Do not discard errors using _ variables. If a function returns an error, check it to make sure the function succeeded.  Handle the error, return it, or, in truly exceptional situations, panic.  This can be checked using the errcheck tool if you have it installed locally.

Do not log the error if it will also be logged by a caller higher up the call stack;  doing so causes the logs to become repetitive.  Instead, consider wrapping the error in order to provide more detail.  To see practical examples of this, see this bad practice and this preferred practice:

##### Bad

```
package main
 
 import (
 	"errors"
 	"log"
 )
 
 func main() {
 	err := foo()
 	if err != nil {
 		log.Printf("error: %+v", err)
 		return
 	}
 }
 
 func foo() error {
 	err := bar()
 	if err != nil {
 		log.Printf("error: %+v", err)
 		return err
 	}
 	return nil
 }
 
 func bar() error {
 	return errors.New("something bad happened")
 }
```

##### Preferred

```
package main

import (
	"errors"
	"fmt"
	"log"
)

func main() {
	err := foo()
	if err != nil {
		log.Printf("error: %+v", err)
		return
	}
}

func foo() error {
	err := bar()
	if err != nil {
		return fmt.Errorf("calling bar: %w", err)
	}
	return nil
}

func bar() error {
	return errors.New("something bad happened")
}
```


Do not use the github.com/pkg/errors package as it is now in maintenance mode since Go 1.13+ added official support for error wrapping.  See https://blog.golang.org/go1.13-errors and https://github.com/fatih/errwrap.

#### Gofmt
Run gofmt on your code to automatically fix the majority of mechanical style issues. Almost all Go code in the wild uses gofmt. The rest of this document addresses non-mechanical style points.

An alternative is to use goimports, a superset of gofmt which additionally adds (and removes) import lines as necessary.

A recommended approach is to ensure your editor supports running of goimports automatically on save.

# TODOs in the code
We don't like TODOs in the code. It is really best if you sort out all issues you can see with the changes before we check the changes in.

# Building, Deploying, and Testing
Please refer to the [Getting Started Guide](./GETTING_STARTED_GUIDE.md) for information on building, deploying, and running tests for Karavi PowerFlex Metrics.
