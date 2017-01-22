"use strict"
var app = angular.module('faasGateway', ['ngMaterial']);
app.controller("home", ['$scope', '$log', '$http', '$location', '$timeout', function($scope, $log, $http, $location, $timeout) {
    $scope.functions = [];
    $http.get("/system/functions").then(function(response) {
      $scope.functions = response.data;
    });

    $scope.showFunction = function(fn) {
        $scope.selectedFunction = fn;
    };

    // TODO: popup + form to create new Docker service.
    $scope.newFunction = function() {
        $scope.functions.push({
            name: "f" +($scope.functions.length+2),
            replicas: 0,
            invokedCount: 0
        });
    };
}]);
