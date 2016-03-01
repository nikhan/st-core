var app = app || {};

/*(function() {
    app.TreeComponent = React.createClass({
        displayName: 'TreeComponent',
        _onMouseDown: function(e) {
            app.NodeStore.setRoot(this.props.tree.id);
        },
        render: function() {
            var style = '';
            if (app.NodeStore.getRoot() === this.props.tree.id) {
                style += 'current-group'
            }

            var label = this.props.tree.id;
            var node = app.NodeStore.getNode(this.props.tree.id);
            if (node.data.label.length > 0) {
                label = node.data.label;
            }

            var children = [
                React.createElement('span', {
                    onMouseDown: this._onMouseDown,
                    className: style,
                }, label)
            ]
            if (this.props.tree.children.length !== 0) {
                var list = this.props.tree.children.map(function(child) {
                    return React.createElement('li', {},
                        React.createElement(app.TreeComponent, {
                            key: child.id,
                            tree: child
                        }, null));
                })
                children.push(React.createElement('ul', {}, list));
            }
            return React.createElement('div', {
                className: 'group_tree'
            }, children);
        }
    })
})();*/

(function() {
    app.GroupListElementComponent = React.createClass({
        displayName: 'GroupListElementComponent',
        render: function() {
            return React.createElement('div', {
                onClick: this._onClick,
            }, React.createElement('a', {
                'href': '#' + this.props.group.id
            }, this.props.group.id))
        }
    });
})();

/* GroupTreeComponent
 * Sidebar widget for displaying group hierarchy, selection/moving between
 * groups.
 */
(function() {
    app.GroupListComponent = React.createClass({
        displayName: 'GroupListComponent',
        getInitialState: function() {
            return {
                groups: []
            }
        },
        componentDidMount: function() {
            app.RootGroupStore.addListener(this._update);
            this._update();
        },
        componentWillUnmount: function() {
            app.RootGroupStore.removeListener(this._update);
        },
        _update: function() {
            this.setState({
                groups: app.RootGroupStore.getGroups()
            })
        },
        render: function() {
            var children = [
                React.createElement('div', {
                    key: 'block_header',
                    className: 'block_header',
                }, 'groups'),
            ];

            children = children.concat(this.state.groups.map(function(g) {
                return React.createElement(app.GroupListElementComponent, {
                    group: g,
                    key: g.id
                }, null)
            }));

            return React.createElement('div', {
                className: 'panel unselectable'
            }, children);
        }
    })
})();
