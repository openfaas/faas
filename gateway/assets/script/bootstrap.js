"use strict"
// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

var app = angular.module('faasGateway', ['ngMaterial']);

app.controller("home", ['$scope', '$log', '$http', '$location', '$timeout', '$mdDialog', '$mdToast',
        function($scope, $log, $http, $location, $timeout, $mdDialog, $mdToast) {
    $scope.functions = [];
    $scope.invocationRequest = "";
    $scope.invocationResponse = "";
    $scope.invocationStatus = "";
    $scope.invocation = {
        contentType: "text"
    };

    $scope.functionTemplate = {
        image: "",
        envProcess: "",
        network: "",
        service: ""
    };

    $scope.invocation.request = ""

    setInterval(function() {
        refreshData();
    }, 1000);

    $scope.showPostInvokedToast = function(val) {
        $mdToast.show(
            $mdToast.simple()
                .textContent(val)
                .position("top right")
                .hideDelay(500)
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
        $scope.invocationResponse = "";
        $scope.invocationStatus = null;

        $http(options)
            .then(function(response) {
                if($scope.invocation && $scope.invocation.contentType == "json") {
                    $scope.invocationResponse = JSON.stringify(response.data, -1, "  ");
        
                } else {
                    $scope.invocationResponse = response.data;    
                }
                $scope.invocationStatus = response.status;
                $scope.showPostInvokedToast("Success");
            }).catch(function(error1) {
                $scope.invocationResponse = error1.statusText + "\n" + error1.data;
                $scope.invocationStatus = error1.status;

                $scope.showPostInvokedToast("Error");
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
            $scope.invocation.contentType = "text";
        }
    };

    var showDialog=function($event) {
       var parentEl = angular.element(document.body);
       $mdDialog.show({
         parent: parentEl,
         targetEvent: $event,
         templateUrl: "newfunction.html",
         locals: {
           item: $scope.functionTemplate
         },
         controller: DialogController
      });
    };

    var DialogController = function($scope, $mdDialog, item) {
        $scope.item = item;
        $scope.closeDialog = function() {
            $mdDialog.hide();
        };

        $scope.createFunc = function() {
            var options = {
                url: "/system/functions",
                data: $scope.item,
                method: "POST",
                headers: { "Content-Type": "application/json"},
                responseType: "json"
            };

            $http(options)
                .then(function(response) {
                    $scope.invocationResponse = response.data;
                    $scope.invocationStatus = response.status;
                }).catch(function(error1) {
                    $scope.invocationResponse = error1;
                    $scope.invocationStatus = null;
                });
            console.log($scope.item);
            $scope.closeDialog();
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
                    headers: { "Content-Type": "application/json"},
                    responseType: "json"
                };

                return $http(options);
            }).then(function(){
                $scope.showPostInvokedToast("Success");
            }).catch(function(err) {
                if (err) {
                    // show error toast only if there actually is an err.
                    // because hitting 'Cancel' also rejects the promise.
                  $scope.showPostInvokedToast("Error");
                }
            });
    };

    fetch();
}]);
