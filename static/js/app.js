var app = app || {};

// TODO 
// This file desperately needs to be refactored. The portion of CoreApp that 
// is related to the stage, the background lines, and the selection of nodes
// can be put into its own component. 

(function() {
    'use strict';

    app.CoreApp = React.createClass({
        displayName: "CoreApp",
        getInitialState: function() {
            return {
                width: null,
                height: null,
                keys: {
                    shift: false,
                },
                selected: [],
                connecting: null,
                library: {
                    x: null,
                    y: null,
                    enabled: false,
                },
            }
        },
        handleSelectionChange: function(rectX, rectY, width, height) {
            var selected = [];

            // check to see which nodes are currently in the selection box
            selected = this.props.model.focusedNodes.filter(function(node) {
                if (!node.data.hasOwnProperty('position')) return false; // we may be able to get rid of this now.
                var position = node.data.position;
                return app.Utils.pointInRect(rectX, rectY, width, height,
                    position.x + this.props.model.focusedGroup.translateX,
                    position.y + this.props.model.focusedGroup.translateY);
            }.bind(this));

            // check to see which edges are in selection box
            selected = selected.concat(this.props.model.focusedEdges.filter(function(node) {
                if (!node.hasOwnProperty('from')) return false; // we may be able to get rid of this now.
                var p1 = node.from;
                var p2 = node.to;
                return (app.Utils.pointInRect(rectX, rectY, width, height,
                        p1.x + this.props.model.focusedGroup.translateX,
                        p1.y + this.props.model.focusedGroup.translateY) &&
                    app.Utils.pointInRect(rectX, rectY, width, height,
                        p2.x + this.props.model.focusedGroup.translateX,
                        p2.y + this.props.model.focusedGroup.translateY));
            }.bind(this)));

            // update the state of the selection box
            this.setState({
                selected: selected,
            })

        },
        documentKeyDown: function(e) {
            if (e.shiftKey === true) this.setState({
                keys: {
                    shift: true
                }
            })
        },
        documentKeyUp: function(e) {
            if (e.shiftKey === false) this.setState({
                keys: {
                    shift: false
                }
            })
        },
        createBlock: function(b) {
            app.Utils.request(
                'POST',
                'blocks', {
                    'type': b,
                    'parent': this.props.model.focusedGroup.data.id,
                    'position': {
                        'x': this.state.library.x - this.props.model.focusedGroup.translateX,
                        'y': this.state.library.y - this.props.model.focusedGroup.translateY
                    }
                },
                function(e) {
                    if (e.status !== 200) {
                        // error
                    } else {
                        this.setState({
                            library: {
                                enabled: false
                            }
                        })
                    }
                }.bind(this))

        },
        createConnection: function(from, to) {
            app.Utils.request(
                'POST',
                'connections', {
                    'from': {
                        'id': from.direction === 'output' ? from.id : to.id,
                        'route': from.direction === 'output' ? from.route : to.route,
                    },
                    'to': {
                        'id': to.direction === 'input' ? to.id : from.id,
                        'route': to.direction === 'input' ? to.route : from.route,
                    }
                },
                function(e) {
                    this.setState({
                        connecting: null
                    })
                }.bind(this)
            )
        },
        componentWillMount: function() {
            window.addEventListener('keydown', this.documentKeyDown);
            window.addEventListener('keyup', this.documentKeyUp);
            window.addEventListener('resize', this.handleResize);

            this.handleResize();
        },
        handleResize: function() {
            this.setState({
                width: document.body.clientWidth,
                height: document.body.clientHeight
            })
        },
        handleMouseUp: function(e) {
            this.setState({
                library: {
                    enabled: false,
                }
            })
        },
        nodeSelect: function(id) {
            var node = this.props.model.entities[id];
            if (this.state.keys.shift === true) {
                if (this.state.selected.indexOf(node) === -1) {
                    this.setState({
                        selected: this.state.selected.concat([node])
                    })
                } else {
                    this.setState({
                        selected: this.state.selected.slice().filter(function(i) {
                            return i != node;
                        })
                    })
                }
            } else {
                this.setState({
                    selected: [node],
                })
            }
        },
        handleDoubleClick: function(e) {
            this.setState({
                library: {
                    enabled: true,
                    x: e.pageX,
                    y: e.pageY
                }
            })
        },
        handleRouteEvent: function(r) {
            if (this.state.connecting === null) {
                this.setState({
                    connecting: r
                })
            }
            if (this.state.connecting !== null) {
                this.createConnection(r, this.state.connecting)
            }

        },
        render: function() {
            var nodes = {
                'source': app.SourceComponent,
                'group': app.GroupComponent,
                'block': app.BlockComponent
            }

            var edges = {
                'link': app.LinkComponent,
                'connection': app.ConnectionComponent
            }

            var nodeElements = this.props.model.focusedNodes.map(function(c) {
                return React.createElement(app.DragContainer, {
                    key: c.data.id,
                    model: c,
                    x: c.data.position.x,
                    y: c.data.position.y,
                    nodeSelect: this.nodeSelect
                }, React.createElement(nodes[c.instance()], {
                    key: c.data.id,
                    model: c,
                    onRouteEvent: this.handleRouteEvent,
                    selected: this.state.selected.indexOf(c) !== -1 ? true : false,
                }, null))
            }.bind(this));

            var edgeElements = this.props.model.focusedEdges.map(function(c) {
                return React.createElement(edges[c.instance()], {
                    key: c.data.id,
                    model: c,
                    nodeSelect: this.nodeSelect,
                    selected: this.state.selected.indexOf(c) !== -1 ? true : false,
                }, null)
            }.bind(this));

            var renderGroups = null;
            if (this.props.model.focusedGroup !== null) {
                renderGroups = React.createElement('g', {
                    transform: 'translate(' +
                        this.props.model.focusedGroup.translateX + ', ' +
                        this.props.model.focusedGroup.translateY + ')',
                    key: 'renderGroups'
                }, edgeElements.concat(nodeElements));
            }

            var background = [];


            if (this.props.model.focusedGroup !== null) {
                var groupList = React.createElement(app.GroupSelectorComponent, {
                    focusedGroup: this.props.model.focusedGroup.data.id,
                    groups: this.props.model.groups,
                    key: "group_list",
                }, null);

                if (this.state.connecting != null) {
                    background.push(React.createElement(app.ConnectionToolComponent, {
                        key: 'tool',
                        connecting: this.state.connecting,
                        node: this.props.model.entities[this.state.connecting.id],
                        translateX: this.props.model.focusedGroup.translateX,
                        translateY: this.props.model.focusedGroup.translateY
                    }, null))
                }
            }

            if (this.props.model.focusedGroup !== null) {
                background.push(React.createElement(app.StageComponent, {
                    key: 'bg',
                    group: this.props.model.focusedGroup,
                    onSelectionChange: this.handleSelectionChange,
                    onDoubleClick: this.handleDoubleClick,
                    onMouseUp: this.handleMouseUp,
                    width: this.state.width,
                    height: this.state.height,
                }, null));
            }

            background.push(renderGroups);

            var stage = React.createElement("svg", {
                className: "stage",
                key: "stage",
                width: this.state.width,
                height: this.state.height,
                onContextMenu: function(e) {
                    e.nativeEvent.preventDefault();
                }
            }, background)

            var panelList = React.createElement(app.PanelListComponent, {
                nodes: this.state.selected,
                key: 'panel_list',
            });

            var children = [stage, groupList, panelList];

            if (this.state.library.enabled === true) {
                children.push(React.createElement(app.AutoCompleteComponent, {
                    key: 'autocomplete',
                    x: this.state.library.x,
                    y: this.state.library.y,
                    options: this.props.model.blockLibrary,
                    onChange: this.createBlock,
                }, null));
            }


            var container = React.createElement("div", {
                className: "app",
            }, children);

            return container
        }
    })
})();
