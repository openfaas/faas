# Microsoft Azure SDK for Go

This project provides various Go packages to perform operations
on Microsoft Azure REST APIs.

[![GoDoc](https://godoc.org/github.com/Azure/azure-sdk-for-go?status.svg)](https://godoc.org/github.com/Azure/azure-sdk-for-go) [![Build Status](https://travis-ci.org/Azure/azure-sdk-for-go.svg?branch=master)](https://travis-ci.org/Azure/azure-sdk-for-go)

> **NOTE:** This repository is under heavy ongoing development and
is likely to break over time. We currently do not have any releases
yet. If you are planning to use the repository, please consider vendoring
the packages in your project and update them when a stable tag is out.

# Packages

## Azure Resource Manager (ARM)

[About ARM](/arm/README.md)

- [analysisservices](/arm/analysisservices)
- [authorization](/arm/authorization)
- [batch](/arm/batch)
- [cdn](/arm/cdn)
- [cognitiveservices](/arm/cognitiveservices)
- [commerce](/arm/commerce)
- [compute](/arm/compute)
- [containerregistry](/arm/containerregistry)
- [containerservice](/arm/containerservice)
- [datalake-analytics/account](/arm/datalake-analytics/account)
- [datalake-store/account](/arm/datalake-store/account)
- [devtestlabs](/arm/devtestlabs)
- [dns](/arm/dns)
- [documentdb](/arm/documentdb)
- [eventhub](/arm/eventhub)
- [intune](/arm/intune)
- [iothub](/arm/iothub)
- [keyvault](/arm/keyvault)
- [logic](/arm/logic)
- [machinelearning/commitmentplans](/arm/machinelearning/commitmentplans)
- [machinelearning/webservices](/arm/machinelearning/webservices)
- [mediaservices](/arm/mediaservices)
- [mobileengagement](/arm/mobileengagement)
- [network](/arm/network)
- [notificationhubs](/arm/notificationhubs)
- [powerbiembedded](/arm/powerbiembedded)
- [recoveryservices](/arm/recoveryservices)
- [redis](/arm/redis)
- [resources/features](/arm/resources/features)
- [resources/links](/arm/resources/links)
- [resources/locks](/arm/resources/locks)
- [resources/policy](/arm/resources/policy)
- [resources/resources](/arm/resources/resources)
- [resources/subscriptions](/arm/resources/subscriptions)
- [scheduler](/arm/scheduler)
- [search](/arm/search)
- [servermanagement](/arm/servermanagement)
- [servicebus](/arm/servicebus)
- [sql](/arm/sql)
- [storage](/arm/storage)
- [trafficmanager](/arm/trafficmanager)
- [web](/arm/web)

## Azure Service Management (ASM), aka classic deployment

[About ASM](/management/README.md)

- [affinitygroup](/management/affinitygroup)
- [hostedservice](/management/hostedservice)
- [location](/management/location)
- [networksecuritygroup](/management/networksecuritygroup)
- [osimage](/management/osimage)
- [sql](/management/sql)
- [storageservice](/management/storageservice)
- [virtualmachine](/management/virtualmachine)
- [virtualmachinedisk](/management/virtualmachinedisk)
- [virtualmachineimage](/management/virtualmachineimage)
- [virtualnetwork](/management/virtualnetwork)
- [vmutils](/management/vmutils)

## Azure Storage SDK for Go

[About Storage](/storage/README.md)

- [storage](/storage)

# Installation

- [Install Go 1.7](https://golang.org/dl/).

- Go get the SDK:

```
$ go get -d github.com/Azure/azure-sdk-for-go
```

> **IMPORTANT:** We highly suggest vendoring Azure SDK for Go as a dependency. For vendoring dependencies, Azure SDK for Go uses [glide](https://github.com/Masterminds/glide). If you haven't already, install glide. Navigate to your project directory and install the dependencies.

```
$ cd your/project
$ glide create
$ glide install
```

# Documentation

Read the Godoc of the repository at [Godoc.org](http://godoc.org/github.com/Azure/azure-sdk-for-go/).

# Contribute

If you would like to become an active contributor to this project please follow the instructions provided in [Microsoft Azure Projects Contribution Guidelines](http://azure.github.io/guidelines/).

# License

This project is published under [Apache 2.0 License](LICENSE).

-----
This project has adopted the [Microsoft Open Source Code of Conduct](https://opensource.microsoft.com/codeofconduct/). For more information see the [Code of Conduct FAQ](https://opensource.microsoft.com/codeofconduct/faq/) or contact [opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional questions or comments.
