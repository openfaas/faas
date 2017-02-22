"use strict"
var app = angular.module('faasGateway', ['ngMaterial']);

app.controller("home", ['$scope', '$log', '$http', '$location', '$timeout', '$mdDialog', 
        function($scope, $log, $http, $location, $timeout, $mdDialog) {
    $scope.functions = [];
    $scope.invocationRequest = "";
    $scope.invocationResponse = "";
    $scope.invocationStatus = "";
    $scope.invocation = {
        contentType: "text"
    };
    $scope.functionTemplate = {
        image: ""
    };
    $scope.invocation.request = ""
    setInterval(function() {
        refreshData();
    }, 1000);

    $scope.fireRequest = function() {

        var options = {
            url: "/function/" + $scope.selectedFunction.name,
            data: $scope.invocation.request,
            method: "POST",
            headers: { "Content-Type": $scope.invocation.contentType == "json" ? "application/json" : "text/plain" },
            responseType: $scope.invocation.contentType
        };

        $http(options)
            .then(function(response) {
                $scope.invocationResponse = response.data;
                $scope.invocationStatus = response.status;
            }).catch(function(error1) {
                $scope.invocationResponse = error1;
                $scope.invocationStatus = null;
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
            $scope.invocation.contentType = "";
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
            console.log($scope.item);
            $scope.closeDialog();
        };
    };

    $scope.newFunction = function() {
        showDialog();
    };

    fetch();
}]);


// '<md-dialog aria-label="List dialog">' +
//            '  <md-dialog-content class="md-padding">'+
//            '<label>Define a function</label>'+
//            '<form name="userForm">'+
//             '<md-input-container>'+
//                 '<label>Image:</label>'+
//                 '<input name="dockerImage" ng-model="item.image" required md-maxlength="200" minlength="4">'+
//            '</md-input-container>'+
//             '<md-input-container>'+
//                 '<label>Name:</label>'+
//                 '<input name="serviceName" ng-model="item.name" required md-maxlength="200" minlength="4">'+
//            '</md-input-container>'+
//            '</form>'+
//            '  </md-dialog-content>' +
//            '  <md-dialog-actions>' +
//            '    <md-button ng-click="closeDialog()" class="md-secondary">' +
//            '      Close Dialog' +
//            '    </md-button>' +
//            '    <md-button ng-click="createFunc()" class="md-primary">' +
//            '      Create' +
//            '    </md-button>' +
//            '  </md-dialog-actions>' +
//            '</md-dialog>'