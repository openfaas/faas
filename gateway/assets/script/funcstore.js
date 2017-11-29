var funcStoreModule = angular.module('faasGateway.funcStore', ['ngMaterial']);

funcStoreModule.service('FuncStoreService', ['$http', function ($http) {
    var self = this;
    this.fetchStore = function (url) {
        return $http.get(url)
            .then(function (resp) {
                return resp.data;
            });
    };

}]);

funcStoreModule.component('funcStore', {
    templateUrl: 'templates/funcstore.html',
    bindings: {
        selectedFunc: '<',
        onSelected: '&',
    },
    controller: ['FuncStoreService', '$mdDialog', function FuncStoreController(FuncStoreService, $mdDialog) {
        var self = this;

        this.storeUrl = 'https://raw.githubusercontent.com/openfaas/store/master/store.json';
        this.selectedFunc = null;
        this.functions = [];
        this.message = '';
        this.searchText = '';

        this.search = function (func) {
            // filter with title and description
            if (!self.searchText || (func.title.toLowerCase().indexOf(self.searchText.toLowerCase()) != -1) ||
                (func.description.toLowerCase().indexOf(self.searchText.toLowerCase()) != -1)) {
                return true;
            }
            return false;
        }

        this.select = function (func, event) {
            self.selectedFunc = func;
            self.onSelected()(func, event);
        };

        this.loadStore = function () {
            self.loading = true;
            self.functions = [];
            self.message = '';
            FuncStoreService.fetchStore(self.storeUrl)
                .then(function (data) {
                    self.loading = false;
                    self.functions = data;
                })
                .catch(function (err) {
                    console.error(err);
                    self.loading = false;
                    self.message = 'Unable to reach GitHub.com';
                });
        }

        this.showInfo = function (func, event) {
            $mdDialog.show(
                $mdDialog.alert()
                .multiple(true)
                .parent(angular.element(document.querySelector('#newfunction-dialog')))
                .clickOutsideToClose(true)
                .title(func.title)
                .textContent(func.description)
                .ariaLabel(func.title)
                .ok('OK')
                .targetEvent(event)
            );
        }

        this.loadStore();

    }]
});