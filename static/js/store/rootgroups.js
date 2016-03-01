(function() {
    var groups = [];

    function RootGroupStore() {}
    RootGroupStore.prototype = Object.create(app.Emitter.prototype);
    RootGroupStore.prototype.getGroups = function() {
        return groups;
    }

    var rgs = new RootGroupStore();

    function rootGroupCreate(event) {
        if (!event.hasOwnProperty('data')) return;
        for (var i = 0; i < event.data.length; i++) {
            groups.push(event.data[i]);
            //            groups[event.data[i].id] = event.data[i];
        }
    }

    function rootGroupDelete(event) {
        if (!event.hasOwnProperty('data')) return;
        for (var i = 0; i < event.data.length; i++) {
            groups = groups.filter(function(g) {
                return !(g.id === event.data[i].id);
            })
        }
    }

    app.Dispatcher.register(function(event) {
        switch (event.action) {
            case 'root_group_create':
                rootGroupCreate(event);
                rgs.emit();
                break;
            case 'root_group_delete':
                rootGroupDelete(event);
                rgs.emit();
                break;
        }
    });

    app.RootGroupStore = rgs;
})();
