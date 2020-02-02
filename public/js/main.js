"use strict";
/*
import 'angular-promise-buttons';
import ngRoute from 'angular-route';
import 'angular-http-etag';
import ls from 'local-storage'
*/
angular.module('mediaweb', ['ngRoute'])
    .controller('HomeController', function ($scope, $route, $routeParams, $location, $http) {
        $scope.$route = $route;
        $scope.$location = $location;
        $scope.$routeParams = $routeParams;
        const homeCtl = this;
    })
    .controller('ConfigController', function ($scope, $route, $routeParams, $location, $http) {
        $scope.$route = $route;
        $scope.$location = $location;
        $scope.$routeParams = $routeParams;

        var configCtl = this;
        $scope.configCtl = configCtl;
        configCtl.config = {};
        configCtl.loaded = false;
        $http.get('api/config').then(function (response) {
            configCtl.config = response.data;
            configCtl.RadarrApiKey = configCtl.config.RadarrApiKey;
        }, function (error) {
            console.log(error)
        });
    })
    .config(function ($routeProvider) {
        $routeProvider
            .when('/config', {
                templateUrl: 'templates/config.html',
                controller: 'ConfigController'
            })
            .otherwise({
                templateUrl: 'templates/main.html'
            });

        // configure html5 to get links working on jsfiddle
        //$locationProvider.html5Mode(true);
    });
/*

        homeCtl.result = {};
        homeCtl.loaded = false;
        homeCtl.query = '';

        homeCtl.submit = function() {
            if(!homeCtl.query) {
                return;
            }
            $http.get(`/api/movie/search?query=${homeCtl.query}`).then(function (response) {
                console.log(response.data);
                homeCtl.result = response.data;
            }, function (error) {
                console.error(error)
            });
        };

 */
