# How to Contribute

Become one of the contributors to this project! We thrive to build a welcoming and open
community for anyone who wants to use the project or contribute to it. There are
just a few small guidelines you need to follow. 

# Table of Content:
* [Introduction](#Introduction)
* [Issues](#Issues)
* [Branching Strategy](#Branching-strategy)
* [Code reviews](#Code-reviews)
* [Signing your commits](#Signing-your-commits)
* [Code Style](#Code-Style)
* [TODOs in the code](#TODOs-in-the-code)
* [Testing](#Testing)
* [I/O Workload](#IO-Workload)

## Introduction

Please contribute! 
Read the below documentation to understand the coding standards and expected practices of this project.

## Issues

We greatly value your feedback! If you have any questions, wish to report a bug, or request a new feature, use our project GitHub Issues page.

We aim to track and document everything related to this repo via the Issues page. The code and documentation are released with no warranties or SLAs and are intended to be supported through a community driven process.

When opening a Bug please include the following information to help with debugging:

1. Version of relevant software: this software, Kubernetes, Dell Storage Platform, Helm, etc.
2. Details of the issue explaining the problem: what, when, where
3. The expected outcome that was not met (if any)
4. Supporting troubleshooting information. __Note: Do not provide private company information that could compromise your company's security.__

An Issue __must__ be created before submitting any pull request. Any pull request that is created should be linked to an Issue.

## Branching Strategy
We are following a scaled trunk branching strategy where short-lived branches are created off of the master branch. When coding is complete, the branch is merged back into master after being approved in a pull request code review.

Steps to create a branch for a bug fix or feature:
1. Fork the repository.
2. Create a branch off of the master branch. The branch name should be descriptive and include the bug fix or feature that it contains.
3. Write code, add tests, and commit to your branch. Optionally, add feature flags to disable any new features that are not yet ready for the release.
4. If other code changes have merged into the upstream master branch, perform a rebase of those changes into your branch.
5. Open a pull request between your branch and the upstream master branch.
6. Once your pull request has merged, your branch can be deleted.

Release branches will be created from the master branch near the time of a planned release when all features are completed. Only critical bug fixes will be merged into the feature branch at this time. All other bug fixes and features can continue to be merged into the master branch. When a feature branch is stable, the branch will be tagged for release and the release branch will be deleted.

## Code Reviews

All submissions, including submissions by project members, require review. We
use GitHub pull requests for this purpose. Consult
[GitHub Help](https://help.github.com/articles/about-pull-requests/) for more
information on using pull requests.

## Signing your commits

We require that developers sign off their commits to certify that they have permission to contribute the code in a pull request. This way of certifying is commonly known as the [Developer Certificate of Origin (DCO)](https://developercertificate.org/). We encourage all contributors to read the DCO text before signing a commit and making contributions.

To make sure that pull requests have all commits signed off, we use the Probot DCO plugin.

### Signing off a commit

Using the git command line
Use either `--signoff` or `-s` with the commit command. See this [document](https://probot.github.io/apps/dco/) for an example.

Make sure you have your user name and e-mail set. The `--signoff` or `-s` option will use the configured user name and e-mail, so it is important to configure it before the first time you commit. Check the following references:

* Setting up your github user name [reference](https://help.github.com/articles/setting-your-username-in-git/)
* Setting up your e-mail address [reference](https://help.github.com/articles/setting-your-commit-email-address-in-git/)

In practice, just add a line to every git commit message:

Signed-off-by: Jane Smith jane.smith@example.com Use your real name (sorry, no pseudonyms or anonymous contributions).

## Code Style

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

## TODOs in the code
We don't like TODOs in the code. It is really best if you sort out all issues you can see with the changes before we check the changes in.

## Testing
Unit tests can be run by executing `make test` from the root directory.

### I/O Workload
Scripts for executing an I/O workload against StorageClasses can be found in the [I/O Workload](./testing/IO_WORKLOAD.md) section.

