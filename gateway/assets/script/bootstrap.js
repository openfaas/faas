"use strict"
var app = angular.module('faasGateway', ['ngMaterial']);

app.controller("home", ['$scope', '$log', '$http', '$location', '$timeout', function($scope, $log, $http, $location, $timeout) {
    $scope.functions = [];
    $scope.invocationRequest = "";
    $scope.invocationResponse = "";
    $scope.invocationStatus = "";
    $scope.invocation = {

    };
    $scope.invocation.request = ""
    setInterval(function() {
        refreshData();
    }, 1000);

    $scope.fireRequest = function() {
        $http({url:"/function/"+$scope.selectedFunction.name, data: $scope.invocation.request, method: "POST",  headers: {"Content-Type": "text/plain"}, responseType: "text"}).
        then(function(response) {
            $scope.invocationResponse = response.data;
            $scope.invocationStatus = response.status;
        }).catch(function(error1) {
            $scope.invocationResponse = error1;
            $scope.invocationStatus = null;
        });

        console.log("POST /function/"+ $scope.selectedFunction.name);
        console.log("Body: " + $scope.invocation.request);
    };

    var refreshData = function () {
        var previous = $scope.functions;

        var cl = function(previousItems) {
            $http.get("/system/functions").then(function(response) {
                if(response && response.data) {
                    if(previousItems.length !=response.data.length) {
                        $scope.functions = response.data;
                    } else {
                        for(var i =0;i<$scope.functions.length;i++) {
                            for(var j =0;j<response.data.length;j++) {
                                if($scope.functions[i].name == response.data[j].name) {
                                    $scope.functions[i].replicas=response.data[j].replicas;
                                    $scope.functions[i].invocationCount=response.data[j].invocationCount;
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
        if($scope.selectedFunction!=fn) {
            $scope.selectedFunction = fn;
            $scope.invocation.request = "";
            $scope.invocationResponse = "";
            $scope.invocationStatus = "";
        }
    };

    // TODO: popup + form to create new Docker service.
    $scope.newFunction = function() {
        $scope.functions.push({
            name: "f" +($scope.functions.length+2),
            replicas: 0,
            invokedCount: 0
        });
    };

    fetch();
}]);
