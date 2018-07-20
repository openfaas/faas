# Integration testing

These tests should be run against the sample stack included in the repository root.

## Deploy the stack
```
./deploy_stack.sh
faas-cli deploy -f ./stack.yml
```

## Remove the stack
1. Delete all OpenFaaS deployed functions
```
faas-cli remove
docker stack rm func
```