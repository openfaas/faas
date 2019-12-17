"use strict"
// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

var app = angular.module('faasGateway', ['ngMaterial', 'ngMessages', 'faasGateway.funcStore']);

app.controller("home", ['$scope', '$log', '$http', '$location', '$interval', '$filter', '$mdDialog', '$mdToast', '$mdSidenav',
    function($scope, $log, $http, $location, $interval, $filter, $mdDialog, $mdToast, $mdSidenav) {
        var FUNCSTORE_DEPLOY_TAB_INDEX = 0;
        var CUSTOM_DEPLOY_TAB_INDEX = 1;

        var newFuncTabIdx = FUNCSTORE_DEPLOY_TAB_INDEX;
        $scope.functions = [];
        $scope.invocationInProgress = false;
        $scope.invocationRequest = "";
        $scope.invocationResponse = "";
        $scope.invocationStatus = "";
        $scope.invocationStart = new Date().getTime();
        $scope.roundTripDuration = "";
        $scope.selectedNamespace = "";
        $scope.allNamespaces = [""];
        $scope.invocation = {
            contentType: "text"
        };

        $scope.baseUrl = $location.absUrl().replace(/\ui\/$/, '');

        try {
            $scope.canCopyToClipboard = document.queryCommandSupported('copy');
        } catch (err) {
            console.error(err);
            $scope.canCopyToClipboard = false;
        }
        $scope.copyClicked = function(e) {
            e.target.parentElement.querySelector('input').select()
            var copySuccessful = false;
            try {
                copySuccessful = document.execCommand('copy');
            } catch (err) {
                console.error(err);
            }
            var msg = copySuccessful ? 'Copied to Clipboard' : 'Copy failed. Please copy it manually';
            showPostInvokedToast(msg);
        }



        $scope.toggleSideNav = function() {
            $mdSidenav('left').toggle();
        };

        $scope.functionTemplate = {
            image: "",
            envProcess: "",
            network: "",
            service: "",
            envVars: {},
            labels: {},
            annotations: {},
            secrets: [],
            namespace: "",
        };

        $scope.invocation.request = "";

        var fetchFunctionsDelay = 3500;
        var queryFunctionDelay = 2500;




        $http.get("../system/namespaces")
            .then(function(response) {
                $scope.allNamespaces = response.data;
                if ($scope.selectedNamespace === "" && response.data && response.data[0] && response.data[0] !== "") {
                    $scope.selectedNamespace = response.data[0]
                    refreshData($scope.selectedNamespace)
                }
            })


        var fetchFunctionsInterval = $interval(function() {
            refreshData($scope.selectedNamespace);
        }, fetchFunctionsDelay);

        var queryFunctionInterval = $interval(function() {
            if($scope.selectedFunction && $scope.selectedFunction.name) {
                refreshFunction($scope.selectedFunction);
            }
        }, queryFunctionDelay);

        $scope.setNamespace = function(namespace) {
            $scope.selectedNamespace = namespace;
            $scope.selectedFunction = undefined;
            $scope.functions = [];
            refreshData($scope.selectedNamespace)
        }

        var refreshFunction = function(functionInstance) {
            const url = "/system/function/" + functionInstance.name;
            $http.get(buildNamespaceAwareURL(url, $scope.selectedNamespace))
            .then(function(response) {
               $scope.selectedFunction.ready = (response.data && response.data.availableReplicas && response.data.availableReplicas > 0);
            })
            .catch(function(err) {
                console.error(err);
            });
        };

        var showPostInvokedToast = function(message, duration) {
            $mdToast.show(
                $mdToast.simple()
                .textContent(message)
                .position("top right")
                .hideDelay(duration || 500)
            );
        };

        $scope.fireRequest = function() {
            var requestContentType = $scope.invocation.contentType == "json" ? "application/json" : "text/plain";
            if ($scope.invocation.contentType == "binary") {
                requestContentType = "binary/octet-stream";
            }

            var fnNamespace = ($scope.selectedNamespace && $scope.selectedNamespace.length > 0) ? "." + $scope.selectedNamespace : "";
            var options = {
                url: "../function/" + $scope.selectedFunction.name + fnNamespace,
                data: $scope.invocation.request,
                method: "POST",
                headers: { "Content-Type": requestContentType },
                responseType: $scope.invocation.contentType == "binary" ? "arraybuffer" : $scope.invocation.contentType
            };

            $scope.invocationInProgress = true;
            $scope.invocationResponse = "";
            $scope.invocationStatus = null;
            $scope.roundTripDuration = "";
            $scope.invocationStart = new Date().getTime()


            var tryDownload = function(data, filename) {
                var caught;

                try {
                    var blob = new Blob([data], { type: "binary/octet-stream" });

                    if (window.navigator.msSaveBlob) { // // IE hack; see http://msdn.microsoft.com/en-us/library/ie/hh779016.aspx
                        window.navigator.msSaveOrOpenBlob(blob, filename);
                    }
                    else {
                        var linkElement = window.document.createElement("a");
                        linkElement.href = window.URL.createObjectURL(blob);
                        linkElement.download = filename;
                        document.body.appendChild(linkElement);
                        linkElement.click();
                        document.body.removeChild(linkElement);
                    }

                } catch (ex) {
                    caught = ex;
                }
                return caught;
            }

            $http(options)
                .then(function (response) {
                    var data = response.data;
                    var status = response.status;

                    if($scope.invocation.contentType == "binary") {
                        var filename = uuidv4();

                        if($scope.selectedFunction.labels) {
                            var ext = $scope.selectedFunction.labels["com.openfaas.ui.ext"];
                            if(ext && ext.length > 0 ) {
                                filename = filename + "." + ext;
                            }
                        }

                        var caught = tryDownload(data, filename);
                        if(caught) {
                            console.log(caught);
                            $scope.invocationResponse = caught
                        } else {
                            $scope.invocationResponse = data.byteLength + " byte(s) received";
                        }

                    } else {

                        if (typeof data == 'object') {
                            $scope.invocationResponse = JSON.stringify(data, null, 2);
                        } else {
                            $scope.invocationResponse = data;
                        }
                    }

                    $scope.invocationInProgress = false;
                    $scope.invocationStatus = status;
                    var now = new Date().getTime();
                    $scope.roundTripDuration = (now - $scope.invocationStart) / 1000;
                    showPostInvokedToast("Success");

                }).catch(function(error1) {
                    $scope.invocationInProgress = false;
                    $scope.invocationResponse = error1.statusText + "\n" + error1.data;
                    $scope.invocationStatus = error1.status;
                    var now = new Date().getTime();
                    $scope.roundTripDuration = (now - $scope.invocationStart) / 1000;

                    showPostInvokedToast("Error");
                });
        };

        var refreshData = function (selectedNamespace) {
            $http.get(buildNamespaceAwareURL("/system/functions", selectedNamespace)).then(function (response) {
                const curNamespace = ($scope.functions.length > 0 && $scope.functions[0].namespace && $scope.functions[0].namespace) ? $scope.functions[0].namespace : "";
                const  newNamespace = (response.data && response.data[0] && response.data[0].namespace) ? response.data[0].namespace : "";

                if (response && response.data && (curNamespace !== newNamespace || $scope.functions.length != response.data.length)) {
                  $scope.functions = response.data;
                  if (!$scope.functions.indexOf($scope.selectedFunction )) {
                      $scope.selectedFunction = undefined;
                  }
                }
            });
        };

        $scope.showFunction = function(fn) {
            if ($scope.selectedFunction != fn) {
                $scope.selectedFunction = fn;
                $scope.invocation.request = "";
                $scope.invocationResponse = "";
                $scope.invocationStatus = "";
                $scope.invocationInProgress = false;
                if (fn.labels && fn.labels['com.openfaas.ui.ext']) {
                  $scope.invocation.contentType = "binary";
                } else {
                  $scope.invocation.contentType = "text";
                }
                $scope.invocation.roundTripDuration = "";
            }
        };

        var showDialog = function($event) {
            var parentEl = angular.element(document.body);
            $mdDialog.show({
                parent: parentEl,
                targetEvent: $event,
                templateUrl: "templates/newfunction.html",
                locals: {
                    item: $scope.functionTemplate,
                    selectedNamespace: $scope.selectedNamespace
                },
                controller: DialogController,
            });
        };

        var DialogController = function($scope, $mdDialog, item, selectedNamespace) {
            var fetchNamespaces = function () {
                $http.get("../system/namespaces")
                    .then(function(response) {
                        $scope.allNamespaces = response.data;
                    })
            }
            $scope.selectedTabIdx = newFuncTabIdx;
            $scope.item = {};
            $scope.selectedFunc = null;
            $scope.envFieldsVisible = false;
            $scope.envVarInputs = [{key: "", value: ""}];

            $scope.secretFieldsVisible = false;
            $scope.secretInputs = [{name: ""}];

            $scope.labelFieldsVisible = false;
            $scope.labelInputs = [{key: "", value: ""}];

            $scope.annotationFieldsVisible = false;
            $scope.annotationInputs = [{key: "", value: ""}];
            $scope.namespaceSelect = selectedNamespace;
            fetchNamespaces();


            $scope.closeDialog = function() {
                $mdDialog.hide();
            };

            $scope.onFuncSelected = function(func) {
                $scope.item.image = func.image;
                $scope.item.service = func.name;
                $scope.item.envProcess = func.fprocess;
                $scope.item.network = func.network;
                $scope.envVarsToEnvVarInputs(func.environment);
                $scope.labelsToLabelInputs(func.labels);
                $scope.annotationsToAnnotationInputs(func.annotations);
                $scope.secretsToSecretInputs(func.secrets);
                $scope.item.namespace = $scope.namespaceSelected();


                $scope.selectedFunc = func;
            }

            $scope.onTabSelect = function(idx) {
                newFuncTabIdx = idx;
            }

            $scope.onStoreTabDeselect = function() {
                $scope.selectedFunc = null;
            }

            $scope.onCustomTabDeselect = function() {
                $scope.item = {};
            }

            $scope.createFunc = function() {
                $scope.item.envVars = $scope.envVarInputsToEnvVars();
                $scope.item.secrets = $scope.secretInputsToSecrets();
                $scope.item.labels = $scope.labelInputsToLabels();
                $scope.item.annotations = $scope.annotationInputsToAnnotations();
                $scope.item.namespace = $scope.namespaceSelected();

                var options = {
                    url: "../system/functions",
                    data: $scope.item,
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    responseType: "text"
                };

                $http(options)
                    .then(function(response) {
                        item.image = "";
                        item.service = "";
                        item.envProcess = "";
                        item.network = "";
                        item.envVars = {};
                        item.labels = {};
                        item.annotations = {};
                        item.secrets = [];
                        item.namespace = "";

                        $scope.validationError = "";
                        $scope.closeDialog();
                        showPostInvokedToast("Function created");
                    }).catch(function(error1) {
                        showPostInvokedToast("Error");
                        $scope.selectedTabIdx = CUSTOM_DEPLOY_TAB_INDEX;
                        $scope.validationError = error1.data;
                    });
            };

            $scope.fnNamespaceSelected = function(inputNamespace) {
                $scope.namespaceSelect = inputNamespace;
            }

            $scope.onEnvInputExpand = function() {
                $scope.envFieldsVisible = !$scope.envFieldsVisible;
            }

            $scope.addEnvVar = function(index) {
                var newInput = {key: "", value: ""};
                $scope.envVarInputs.splice(index + 1, 0, newInput);
            }

            $scope.removeEnvVar = function($event, envVar) {
                var index = $scope.envVarInputs.indexOf(envVar);
                if ($event.which == 1) {
                    $scope.envVarInputs.splice(index, 1);
                }
            }

            $scope.onSecretInputExpand = function() {
                $scope.secretFieldsVisible = !$scope.secretFieldsVisible;
            }

            $scope.addSecret = function(index) {
                var newInput = {name: ""};
                $scope.secretInputs.splice(index + 1, 0, newInput);
            }

            $scope.removeSecret = function($event, secret) {
                var index = $scope.secretInputs.indexOf(secret);
                if ($event.which == 1) {
                    $scope.secretInputs.splice(index, 1);
                }
            }

            $scope.onLabelInputExpand = function() {
                $scope.labelFieldsVisible = !$scope.labelFieldsVisible;
            }

            $scope.addLabel = function(index) {
                var newInput = {key: "", value: ""};
                $scope.labelInputs.splice(index + 1, 0, newInput);
            }

            $scope.removeLabel = function($event, label) {
                var index = $scope.labelInputs.indexOf(label);
                if ($event.which == 1) {
                    $scope.labelInputs.splice(index, 1);
                }
            }

            $scope.onAnnotationInputExpand = function() {
                $scope.annotationFieldsVisible = !$scope.annotationFieldsVisible;
            }

            $scope.addAnnotation = function(index) {
                var newInput = {key: "", value: ""};
                $scope.annotationInputs.splice(index + 1, 0, newInput);
            }

            $scope.removeAnnotation = function($event, annotation) {
                var index = $scope.annotationInputs.indexOf(annotation);
                if ($event.which == 1) {
                    $scope.annotationInputs.splice(index, 1);
                }
            }

            $scope.envVarInputsToEnvVars = function() {
                var self = this;
                var result = {};
                for(var i = 0; i < self.envVarInputs.length; i++) {
                    if (self.envVarInputs[i].key && self.envVarInputs[i].value) {
                        result[self.envVarInputs[i].key] = self.envVarInputs[i].value;
                    }
                }

                return result;
            }

            $scope.envVarsToEnvVarInputs = function(envVars) {
                var result = [];
                for (var e in envVars) {
                    result.push({key: e, value: envVars[e]});
                }

                if (result.length > 0) {
                    // make the fields visible if deploying from store with values
                    $scope.envFieldsVisible = true;
                } else {
                    result.push({key: "", value: ""});
                    $scope.envFieldsVisible = false;
                }

                $scope.envVarInputs = result;
            }

            $scope.secretInputsToSecrets = function() {
                var self = this;
                var result = [];
                for(var i = 0; i < self.secretInputs.length; i++) {
                    if (self.secretInputs[i].name) {
                        result.push(self.secretInputs[i].name);
                    }
                }

                return result;
            }

            $scope.secretsToSecretInputs = function(secrets) {
                var result = [];
                for (var secret in secrets) {
                    result.push({name: secret});
                }

                if (result.length > 0) {
                    // make the fields visible if deploying from store with values
                    $scope.secretFieldsVisible = true;
                } else {
                    result.push({name: ""});
                    $scope.secretFieldsVisible = false;
                }

                $scope.secretInputs = result;
            }

            $scope.labelInputsToLabels = function() {
                var self = this;
                var result = {};
                for(var i = 0; i < self.labelInputs.length; i++) {
                    if (self.labelInputs[i].key && self.labelInputs[i].value) {
                        result[self.labelInputs[i].key] = self.labelInputs[i].value;
                    }
                }

                return result;
            }

            $scope.labelsToLabelInputs = function(labels) {
                var result = [];
                for (var l in labels) {
                    result.push({key: l, value: labels[l]});
                }

                if (result.length > 0) {
                    // make the fields visible if deploying from store with values
                    $scope.labelFieldsVisible = true;
                } else {
                    result.push({key: "", value: ""});
                    $scope.labelFieldsVisible = false;
                }

                $scope.labelInputs = result;
            }

            $scope.annotationInputsToAnnotations = function() {
                var self = this;
                var result = {};
                for(var i = 0; i < self.annotationInputs.length; i++) {
                    if (self.annotationInputs[i].key && self.annotationInputs[i].value) {
                        result[self.annotationInputs[i].key] = self.annotationInputs[i].value;
                    }
                }

                return result;
            }

            $scope.namespaceSelected = function() {
                var self = this;
                return self.namespaceSelect;
            }

            $scope.annotationsToAnnotationInputs = function(annotations) {
                var result = [];
                for (var a in annotations) {
                    result.push({key: a, value: annotations[a]});
                }

                if (result.length > 0) {
                    // make the fields visible if deploying from store with values
                    $scope.annotationFieldsVisible = true;
                } else {
                    result.push({key: "", value: ""});
                    $scope.annotationFieldsVisible = false;
                }

                $scope.annotationInputs = result;
            }
        };

        $scope.newFunction = function() {
            showDialog();
        };

        $scope.deleteFunction = function($event) {
            var confirm = $mdDialog.confirm()
                .title('Delete Function')
                .textContent('Are you sure you want to delete ' + $scope.selectedFunction.name + '?')
                .ariaLabel('Delete function')
                .targetEvent($event)
                .ok('OK')
                .cancel('Cancel');

            $mdDialog.show(confirm)
                .then(function() {
                    var options = {

                        url: buildNamespaceAwareURL("/system/functions", $scope.selectedNamespace),
                        data: {
                            functionName: $scope.selectedFunction.name
                        },
                        method: "DELETE",
                        headers: { "Content-Type": "application/json" },
                        responseType: "json"
                    };

                    return $http(options);
                }).then(function() {
                    showPostInvokedToast("Success");
                }).catch(function(err) {
                    if (err) {
                        // show error toast only if there actually is an err.
                        // because hitting 'Cancel' also rejects the promise.
                        showPostInvokedToast("Error");
                    }
                });
        };
    }
]);

function uuidv4() {
    var cryptoInstance = window.crypto || window.msCrypto; // for IE11
    return ([1e7]+-1e3+-4e3+-8e3+-1e11).replace(/[018]/g, function(c) {
        return (c ^ cryptoInstance.getRandomValues(new Uint8Array(1))[0] & 15 >> c / 4).toString(16)
    })
}

function buildNamespaceAwareURL(path, namespace) {
    let newUrl = path.startsWith("/")? ".." + path: "../" + path;

    if (namespace && namespace.length > 0){
        newUrl +=  "?namespace=" +  namespace
    }
    return newUrl
}