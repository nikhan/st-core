var app = app || {};

// TODO: remove inputs/outputs ascending
// TODO: node emit event for route removal.

(function() {
    var graph = {};
    var root = null;

    function Element() {}
    Element.prototype = Object.create(app.Emitter.prototype);
    Element.prototype.constructor = Element;

    function requestCreate(event) {
        console.log(event);
    }

    app.Dispatcher.register(function(event) {
        switch (event.action) {
            case 'request_create':
                requestCreate(event);
                break;
            case 'create':
                break;
            case 'delete':
                break;
            case 'update_value':
                break;
            case 'update_alias':
                break;
            case 'update_position':
                break;
            case 'update_group_route_alias':
                break;
            case 'update_group_route_hidden':
                break;
        }
    })


})();
