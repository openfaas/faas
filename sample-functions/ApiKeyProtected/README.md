### Api-Key-Protected sample

See the [secure secret management guide](../guide/secure_secret_management.md) for instructions on how to use this function.

When calling via the gateway pass the additional header "X-Api-Key", if it matches the `secret_api_key` value then the function will give access, otherwise access denied.

