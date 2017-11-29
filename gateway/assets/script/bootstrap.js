"use strict"
// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

var app = angular.module('faasGateway', ['ngMaterial', 'faasGateway.funcStore']);

app.controller("home", ['$scope', '$log', '$http', '$location', '$timeout', '$mdDialog', '$mdToast', '$mdSidenav',
    function($scope, $log, $http, $location, $timeout, $mdDialog, $mdToast, $mdSidenav) {
        var newFuncTabIdx = 0;
        $scope.functions = [];
        $scope.invocationInProgress = false;
        $scope.invocationRequest = "";
        $scope.invocationResponse = "";
        $scope.invocationStatus = "";
        $scope.invocationStart = new Date().getTime();
        $scope.roundTripDuration = "";
        $scope.invocation = {
            contentType: "text"
        };

        $scope.toggleSideNav = function() {
            $mdSidenav('left').toggle();
        };

        $scope.functionTemplate = {
            image: "",
            envProcess: "",
            network: "",
            service: ""
        };

        $scope.invocation.request = "";

        setInterval(function() {
            refreshData();
        }, 1000);

        var showPostInvokedToast = function(message, duration) {
            $mdToast.show(
                $mdToast.simple()
                .textContent(message)
                .position("top right")
                .hideDelay(duration || 500)
            );
        };

        $scope.fireRequest = function() {
            var options = {
                url: "/function/" + $scope.selectedFunction.name,
                data: $scope.invocation.request,
                method: "POST",
                headers: { "Content-Type": $scope.invocation.contentType == "json" ? "application/json" : "text/plain" },
                responseType: $scope.invocation.contentType
            };
            $scope.invocationInProgress = true;
            $scope.invocationResponse = "";
            $scope.invocationStatus = null;
            $scope.roundTripDuration = "";
            $scope.invocationStart = new Date().getTime()

            $http(options)
                .then(function(response) {
                    if (typeof response.data == 'object') {
                        $scope.invocationResponse = JSON.stringify(response.data, null, 2);
                    } else {
                        $scope.invocationResponse = response.data;
                    }
                    $scope.invocationInProgress = false;
                    $scope.invocationStatus = response.status;
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

        var refreshData = function() {
            var previous = $scope.functions;

            var cl = function(previousItems) {
                $http.get("/system/functions").then(function(response) {
                    if (response && response.data) {
                        if (previousItems.length != response.data.length) {
                            $scope.functions = response.data;
                        } else {
                            for (var i = 0; i < $scope.functions.length; i++) {
                                for (var j = 0; j < response.data.length; j++) {
                                    if ($scope.functions[i].name == response.data[j].name) {
                                        $scope.functions[i].replicas = response.data[j].replicas;
                                        $scope.functions[i].invocationCount = response.data[j].invocationCount;
                                    }
                                }
                            }
                        }
                    }
                });
            };
            cl(previous);
        }

        var fetch = function() {
            $http.get("/system/functions").then(function(response) {
                $scope.functions = response.data;
            });
        };

        $scope.showFunction = function(fn) {
            if ($scope.selectedFunction != fn) {
                $scope.selectedFunction = fn;
                $scope.invocation.request = "";
                $scope.invocationResponse = "";
                $scope.invocationStatus = "";
                $scope.invocationInProgress = false;
                $scope.invocation.contentType = "text";
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
                    item: $scope.functionTemplate
                },
                controller: DialogController
            });
        };

        var DialogController = function($scope, $mdDialog, item) {
            $scope.selectedTabIdx = newFuncTabIdx;
            $scope.item = item;
            $scope.selectedFunc = null;
            $scope.closeDialog = function() {
                $mdDialog.hide();
            };

            $scope.onFuncSelected = function(func) {
                $scope.item.image = func.image;
                $scope.item.service = func.name;
                $scope.item.envProcess = func.fprocess;
                $scope.item.network = func.network;
                $scope.selectedFunc = func;
            }
            
            $scope.onTabSelect = function(idx) {
                newFuncTabIdx = idx;
            }

            $scope.onStoreTabDeselect = function() {
                $scope.selectedFunc = null;
            }

            $scope.onManualTabDeselect = function() {
                $scope.item = {};
            }

            $scope.createFunc = function() {
                var options = {
                    url: "/system/functions",
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
                        $scope.validationError = "";
                        $scope.closeDialog();
                        showPostInvokedToast("Function created");
                    }).catch(function(error1) {
                        $scope.validationError = error1.data;
                    });
            };
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
                        url: "/system/functions",
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

        fetch();
    }
]);
