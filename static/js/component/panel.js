var app = app || {};

/* PanelComponent & PanelEditableComponent
 * Produces a list of fields that are the current representation of input
 * values for blocks/groups that are sent to the component.
 *
 * TODO: fix the {'data': ...} nonsense
 */

(function() {
    app.RoutePanelInput = React.createClass({
        getInitialState: function() {
            return {
                name: '',
                type: '',
                value: '',
            }
        },
        componentDidMount: function() {
            app.RouteStore.getRoute(this.props.id).addListener(this._update);
            this._update();
        },
        componentWillUnmount: function() {
            app.RouteStore.getRoute(this.props.id).removeListener(this._update);
        },
        _update: function() {
            var route = app.RouteStore.getRoute(this.props.id);
            var value = '';
            if (route.data.value !== null) {
                value = JSON.stringify(route.data.value.data);
            }

            this.setState({
                name: route.data.name,
                type: route.data.type,
                value: value,
            })
        },
        _handleChange: function(event) {
            this.setState({
                value: event.target.value
            });
        },
        _onKeyDown: function(event) {
            if (event.keyCode !== 13) return;

            var value = null;
            if (this.state.value !== null) {
                try {
                    value = {
                        data: JSON.parse(this.state.value)
                    }
                } catch (e) {
                    console.log(e);
                }
            }

            app.Dispatcher.dispatch({
                action: app.Actions.APP_REQUEST_ROUTE_UPDATE,
                id: this.props.id,
                value: value
            })
        },
        render: function() {
            return React.createElement('div', {}, [
                React.createElement('div', {
                    className: 'label',
                    key: 'label',
                }, this.state.name),
                React.createElement('input', {
                    type: 'text',
                    ref: 'value',
                    key: 'value',
                    value: this.state.value,
                    onChange: this._handleChange,
                    onKeyDown: this._onKeyDown,
                }, null)
            ]);
        }
    });
})();

(function() {
    app.GroupNameComponent = React.createClass({
        displayName: 'GroupNameComponent',
        getInitialState: function() {
            return {
                value: JSON.stringify(app.NodeStore.getNode(this.props.id).data.label)
            }
        },
        componentDidMount: function() {
            app.NodeStore.getNode(this.props.id).addListener(this._update);
        },
        componentWillUnmount: function() {
            app.NodeStore.getNode(this.props.id).removeListener(this._update);
        },
        _update: function() {
            this.setState({
                value: JSON.stringify(app.NodeStore.getNode(this.props.id).data.label)
            })
        },
        _handleChange: function(event) {
            this.setState({
                value: event.target.value
            })
        },
        _onKeyDown: function() {
            if (event.keyCode !== 13) return;

            var value = null;
            if (this.state.value !== null) {
                try {
                    value = JSON.parse(this.state.value)
                } catch (e) {
                    console.log(e);
                }
            }

            app.Dispatcher.dispatch({
                action: app.Actions.APP_REQUEST_NODE_LABEL,
                id: this.props.id,
                label: value
            })
        },
        render: function() {
            return React.createElement('div', {}, [
                React.createElement('div', {
                    className: 'label',
                    key: 'label',
                }, 'label'),
                React.createElement('input', {
                    type: 'text',
                    ref: 'value',
                    key: 'value',
                    value: this.state.value,
                    onChange: this._handleChange,
                    onKeyDown: this._onKeyDown,
                }, null)
            ]);

        }
    })
})();


(function() {
    app.RoutesPanelComponent = React.createClass({
        displayName: 'PanelComponent',
        componentDidMount: function() {
            app.NodeStore.getNode(this.props.id).addListener(this._update);
            this._update();
        },
        componentWillUnmount: function() {
            app.NodeStore.getNode(this.props.id).removeListener(this._update);
        },
        _update: function() {
            this.render();
            /*var route = app.RouteStore.getRoute(this.props.id);
            var value = '';
            if (route.data.value !== null) {
                value = JSON.stringify(route.data.value.data);
            }

            this.setState({
                name: route.data.name,
                type: route.data.type,
                value: value,
            })*/
        },
        render: function() {
            var block = app.NodeStore.getNode(this.props.id);

            var children = [
                React.createElement('div', {
                    key: 'block_header',
                    className: 'block_header',
                }, block.data.type),
                React.createElement(app.GroupNameComponent, {
                    id: this.props.id
                }, null)
            ];


            // TODO: optimize this!
            // this retrieves _all_ routes for a block, seems unnecesary
            children = children.concat(block.routes.filter(function(id) {
                return app.RouteStore.getRoute(id).direction === 'input';
            }).map(function(id) {
                return React.createElement(app.RoutePanelInput, {
                    id: id,
                    key: id,
                }, null)
            }));

            return React.createElement('div', {
                className: 'panel'
            }, children);
        }
    })
})();
