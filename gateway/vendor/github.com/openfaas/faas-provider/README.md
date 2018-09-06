faas-provider
==============

This is a common template or interface for you to start building your own OpenFaaS backend.

Checkout the [backends guide here](https://github.com/openfaas/faas/blob/master/guide/backends.md) before starting.

OpenFaaS projects use the MIT License and are written in Golang. We encourage the same for external / third-party providers.

### How to use this code

We will setup all the standard HTTP routes for you, then start listening on a given TCP port - it should be 8080.

Just implement the supplied routes.

For an example checkout the [server.go](https://github.com/openfaas/faas-netes/blob/master/server.go) file in the [faas-netes](https://github.com/openfaas/faas-netes) Kubernetes backend.

I.e.:

```golang
	bootstrapHandlers := bootTypes.FaaSHandlers{
		FunctionProxy:  handlers.MakeProxy(),
		DeleteHandler:  handlers.MakeDeleteHandler(clientset),
		DeployHandler:  handlers.MakeDeployHandler(clientset),
		FunctionReader: handlers.MakeFunctionReader(clientset),
		ReplicaReader:  handlers.MakeReplicaReader(clientset),
		ReplicaUpdater: handlers.MakeReplicaUpdater(clientset),
		InfoHandler:    handlers.MakeInfoHandler(),
	}
	var port int
	port = 8080
	bootstrapConfig := bootTypes.FaaSConfig{
		ReadTimeout:  time.Second * 8,
		WriteTimeout: time.Second * 8,
		TCPPort:      &port,
	}

	bootstrap.Serve(&bootstrapHandlers, &bootstrapConfig)
```
