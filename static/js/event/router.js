var app = app || {};

(function() {

    // this is a bad thing
    /*var sanitizeEvent = {
        'block_create': app.Actions.WS_BLOCK_CREATE,
        'block_update': app.Actions.WS_BLOCK_UPDATE,
        'block_delete': app.Actions.WS_BLOCK_DELETE,
        'block_info': app.Actions.WS_BLOCK_UPDATE_STATUS,
        'group_create': app.Actions.WS_GROUP_CREATE,
        'group_update': app.Actions.WS_GROUP_UPDATE,
        'group_delete': app.Actions.WS_GROUP_DELETE,
        'groupRoute_update': app.Actions.WS_GROUPROUTE_UPDATE,
        'child_create': app.Actions.WS_GROUP_ADD_CHILD,
        'child_delete': app.Actions.WS_GROUP_REMOVE_CHILD,
        'source_create': app.Actions.WS_SOURCE_CREATE,
        'source_update': app.Actions.WS_SOURCE_UPDATE,
        'param_update': app.Actions.WS_SOURCE_UPDATE_PARAMS,
        'source_delete': app.Actions.WS_SOURCE_DELETE,
        'link_create': app.Actions.WS_LINK_CREATE,
        'link_delete': app.Actions.WS_LINK_DELETE,
        'connection_create': app.Actions.WS_CONNECTION_CREATE,
        'connection_delete': app.Actions.WS_CONNECTION_DELETE,
    }

    function router(event) {
        var action = sanitizeEvent[event.type + '_' + event.action];
        switch (event.type) {
            case 'block':
                app.Dispatcher.dispatch({
                    action: action,
                    id: event.data.block.id,
                    data: event.data.block
                });
                break;
            case 'route':
                app.Dispatcher.dispatch({
                    action: app.Actions.APP_ROUTE_UPDATE,
                    id: event.data.id + '_' + event.data.route + '_input',
                    value: event.data.value
                });
                break;
            case 'param':
                app.Dispatcher.dispatch({
                    action: action,
                    id: event.data.id,
                    value: event.data
                })
                break;
            case 'connection':
                app.Dispatcher.dispatch({
                    action: action,
                    id: event.data.connection.id,
                    data: event.data.connection
                });
                break;
            case 'group':
                app.Dispatcher.dispatch({
                    action: action,
                    id: event.data.group.id,
                    data: event.data.group
                });
                break;
            case 'groupRoute':
                app.Dispatcher.dispatch({
                    action: action,
                    id: event.data.group.id,
                    data: event.data
                });
                break;
            case 'child':
                app.Dispatcher.dispatch({
                    action: action,
                    id: event.data.group.id,
                    child: event.data.child.id
                });
                break;
            case 'source':
                app.Dispatcher.dispatch({
                    action: action,
                    id: event.data.source.id,
                    data: event.data.source
                });
                break;
            case 'link':
                app.Dispatcher.dispatch({
                    action: action,
                    id: event.data.link.id,
                    data: event.data.link
                });
                break;
            default:
                console.warn('unexpected: ', event);
        }
    }*/

    var topic = null;
    var ws = new WebSocket('ws://localhost:7071/ws');

    ws.onmessage = function(m) {
        console.log(m.data);
        app.Dispatcher.dispatch(JSON.parse(m.data));
    }.bind(this)

    function unsubscribe(id) {
        ws.send(JSON.stringify({
            action: 'unsubscribe',
            id: id
        }));
    }

    function subscribe(event) {
        ws.send(JSON.stringify({
            action: 'subscribe',
            id: event.id
        }));
    }

    app.Dispatcher.register(function(event) {
        switch (event.action) {
            case 'subscribe':
                topic = event.id;
                break;
            case 'app_request_subscribe':
                if (topic !== null) {
                    unsubscribe(topic);
                }
                app.Dispatcher.dispatch({
                    action: 'app_reset_graph'
                })
                subscribe(event);
                break;
        }
    });

    //ws.onopen = function() {
    //    ws.send('{"id":"default"}');
    //}

    //app.Router = router;
})();
