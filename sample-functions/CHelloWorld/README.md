Hello World in C
===================

This is hello world in C using GCC and Alpine Linux.

It also makes use of a multi-stage build and a `scratch` container for the runtime.

```
$ faas-cli build -f ./stack.yml
```

If pushing to a remote registry change the name from `alexellis` to your own Hub account.

```
$ faas-cli push -f ./stack.yml
$ faas-cli deploy -f ./stack.yml
```

Then invoke via `curl`, `faas-cli` or the UI.

