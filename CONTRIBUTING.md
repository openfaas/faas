# Contributing

## Guidelines

Guidelines for contributing.

### How can I get involved?

The Slack community is the best place to keep up to date with the project and to get help contributing. Here we exchange ideas, ask questions and chat about OpenFaaS. There are also channels for Raspberry Pi/ARM, Kubernetes and other cloud-native topics. (*See below for how to join*)

There are a number of areas where contributions can be accepted:

* Write Golang code for the CLI, Gateway or other providers
* Write features for the front-end UI (JS, HTML, CSS)
* Write sample functions in any language
* Review pull requests
* Test out new features or work-in-progress
* Get involved in design reviews and technical proof-of-concepts (PoCs)
* Help release and package OpenFaaS including the helm chart, compose files, `kubectl` YAML, marketplaces and stores
* Manage, triage and research Issues and Pull Requests
* Engage with the growing community by providing technical support on Slack/GitHub
* Create docs, guides and write blogs
* Speak at meet-ups, conferences or by helping folks with OpenFaaS on Slack

This is just a short list of ideas, if you have other ideas for contributing please make a suggestion.

### I want to contribute on GitHub

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

If you are making changes to code that use goroutines, consider adding `goleak` to your test to help ensure that we are not leaking any goroutines. Simply add

```go
defer goleak.VerifyNoLeaks(t)
```

at the very beginning of the test, and it will fail the test if it detects goroutines that were opened but never cleaned up at the end of the test.

#### I have a question, a suggestion or need help

