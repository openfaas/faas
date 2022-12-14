# Contributing

## Guidelines

Guidelines for contributing.

### First impressions - introducing yourself and your use-case

One of the best ways to participate within a new open source communities is to introduce yourself and your use-case. This builds goodwill, but also means the community can start to understand your needs and how best to help you.

Given that the community is made up of volunteers, making a good first impression is important to getting their ear and attention. 

Here is a simple introduction you could try:

> "We at Company Y have an issue with X, and would like some help. What we're trying to achieve is Z" 

The more context you can give the community, the more the community can be of help to you. If you are using OpenFaaS as a hobbyist or as a student, then please also let us know so that we can decide how to prioritise all the requests we receive from users. 

A common example of a poor introduction would be asking for technical support without providing any context, or introduction.

> "We are running into issue X. Can you fix it?"
> 
> "We also had this issue"

These kinds of interactions start with "we" and since we is a pronoun, it becomes an anonymous request detached from any context or relationship with the community. The fix is easy, just say who you are and what your interest is, and what your ideal outcome is.

The primary ways to engage with the community are via GitHub Issues and [Enterprise Support](https://openfaas.com/support/).

* GitHub Issues - for suspected bugs and feature requests, fill out the whole template. Do not use GitHub issues to ask for help with performance/load-testing and/or tuning, this is a professional service which you can get via Enterprise Support.
* Enterprise Support - you will have an agreed way to contact OpenFaaS Ltd for direct support and help

See also: [The no-excuses guide to introducing yourself to a new open source project](https://opensource.com/education/13/7/introduce-yourself-open-source-project)

### How can I get involved?

There are a number of areas where contributions can be accepted:

* Write Golang code for the CLI, Gateway or other providers
* Write features for the front-end UI (JS, HTML, CSS)
* Write sample functions in any language
* Review pull requests
* Test out new features or work-in-progress
* Get involved in design reviews and technical proof-of-concepts (PoCs)
* Help release and package OpenFaaS including the helm chart, compose files, `kubectl` YAML, marketplaces and stores
* Manage, triage and research Issues and Pull Requests
* Engage with the growing community by providing technical support on GitHub
* Create docs, guides and write blogs
* Speak at meet-ups, conferences or by asking where you can be of help

This is just a short list of ideas, if you have other ideas for contributing please make a suggestion.

### I want to contribute on GitHub

#### I've found a security issue

Please follow [responsible disclosure practices](https://en.wikipedia.org/wiki/Responsible_disclosure) and send an email to support@openfaas.com. Bear in mind that instructions on how to reproduce the issue are key to proving an issue exists, and getting it resolved. Suggested solutions are also weclome.

#### I've found a typo

* A Pull Request is not necessary. Raise an [Issue](https://github.com/openfaas/faas/issues) and we'll fix it as soon as we can. 

#### I have a (great) idea

The OpenFaaS maintainers would like to make OpenFaaS the best it can be and welcome new contributions that align with the project's goals. Our time is limited so we'd like to make sure we agree on the proposed work before you spend time doing it. Saying "no" is hard which is why we'd rather say "yes" ahead of time. You need to raise a proposal.

Every feature carries a cost - a cost if developed wrong, a cost to carry and maintain it and if it wasn't needed in the first place then this is an unnecessary burden. See [Yagni from Martin Fowler](https://martinfowler.com/bliki/Yagni.html). The best proposals are defensible with real data and are more than a hypothesis.

**Please do not raise a proposal after doing the work - this is counter to the spirit of the project. It is hard to be objective about something which has already been done**

What makes a good proposal?

* Brief summary including motivation/context
* Any design changes
* Pros + Cons
* Effort required up front
* Effort required for CI/CD, release, ongoing maintenance
* Migration strategy / backwards-compatibility
* Mock-up screenshots or examples of how the CLI would work
* Clear examples of how to reproduce any issue the proposal is addressing

Once your proposal receives a `design/approved` label you may go ahead and start work on your Pull Request.

If you are proposing a new tool or service please do due diligence. Does this tool already exist in a 3rd party project or library? Can we reuse it? For example: a timer / CRON-type scheduler for invoking functions is a well-solved problem, do we need to reinvent the wheel?

Every effort will be made to work with contributors who do not follow the process. Your PR may be closed or marked as `invalid` if it is left inactive, or the proposal cannot move into a `design/approved` status.

#### Paperwork for Pull Requests

Please read this whole guide and make sure you agree to the Developer Certificate of Origin (DCO) agreement (included below):

* See guidelines on commit messages (below)
* Sign-off your commits (`git commit --signoff` or `-s`)
* Complete the whole template for issues and pull requests
* [Reference addressed issues](https://help.github.com/articles/closing-issues-using-keywords/) in the PR description & commit messages - use 'Fixes #IssueNo' 
* Always give instructions for testing
 * Provide us CLI commands and output or screenshots where you can

##### Commit messages

The first line of the commit message is the *subject*, this should be followed by a blank line and then a message describing the intent and purpose of the commit. These guidelines are based upon a [post by Chris Beams](https://chris.beams.io/posts/git-commit/).

* When you run `git commit` make sure you sign-off the commit by typing `git commit --signoff` or `git commit -s`
* The commit subject-line should start with an uppercase letter
* The commit subject-line should not exceed 72 characters in length
* The commit subject-line should not end with punctuation (., etc)

> Note: please do not use the GitHub suggestions feature, since it will not allow your commits to be signed-off.

When giving a commit body:
* Leave a blank line after the subject-line
* Make sure all lines are wrapped to 72 characters

Here's an example that would be accepted:

```
Add alexellis to the .DEREK.yml file

We need to add alexellis to the .DEREK.yml file for project maintainer
duties.

Signed-off-by: Alex Ellis <alex@openfaas.com>
```

Some invalid examples:

```
(feat) Add page about X to documentation
```

> This example does not follow the convention by adding a custom scheme of `(feat)`

```
Update the documentation for page X so including fixing A, B, C and D and F.
```

> This example will be truncated in the GitHub UI and via `git log --oneline`


If you would like to ammend your commit follow this guide: [Git: Rewriting History](https://git-scm.com/book/en/v2/Git-Tools-Rewriting-History)

#### Unit testing with Golang

Please follow style guide on [this blog post](https://blog.alexellis.io/golang-writing-unit-tests/) from [The Go Programming Language](https://www.amazon.co.uk/Programming-Language-Addison-Wesley-Professional-Computing/dp/0134190440)

Specifically, the style means using Golang code to evaluate whether a tested method produced the correct result, for instance:

```golang
func TestSum(t *testing.T) {
    want := 10
    got := Sum(5, 5)
    if want != got {
       t.Fatal("want: %d, but got: %d, want, got)
    }
}
```

Making use of test tables, additional comparison libraries or helper functions is acceptable.

This kind of usage will not be merged into the codebase:

```golang
func TestSomething(t *testing.T) {
  assert := assert.New(t)

  // assert equality
  assert.Equal(Sum(5, 5), 10, "they should be equal")
```

Please do not introduce .NET/Java-style assertion libararies such as [stretchr/testify](https://github.com/stretchr/testify).

**A note on go-routines**

If you are making changes to code that use goroutines, consider adding `goleak` to your test to help ensure that we are not leaking any goroutines. Simply add

```go
defer goleak.VerifyNoLeaks(t)
```

at the very beginning of the test, and it will fail the test if it detects goroutines that were opened but never cleaned up at the end of the test.

#### I have a question, a suggestion or need help

If you have a deeply technical request or need help debugging your application then you should prepare a simple, public GitHub repository with the minimum amount of code required to reproduce the issue. 

If you feel there is an issue with OpenFaaS or were unable to get the help you needed from the GitHub, [then send us an email](https://openfaas.com/support/)

#### Setting expectations, support and SLAs

* What kind of support can I expect for free?

    OpenFaaS is licensed in a way that enables you to use the source code in or with your project or product.

    If you are using one of the Open Source projects within the openfaas or openfaas-incubator repository, then help may be offered on a limited, good-will basis by volunteers, but if you are a commercial user, you will need to purchase support for timely help.
    
    Please be respectful of any time given to you and your needs. The person you are requesting help from may not reside in your timezone and contacting them via direct message is inappropriate.

    Enterprise support is the best place to ask questions, suggest features, and to get help. The GitHub issue tracker can be used for suspected issues with the codebase or deployment artifacts. The whole template must be filled out in detail.

* Doesn't Open Source mean that everything is free?

    The OpenFaaS projects are licensed as MIT which means that you are free to use, modify and distribute the software within the terms of the license.
    
    Contributions, suggestions and feedback is welcomed in the appropriate channels as outlined in this guide. The MIT license does not cover support for PRs, Issues, Technical
    Support questions, feature requests and technical support/professional services which you may require; the preceding are not free and have a cost to those providing the services. Where possible, this time may be volunteered for free, but it is not unlimited.

* What is the SLA for my Issue?

    Issues are examined, triaged and answered on a best effort basis by volunteers and community contributors. This means that you may receive an initial response within any time period such as: 1 minute, 1 hour, 1 day, or 1 week. There is no implicit meaning to the time between you raising an issue and it being answered or resolved.

    If you see an issue which does not have a response or does not have a resolution, it does not mean that it is not important, or that it is being ignored. It simply means it has not been worked on by a volunteer yet.

    Please take responsibility for following up on your Issues if you feel further action is required.

    If you are a business using OpenFaaS and need timely and attentive responses, then you should purchase Enterprise Support from OpenFaaS Ltd.

* What is the SLA for my Pull Request?

    In a similar way to Issues, Pull Requests are triaged, reviewed, and considered by a team of volunteers - the Core Team,  Members Team and the Project Lead. There are dozens of components that make up the OpenFaaS project and a limited amount of people. Sometimes PRs may become blocked or require further action.
    
    Please take responsibility for following up on your Pull Requests if you feel further action is required.
    
* Why may your PR be delayed?

    * The contributing guide was not followed in some way

    * The commits are not signed-off (the Derek bot will try to help you)

    * The commits need to be rebased

    * Changes have been requested

    More information, a use-case, or context may be required for the change to be accepted.

* What if I am a GitHub Sponsor?

    If you [sponsor OpenFaaS on GitHub](https://github.com/sponsors/openfaas), then you will show up as a Sponsor on your issues and PRs which is one way to show your support for the community and project. Whilst the entry-level sponsorship is only 25 USD / mo, you will benefit from access to regular updates on project development via the [Treasure Trove portal](https://faasd.exit.openfaas.pro/function/trove/). Your company can also take up a GitHub Sponsorship using their GitHub organisation's existing billing relationship.

* What if I need more than that?

    If you're a company using any of these projects, you can get the following through an [Enterprise Support agreement with OpenFaaS Ltd](https://openfaas.com/support/) so that the time and resources required to support your business are paid for.

    A support agreement can be tailored to your needs, you may benefit from support, if you need any of the following:

    * security issues patched in a timely manner for all 40 +/- open source components
    * priority responses to issues/PRs
    * immediate help and access to experts

#### I need to add a dependency

All projects use [Go modules](https://github.com/golang/go/wiki/Modules) and vendoring. The concept of `vendoring` is still broadly used in projects written in Go. This means that a copy of the source-code of dependencies is stored within each repository in the `vendor` folder. It allows for a repeatable build and isolates change.

### How are releases made?

Releases are made by the *Project Lead* on a regular basis and when deemed necessary. If you want to request a new release then mention this on your PR or Issue.

Releases are cut with `git` tags and a successful Travis build results in new binary artifacts and Docker images being published to the Docker Hub and Quay.io. See the "Build" badge on each GitHub README file for more.

How are credentials managed for quay.io and the Docker Hub? These credentials are maintained by the *Project Lead*.

## Governance

OpenFaaS is an independent open-source project which was created by the Project Lead Alex Ellis in 2016. OpenFaaS is now hosted by OpenFaaS Ltd. The project is maintained and developed by a number of regular volunteers and a wider community of open-source developers.

OpenFaaS Ltd (company no. 11076587) hosts and sponsors the development and maintenance of OpenFaaS. OpenFaaS Ltd provides professional services, consultation and support. Email: [sales@openfaas.com](mailto:sales@openfaas.com) to find out more.

OpenFaaS &reg; is a registered trademark in England and Wales.

#### Project Lead

Responsibility for the project starts with the *Project Lead*, who delegates specific responsibilities and the corresponding authority to the Core and Members team.

Some duties include:

* Setting overall technical & community leadership
* Engaging end-user community to advocate needs of end-users and to capture case-studies
* Defining and curating roadmap for OpenFaaS & OpenFaaS Cloud
* Building a community and team of contributors
* Community & media briefings, out-bound communications, partnerships, relationship management and events

### How do I become a maintainer?

In the OpenFaaS community there are four levels of structure or maintainership:

* Core Team (GitHub org)
* Members Team (GitHub org)
* Those with Derek access
* The rest of the community.

#### Core Team

The Core Team have the ear of the Project Lead. They help with strategy, project maintenance, community management, and make a regular commitment of time to the project on a weekly basis.

Each member will be responsible for, or be a subject-matter-expert (SME) for a sub-system of OpenFaaS and will be granted write (push) access to the related repositories.

The Core Team have the same responsibilities and perks of the Membership Team, in addition will need to keep in close contact with the rest of the Core Team and the Project Lead.

* Members are listed on the project homepage as being part of the Core group and are shown first.
* Members are expected to attend 1:1 Zoom calls with the Project Lead up to once per month
* Members will notify the Project Lead and Core Team of any leave of a week or more and set a status in Slack of "away".

Core Team attend all project meetings and calls. Allowances will be made for timezones and other commitments.

The Core Team includes:

- Alex Ellis (@alexellis) - Lead
- Lucas Roesler (@LucasRoesler) - SME for logs, provider model and secrets

#### Members Team

The Members Team are contributors who are well-known to the community with a track record of:

* fixing, testing and triaging issues and PRs
* offering support to the project
* providing feedback and being available to help where needed
* testing and reviewing pull requests
* joining contributor meetings and supporting new contributors

> Note: An essential skill for being in a team is communication. If it is not possible to communicate on a regular basis then, then membership may not be for you. You are welcome to contribute as part of the wider community.

Varying levels of write access are made available via the project bot [Derek](https://github.com/alexellis/derek) to help regular contributors transition to the Members Team.

Members Team Perks:
* access to a private Slack channel
* profile posted on the Team page of the OpenFaaS website
* membership of the GitHub organisations openfaas/openfaas-incubator

Upon request and subject to availability:
* 1:1 coaching & mentorship
* help with speaking opportunities and CfP submissions
* help with CV, resume and LinkedIn profile
* review, and promotion of blogs and tutorials on social media

The Members Team are expected to:

* participate in the members channel and engage with the topics
* participate in community Zoom calls (when possible within your timezone)
* make regular contributions to the project codebase
* take an active role in the public channels: #contributors and #openfaas
* comment on and engage with project proposals
* attend occasional 1:1 meetings with members of the Core Team or the Project Lead

This group is intended to be an active team that shares the load and collaborates together. This means engaging in topics on Slack, working with other teammates, sharing ideas, helping the users and raising issues with the Core Team.

The Members Team will notify their team in the *members* channel about any planned leave of a week or more and set a status in Slack of "away".

#### Changing teams

Every contributor to OpenFaaS is a volunteer, including the *Project Lead* and nobody is paid to work on OpenFaaS.

Motivations and life-circumstances can change over time. If this is expected to be a short-term change, then speak to the *Project Lead* about a sabbatical arrangement with perks and membership retained for that time.

You may move from the Core Team to the Members Team. Please notify the *Project Lead*.

If you can no-longer commit to being part of a team, then you may move to Community Contributor status and retain your access to Derek for as long as it is useful to you.

#### Stepping-down and emeritus status

> emeritus: (of the former holder of an office, especially a university professor) having retired but allowed to retain their title as an honour.

Some guidelines on stepping down:

> When somebody leaves or disengages from the project, we ask that they do so in a way that minimises disruption to the project. They should tell [*The Project Lead*, that] they are leaving and take the proper steps to ensure that others can pick up where they left off.

Quoted from the [Ubuntu community guidelines](https://ubuntu.com/community/code-of-conduct).

It's reasonable to expect that some people may no longer be able to continue their Open Source contributions actively, but would like to remain a part of the project and to continue to be recognised.

#### Access to Derek

If you have been added to the `.DEREK.yml` file in the root of an OpenFaaS repository then you can help us manage the community and contributions by issuing comments on Issues and Pull Requests. See [Derek](https://github.com/alexellis/derek) for available commands.

If you are a contributor then you are welcome to request access.

## Branding guidelines

For press, branding, logos and marks see the [OpenFaaS media repository](https://github.com/openfaas/media).

## Community

This project is written in Golang but many of the community contributions so far have been through blogging, speaking engagements, helping to test and drive the backlog of OpenFaaS. If you'd like to help in any way then that would be more than welcome whatever your level of experience.

### Community file

The [community.md](https://github.com/openfaas/faas/blob/master/community.md) file highlights blogs, talks and code repos with example FaaS functions and usages. Please send a Pull Request if you are doing something cool with OpenFaaS.

### Roadmap

See also: [OpenFaaS Pro](https://docs.openfaas.com/openfaas-pro/introduction/)

## License

This project is licensed under the MIT License.

### Copyright notice

It is important to state that you retain copyright for your contributions, but agree to license them for usage by the project and author(s) under the MIT license. Git retains history of authorship, but we use a catch-all statement rather than individual names. 

Please add a Copyright notice to new files you add where this is not already present.

```
// Copyright (c) OpenFaaS Author(s) 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.
```

### Sign your work

> Note: every commit in your PR or Patch must be signed-off.

The sign-off is a simple line at the end of the explanation for a patch. Your
signature certifies that you wrote the patch or otherwise have the right to pass
it on as an open-source patch. The rules are pretty simple: if you can certify
the below (from [developercertificate.org](http://developercertificate.org/)):

```
Developer Certificate of Origin
Version 1.1

Copyright (C) 2004, 2006 The Linux Foundation and its contributors.
1 Letterman Drive
Suite D4700
San Francisco, CA, 94129

Everyone is permitted to copy and distribute verbatim copies of this
license document, but changing it is not allowed.

Developer's Certificate of Origin 1.1

By making a contribution to this project, I certify that:

(a) The contribution was created in whole or in part by me and I
    have the right to submit it under the open source license
    indicated in the file; or

(b) The contribution is based upon previous work that, to the best
    of my knowledge, is covered under an appropriate open source
    license and I have the right under that license to submit that
    work with modifications, whether created in whole or in part
    by me, under the same open source license (unless I am
    permitted to submit under a different license), as indicated
    in the file; or

(c) The contribution was provided directly to me by some other
    person who certified (a), (b) or (c) and I have not modified
    it.

(d) I understand and agree that this project and the contribution
    are public and that a record of the contribution (including all
    personal information I submit with it, including my sign-off) is
    maintained indefinitely and may be redistributed consistent with
    this project or the open source license(s) involved.
```

Then you just add a line to every git commit message:

    Signed-off-by: Joe Smith <joe.smith@email.com>

Use your real name (sorry, no pseudonyms or anonymous contributions.)

If you set your `user.name` and `user.email` git configs, you can sign your
commit automatically with `git commit -s`.

Please sign your commits with `git commit -s` so that commits are traceable.

This is different from digital signing using GPG, GPG is not required for 
making contributions to the project. 

If you forgot to sign your work and want to fix that, see the following 
guide: [Git: Rewriting History](https://git-scm.com/book/en/v2/Git-Tools-Rewriting-History)

