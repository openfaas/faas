basic-auth
============

This component implements [Basic Authentication](https://en.wikipedia.org/wiki/Basic_access_authentication) as an OpenFaaS authentication plug-in.

To run this plugin you will need to create and bind a secret named `basic-auth-user` and `basic-auth-password`

| Option                          | Usage             |
|---------------------------------|--------------|
| `port`                          | Set the HTTP port |
| `secret_mount_path`             | It is recommended that this is set to `/var/openfaas/secrets` |
| `user_filename`                 | File to read from disk for username, default empty |
| `pass_filename`                 | File to read from disk for username, default empty |