If you have a simple question you can [join the Slack community](https://docs.openfaas.com/community) and ask there, but please bear in mind that contributors may live in a different timezone or be working to a different timeline to you. If you have an urgent request then let them know about this.

If you have a deeply technical request or need help debugging your application then you should prepare a simple, public GitHub repository with the minimum amount of code required to reproduce the issue. 

If you feel there is an issue with OpenFaaS or were unable to get the help you needed from the Slack channels then raise an issue on one of the GitHub repositories.

#### Setting expectations, support and SLAs

* What kind of support can I expect for free?

    If you are using one of the Open Source projects within the openfaas or openfaas-incubator repository, then help is offered on a good-will basis by volunteers. You can also request help from employees of OpenFaaS Ltd who host the OpenFaaS projects.

    Please be respectful of volunteer time, it is often limited to evenings and weekends. The person you are requesting help from may not reside in your timezone. 

    The Slack workspace is the best place to ask questions, suggest features, and to get help. The GitHub issue tracker can be used for suspected issues with the codebase or deployment artifacts.

* Doesn't Open Source mean that everything is free?

    The OpenFaaS projects are licensed as MIT which means that you are free to use, modify and distribute the software within the terms of the license.
    
    Contributions, suggestions and feedback is welcomed in the appropriate channels as outlined in this guide. The MIT license does not cover support for PRs, Issues, Slack questions, feature requests and technical support/professional services which you may require; the preceding are not free and have a cost to those providing the services. Where possible, this time may be volunteered for free, but it is not unlimited.

* What is the SLA for my Issue?

    Issues are examined, triaged and answered on a best effort basis by volunteers and community contributors. This means that you may receive an initial response within any time period such as: 1 minute, 1 hour, 1 day, or 1 week. There is no implicit meaning to the time between you raising an issue and it being answered or resolved.

    If you see an issue which does not have a response or does not have a resolution, it does not mean that it is not important, or that it is being ignored. It simply means it has not been worked on by a volunteer yet.

    Please take responsibility for following up on your Issues if you feel further action is required.

* What is the SLA for my Pull Request?

    In a similar way to Issues, Pull Requests are triaged, reviewed, and considered by a team of volunteers - the Core Team,  Members Team and the Project Lead. There are dozens of components that make up the OpenFaaS project and a limited amount of people. Sometimes PRs may become blocked or require further action.
    
    Please take responsibility for following up on your Pull Requests if you feel further action is required.
    
* Why may your PR be delayed?

    * The contributing guide was not followed in some way

    * The commits are not signed-off (the Derek bot will try to help you)

    * The commits need to be rebased

    * Changes have been requested

    More information, a use-case, or context may be required for the change to be accepted.

* What if I need more than that?

    If you're a company using any of these projects, you can get the following through a support agreement with OpenFaaS Ltd so that the time can be paid for to help your business.

    A support agreement can be tailored to your needs, you may benefit from support, if you need any of the following:

    * responses within N hours/days on issues/PRs
    * feature prioritisation
    * urgent help
    * 1:1 consultations
    * or any other level of professional services

#### I need to add a dependency

The concept of `vendoring` is used in projects written in Go. This means that a copy of the source-code of dependencies is stored within each repository in the `vendor` folder. It allows for a repeatable build and isolates change.

The chosen tool for vendoring code in the project is [dep](https://github.com/golang/dep).

> Note: despite the availability of [Go modules](https://github.com/golang/go/wiki/Modules) in Go 1.11, they are not being used in the project at this time. If and when the decision is made to move, a complete overhaul of all repositories will need to be made in a coordinated fashion including: CI, releases, unit-, integration- and e2e-tests. This is not a trivial amount of effort.

**An update on Go modules*

A limited effort is being made by the maintainers to move some projects to Go Modules and Go 1.13. This approach will retain the use of the vendor folder, which is a valid strategy for adopting Go modules and is [documented in the Go wiki](https://github.com/golang/go/wiki/Modules). There are different trade-offs between using a vendor folder and not. If you do not understand these trade-offs well, you may explore the topic at your own leisure. We are not inviting the broader community to submit PRs to convert additional projects at this time.

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

The Core Team includes:

- Alex Ellis (@alexellis) - Lead
- Lucas Roesler (@LucasRoesler) - SME for logs, provider model and secrets

The Core Team have the ear of the Project Lead. They help with strategy, project maintenance, community management, and make a regular commitment of time to the project on a weekly basis. The Core Team will usually be responsible for, or be a subject-matter-expert (SME) for a sub-system of OpenFaaS. Core Team may be granted write (push) access to one or more sub-systems.

The Core Team gain access to a private *core* channel and are required to participate on a regular basis.

The Core Team have the same responsibilities and perks of the Membership Team, in addition will need to keep in close contact with the rest of the Core Team and the Project Lead.

* Core Team are expected to attend 1:1 Zoom calls with the Project Lead up to once per month
* Core Team members will notify the Project Lead and Core Team of any leave of a week or more and set a status in Slack of "away".

Core Team attend all project meetings and calls. Allowances will be made for timezones and other commitments.

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

After stepping-down, you may request to have an [emeritus status](https://www.lexico.com/en/definition/emeritus) and to be listed in the [BACKERS.md](/BACKERS.md) file for a period of time.

> emeritus: (of the former holder of an office, especially a university professor) having retired but allowed to retain their title as an honour.

Some guidelines on stepping down:

> When somebody leaves or disengages from the project, we ask that they do so in a way that minimises disruption to the project. They should tell [*The Project Lead*, that] they are leaving and take the proper steps to ensure that others can pick up where they left off.

Quoted from the [Ubuntu community guidelines](https://ubuntu.com/community/code-of-conduct).

It's reasonable to expect that some people may no longer be able to continue their Open Source contributions actively, but would like to remain a part of the project and to continue to be recognised.

#### Access to Derek

If you have been added to the `.DEREK.yml` file in the root of an OpenFaaS repository then you can help us manage the community and contributions by issuing comments on Issues and Pull Requests. See [Derek](https://github.com/alexellis/derek) for available commands.

If you are a contributor then you are welcome to request access.

#### Community/project meetings and calls

The community calls are held on Zoom on a regular basis with invitations sent out via email ahead of time.

General format:

- Project updates/briefing
- Round-table intros/updates
- Demos of features/new work from community
- Q&A

If you would like invites, sign-up to Slack and pick "Yes" to Community Events and Updates.

## Branding guidelines

For press, branding, logos and marks see the [OpenFaaS media repository](https://github.com/openfaas/media).

## Community

This project is written in Golang but many of the community contributions so far have been through blogging, speaking engagements, helping to test and drive the backlog of OpenFaaS. If you'd like to help in any way then that would be more than welcome whatever your level of experience.

### Community file

The [community.md](https://github.com/openfaas/faas/blob/master/community.md) file highlights blogs, talks and code repos with example FaaS functions and usages. Please send a Pull Request if you are doing something cool with OpenFaaS.

### Slack

There is an Slack community which you are welcome to join to discuss OpenFaaS, OpenFaaS Cloud, Kubernetes, Serverless, FaaS, IoT, and ARM / Raspberry Pi.

[Join Slack here](https://docs.openfaas.com/community/)

Getting the best out of Slack

* Explore the channels and post your question or comment into one of them
* Don't post the same question into multiple channels if you don't get an answer immediately
* Slack is asynchronous and you may not receive an answer to your question. Commercial support is available, but you are also welcome to ask the question again or in a different way after a reasonable amount of time has passed.
* Shorter, specific questions are more likely to be answered than longer bodies of text
* If you can share a code example, it may improve the chances of getting help. There is no need to share confidential code, instead try to create a small sample that reproduces the issue.

* Please be kind to your other participants within the community and bear in mind that English may not be everybody's first language.
* Moderators reserve the right to remove messages which may be deemed to be antagonistic, lacking context, a plug for a product, or off-topic. 

### Roadmap

* See the [2019 Project Update](https://www.openfaas.com/blog/project-update/)

* Browse open issues in [openfaas/faas](https://github.com/openfaas/faas/issues)

* Join the [2020 Roadmap on Trello](https://trello.com/invite/b/5OpMyrBP/ade103a10ae1e38eb5d3eee7955260a9/2020-openfaas-roadmap)

For commercial users, please feel free to ask about support, backlog prioritisation and feature development. Email sales@openfaas.com.

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

