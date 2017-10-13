### Api-Key-Protected sample

To use this sample provide a secret for the container/service in `secret_api_key` using [Docker Swarm Secret](https://docs.docker.com/engine/swarm/secrets/#defining-and-using-secrets-in-compose-files).

Then when calling via the gateway pass the additional header "X-Api-Key", if it matches the `secret_api_key` value then the function will give access, otherwise access denied.

