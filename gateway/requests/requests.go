// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

// Package requests package provides a client SDK or library for
// the OpenFaaS gateway REST API
package requests

// DeleteFunctionRequest delete a deployed function
type DeleteFunctionRequest struct {
	FunctionName string `json:"functionName"`
}
