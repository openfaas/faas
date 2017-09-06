## Contributing

### Guidelines

Guidelines for contributing.

**I've found a typo**

* A Pull Request is not necessary. Raise an [Issue](https://github.com/alexellis/faas/issues) and we'll fix it as soon as we can. 

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
* Always give instructions for testing
 * Give us CLI commands and output or screenshots where you can 

**I have a question, a suggestion or need help**

Please raise an Issue or email alex@openfaas.com for an invitation to our Slack community.

**I need to add a dependency**

We are using the `vndr` tool across all projects. Get [started here](https://github.com/LK4D4/vndr).

### Community

This project is written in Golang but many of the community contributions so far have been through blogging, speaking engagements, helping to test and drive the backlog of FaaS. If you'd like to help in any way then that would be more than welcome whatever your level of experience.

#### Community file

The [community.md](https://github.com/alexellis/faas/blob/master/community.md) file highlights blogs, talks and code repos with example FaaS functions and usages. Please send a Pull Request if you are doing something cool with FaaS.

#### Roadmap

Checkout the [roadmap](https://github.com/alexellis/faas/blob/master/ROADMAP.md) and [open issues](https://github.com/alexellis/faas/issues).

#### Slack

There is an Slack community which you are welcome to join to discuss FaaS, IoT and Raspberry Pi projects. Ping [Alex Ellis](https://github.com/alexellis) with your email address so that an invite can be sent out.

### License

This project is licensed under the MIT License.

#### Sign your work

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
