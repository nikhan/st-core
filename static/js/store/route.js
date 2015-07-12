var app = app || {};

(function() {
    var routes = {};

    function Route(data) {
        this.blocked = false;
        this.data = data;
        this.active = data.hasOwnProperty('value') && data.value != null;
    }

    Route.prototype = Object.create(app.Emitter.prototype);
    Route.constructor = Route;

    // we've received an update for the value of the route
    Route.prototype.updateData = function(data) {
        this.data.value = data;
        this.active = data !== null;
    }

    // we've received an update for the status of the route
    Route.prototype.update = function(data) {
        this.blocked = data;
    }

    function RouteStore() {}
    RouteStore.prototype = Object.create(app.Emitter.prototype);
    RouteStore.constructor = RouteStore;

    RouteStore.prototype.getRoute = function(id) {
        return routes[id];
    }

    var rs = new RouteStore();

    function createRoute(route) {
        if (routes.hasOwnProperty(route.id) === true) {
            console.warn('could not create route:', route.id, ' already exists');
            return
        }
        routes[route.id] = new Route(route.data);
    }

    function deleteRoute(id) {
        if (routes.hasOwnProperty(id) === false) {
            console.warn('could not delete route: ', id, ' does not exist');
            return
        }
        delete routes[id]
    }

    app.Dispatcher.register(function(event) {
        switch (event.action) {
            case app.Actions.APP_ROUTE_CREATE:
                createRoute(event);
                rs.emit();
                break;
            case app.Actions.APP_ROUTE_DELETE:
                deleteRoute(action.id);
                rs.emit();
                break;
            case app.Actions.APP_ROUTE_UPDATE_POSITION:
                break;
            case app.Actions.APP_ROUTE_UPDATE:
                routes[event.id].updateData(event.value);
                routes[event.id].emit();
                break;
            case app.Actions.APP_ROUTE_UPDATE_STATUS:
                routes[event.id].update(event.blocked);
                routes[event.id].emit();
                break;
            case app.Actions.APP_ROUTE_UPDATE_CONNECTED:
                break;
        }
    })

    app.RouteStore = rs;
}())