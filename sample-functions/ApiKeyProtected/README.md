### Api-Key-Protected sample

To use this sample provide an env variable for the container/service in `secret_api_key`.

Then when calling via the gateway pass the additional header "X-Api-Key", if it matches the `secret_api_key` value then the function will give access, otherwise access denied.

