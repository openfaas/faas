## Contributing

### Guidelines

Guidelines for contributing.

**How can I get involved?**

First of all, we'd love to welcome you into our Slack community where we exchange ideas, ask questions and chat about OpenFaaS, Raspberry Pi and other cloud-native technology. (*See below for how to join*)

We have a number of areas where we can accept contributions:

* Write Golang code for the CLI, Gateway or other providers
* Write for our front-end UI (JS, HTML, CSS)
* Write sample functions in any language
* Review pull requests
* Test out new features or work-in-progress
* Get involved in design reviews and technical proof-of-concepts (PoCs)
* Help us release and package OpenFaaS including the helm chart, compose files, kubectl YAML, marketplaces and stores
* Manage, triage and research Issues and Pull Requests
* Help our growing community feel at home
* Create docs, guides and write blogs
* Speak at meet-ups, conferences or by helping folks with OpenFaaS on Slack

This is just a short list of ideas, if you have other ideas for contributing please make a suggestion.

**I've found a typo**

* A Pull Request is not necessary. Raise an [Issue](https://github.com/openfaas/faas/issues) and we'll fix it as soon as we can. 

**I have a [great] idea**

The OpenFaaS maintainers would like to make OpenFaaS the best it can be and welcome new contributions that align with the project's goals. Our time is limited so we'd like to make sure we agree on the proposed work before you spend time doing it. Saying "no" is hard which is why we'd rather say "yes" ahead of time.

What makes a good proposal?

* Brief summary including motivation/context
* Any design changes
* Pros + Cons
* Effort required
* Mock-up screenshots or examples of how the CLI would work

**Paperwork for Pull Requests**

Please read this whole guide and make sure you agree to our DCO agreement (included below):

* Sign-off your commits 
* Complete the whole template for issues and pull requests
* [Reference addressed issues](https://help.github.com/articles/closing-issues-using-keywords/) in the PR description & commit messages - use 'Fixes #IssueNo' 
* Always give instructions for testing
* Provide us CLI commands and output or screenshots where you can 

**Unit testing with Golang**

Please follow style guide on [this blog post](https://blog.alexellis.io/golang-writing-unit-tests/) from [The Go Programming Language](https://www.amazon.co.uk/Programming-Language-Addison-Wesley-Professional-Computing/dp/0134190440)

**I have a question, a suggestion or need help**

Please raise an Issue or email alex@openfaas.com for an invitation to our Slack community.

**I need to add a dependency**

We are using the `vndr` tool across all projects. Get [started here](https://github.com/LK4D4/vndr).

**How do I become a maintainer?**

Maintainers are well-known contributors who help with:
* Fixing, testing and triaging issues
* Joining contributor meetings and supporting new contributors
* Testing and reviewing pull requests
* Offering other project support and strategical advice

Varying levels of write access are made available via our project bot [Derek](https://github.com/alexellis/derek) to help regular contributors transition to maintainers.

**How do I work with Derek the bot?**

If you have been added to the MAINTAINERS file in the root of an OpenFaaS repository then you can help us manage our community and contributions by issuing comments on Issues and Pull Requests. See [Derek](https://github.com/alexellis/derek) for available commands.

**Governance**

OpenFaaS is an independent project created by Alex Ellis which is now being built by a growing community of contributors.

### Community

This project is written in Golang but many of the community contributions so far have been through blogging, speaking engagements, helping to test and drive the backlog of FaaS. If you'd like to help in any way then that would be more than welcome whatever your level of experience.

#### Community file

The [community.md](https://github.com/openfaas/faas/blob/master/community.md) file highlights blogs, talks and code repos with example FaaS functions and usages. Please send a Pull Request if you are doing something cool with FaaS.

#### Roadmap

Checkout the [roadmap](https://github.com/openfaas/faas/blob/master/ROADMAP.md) and [open issues](https://github.com/openfaas/faas/issues).

#### Slack

There is an Slack community which you are welcome to join to discuss FaaS, IoT and Raspberry Pi projects. Ping [Alex Ellis](https://github.com/alexellis) with your email address so that an invite can be sent out.

Email: alex@openfaas.com - please send in a one-liner about yourself so we can give you a warm welcome and help you get started.

### License

This project is licensed under the MIT License.

#### Copyright notice

Please add a Copyright notice to new files you add where this is not already present:

```
// Copyright (c) OpenFaaS Project 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.
```

#### Sign your work

> Note: all of the commits in your PR/Patch must be signed-off.

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

* Please sign your commits with `git commit -s` so that commits are traceable.
