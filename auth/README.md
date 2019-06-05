auth plugins
============

Auth plugins must implement request checking on a HTTP port and path such as `:8080/validate`.

* Valid requests: return 2xx
* Invalid requests: return non 2xx

It is up to the developer to pick whether a request body is required for validation. For strategies such as [Basic Authentication](https://en.wikipedia.org/wiki/Basic_access_authentication), headers are sufficient.

Plugins available:

* [basic-auth](./basic-auth/)
