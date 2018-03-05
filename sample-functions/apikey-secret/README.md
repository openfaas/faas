### Sample: apikey-secret

This function returns access denied, or unlocked depending on whether your header for X-Api-Key matches a secret in the cluster called `secret_api_key`.

See the [secure secret management guide](../guide/secure_secret_management.md) for more information on secrets.

## Trying the sample:

```

$ docker secret remove secret_api_key  # make sure we delete any existing secret

# Create a secret with Swarm
$ echo "secret_value_goes_here" | docker secret create secret_api_key

# Deploy this sample with Docker Swarm and attach the secret to it

$ cd faas/sample-functions/
$ faas-cli deploy --filter apikey-secret --secret secret_api_key

# Now invoke the function with a good value:

$ echo -n | faas invoke --header "X-Api-Key=secret_value_goes_here" apikey-secret
You unlocked the function.

# Now invoke with a bad value:

echo -n | faas invoke --header "X-Api-Key=wrong_secret_value_goes_here" apikey-secret
Access was denied.

```