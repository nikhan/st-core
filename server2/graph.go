package stserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sort"
	"strconv"
	"sync"

	"github.com/nytlabs/st-core/core"
)

func refString(s string) *string {
	return &s
}

func refElementID(id ElementID) *ElementID {
	return &id
}

func refInt(i int) *int {
	return &i
}

func refBool(b bool) *bool {
	return &b
}

type PubSubMessage struct {
	Topic   string
	Message interface{}
}

type PubSub struct {
	sync.Mutex
	topics  map[string]map[chan interface{}]struct{}
	publish chan *PubSubMessage
}

func NewPubSub() *PubSub {
	pb := PubSub{
		topics:  make(map[string]map[chan interface{}]struct{}),
		publish: make(chan *PubSubMessage),
	}
	go pb.listen()
	return &pb
}

func (p *PubSub) listen() {
	for {
		select {
		case m := <-p.publish:
			if subscribers, ok := p.topics[m.Topic]; ok {
				for subscriber, _ := range subscribers {
					subscriber <- m.Message
				}
			}
		}
	}
}

func (p *PubSub) Subscribe(topic string, subscription chan interface{}) {
	p.Lock()
	defer p.Unlock()
	if _, ok := p.topics[topic]; !ok {
		p.topics[topic] = make(map[chan interface{}]struct{})
	}
	p.topics[topic][subscription] = struct{}{}
}

func (p *PubSub) Unsubscribe(subscription chan interface{}) error {
	p.Lock()
	defer p.Unlock()
	for _, topic := range p.topics {
		for subscriber, _ := range topic {
			if subscriber == subscription {
				delete(topic, subscription)
				return nil
			}
		}
	}
	return errors.New("could not delete channel, does not exist")
}

func (p *PubSub) Publish(topic string, message interface{}) {
	p.publish <- &PubSubMessage{
		Topic:   topic,
		Message: message,
	}
}

type Graph struct {
	sync.Mutex
	elements      map[ElementID]*Element
	elementParent map[ElementID]ElementID
	routeToEdge   map[ElementID]map[ElementID]struct{}
	Changes       chan *PubSubMessage
	index         int64
	blockLibrary  map[string]core.Spec
	sourceLibrary map[string]core.SourceSpec
	*PubSub
}

// NewGraph returns a reference to an initialized Graph
func NewGraph() *Graph {
	return &Graph{
		elements:      make(map[ElementID]*Element),
		elementParent: make(map[ElementID]ElementID),
		routeToEdge:   make(map[ElementID]map[ElementID]struct{}),
		Changes:       make(chan *PubSubMessage),
		index:         0,
		blockLibrary:  core.GetLibrary(),
		sourceLibrary: core.GetSources(),
		PubSub:        NewPubSub(),
	}
}

// generateID generates a new unique ID for the running graph as a string.
func (g *Graph) generateID() ElementID {
	g.index += 1
	return ElementID(strconv.FormatInt(g.index, 10))
}

// addRouteFromPins accepts a slice of core.Spec pins, a direction, and adds them to the graph.
func (g *Graph) addRoutesFromPins(pins []core.Pin, direction string) ([]*ElementItem, error) {
	routes := make([]*Element, len(pins))
	elementType := ROUTE
	for i, _ := range pins {
		routes[i] = &Element{
			Type:      &elementType,
			Name:      &pins[i].Name,
			JSONType:  &pins[i].Type,
			Direction: &direction,
		}
	}
	return g.Add(routes, nil)
}

// addBlock initializes a core.Block and adds a Block to the graph.
// If addBlock receives a Block with no routes, it will automatically generate them from the corresponding core.Spec.
// If addBlock receives a Block with routes, it will assume that those routes have already been created.
func (g *Graph) addBlock(e *Element) []*ElementItem {
	var newIDs []*ElementItem

	spec := g.blockLibrary[*e.Spec]

	if e.Routes == nil {
		// no routes were sent with this block
		// that means we need to create them.
		inputs, err := g.addRoutesFromPins(spec.Inputs, INPUT)
		if err != nil {
			// if we get here something is very wrong.
			log.Fatal(err)
		}
		outputs, err := g.addRoutesFromPins(spec.Outputs, OUTPUT)
		if err != nil {
			// if we get here something is very wrong.
			log.Fatal(err)
		}
		newIDs = append(inputs, outputs...)

		// add a source route if this block needs a source
		if spec.Source != core.NONE {
			// the creation of this route is obscene
			// core.Source should really just be a string
			elementType := ROUTE
			elementDirection := INPUT
			elementJSONType := core.ANY
			name, _ := json.Marshal(spec.Source)
			names := string(name)
			elements := []*Element{&Element{
				Type:      &elementType,
				Name:      &names,
				JSONType:  &elementJSONType,
				Direction: &elementDirection,
				Source:    &spec.Source,
			}}

			sourceID, err := g.Add(elements, nil)
			if err != nil {
				log.Fatal(err)
			}
			newIDs = append(newIDs, sourceID[0])
		}

		e.Routes = newIDs
	}

	g.elements[*e.ID] = e
	return newIDs
}

// addSource initializes a core.Source and adds a Source to the graph.
// TODO: addSource always adds a route to the graph regardless of whether or not the imported Source has one.
// addSource should function similar to addBlock, and only create a new route if the imported Source does not have one.
func (g *Graph) addSource(e *Element) []*ElementItem {
	var newIDs []*ElementItem

	spec := g.sourceLibrary[*e.Spec]

	if e.Routes == nil {
		elementType := ROUTE
		elementDirection := OUTPUT
		elementJSONType := core.ANY

		elements := []*Element{&Element{
			Type:      &elementType,
			Name:      &spec.Name,
			JSONType:  &elementJSONType,
			Direction: &elementDirection,
			Source:    &spec.Type,
		}}

		var err error
		newIDs, err = g.Add(elements, nil)
		if err != nil {
			log.Fatal(err)
		}

		e.Routes = newIDs
	}

	g.elements[*e.ID] = e
	return newIDs
}

func (g *Graph) addRouteAscending(parent ElementID, route ElementID) error {
	err := g.validateIDs(parent, route)
	if err != nil {
		return err
	}

	hidden := false

	// check to see if this group already has this route added
	groupRoute, ok := g.elements[parent].GetRoute(route)
	if !ok {
		g.elements[parent].Routes = append(g.elements[parent].Routes, &ElementItem{
			ID:     refElementID(route),
			Hidden: refBool(hidden),
			Alias:  refString(""),
		})
		sort.Sort(ByID(g.elements[parent].Routes))
	} else {
		hidden = *groupRoute.Hidden
	}

	if parentID, ok := g.elementParent[parent]; ok && !hidden {
		return g.addRouteAscending(parentID, route)
	}

	return nil
}

// deleteRouteAscending deletes a route for a node and its parents.
func (g *Graph) deleteRouteAscending(parent ElementID, route ElementID) error {
	err := g.validateIDs(parent, route)
	if err != nil {
		return err
	}

	//group, ok := g.elements[parent].(*Group)
	//if !ok {
	//	return errors.New(fmt.Sprintf("deleteRouteAscending: %s not a group", parent))
	//}

	index := -1
	for i, r := range g.elements[parent].Routes {
		if *r.ID == route {
			index = i
		}
	}

	if index == -1 {
		return errors.New(fmt.Sprintf("deleteRouteAscending: %s does not have route %s", parent, route))
	}

	g.elements[parent].Routes = append(g.elements[parent].Routes[:index], g.elements[parent].Routes[index+1:]...)

	if parentID, ok := g.elementParent[parent]; ok {
		return g.deleteRouteAscending(parentID, route)
	}

	return nil
}

// addChild adds a child to a group.
// addChild automatically adds all routes from that child to that group and its parents.
// TODO: addChild should update g.elementParent
func (g *Graph) addChild(parent ElementID, child ElementID) {
	group := g.elements[parent]
	node := g.elements[child]
	group.Children = append(group.Children, &ElementItem{ID: refElementID(child)})
	sort.Sort(ByID(group.Children))

	g.elementParent[child] = parent

	for _, route := range node.Routes {
		if route.Hidden != nil && *route.Hidden {
			continue
		}
		g.addRouteAscending(parent, *route.ID)
	}
}

// deleteChild removes a child from a group.
// deleteChild automatically removes all routes from that group and its parents.
// TODO: deleteChild should update g.elementParent
func (g *Graph) deleteChild(parent ElementID, child ElementID) {
	group := g.elements[parent]
	node := g.elements[child]

	index := -1
	for i, c := range group.Children {
		if *c.ID == child {
			index = i
		}
	}

	delete(g.elementParent, child)

	group.Children = append(group.Children[:index], group.Children[index+1:]...)

	for _, route := range node.Routes {
		g.deleteRouteAscending(parent, *route.ID)
	}
}

// addGroup adds a group to the graph.
// If the added group has children, we automatically add those children to the new group.
func (g *Graph) addGroup(e *Element) {
	if e.Routes == nil {
		e.Routes = []*ElementItem{}
	}
	tmpChildren := e.Children
	e.Children = []*ElementItem{}

	g.elements[*e.ID] = e

	for _, child := range tmpChildren {
		g.addChild(*e.ID, *child.ID)
	}
}

// addConnection creates a connection and adds it to the graph.
func (g *Graph) addConnection(e *Element) {
	g.routeToEdge[*e.SourceID][*e.ID] = struct{}{}
	g.routeToEdge[*e.TargetID][*e.ID] = struct{}{}
	g.elements[*e.ID] = e
}

// addLink creates a link and adds it to the graph.
func (g *Graph) addLink(e *Element) {
	g.routeToEdge[*e.SourceID][*e.ID] = struct{}{}
	g.routeToEdge[*e.TargetID][*e.ID] = struct{}{}
	g.elements[*e.ID] = e
}

// addRoute creates a route and adds it to the graph.
func (g *Graph) addRoute(e *Element) {
	g.routeToEdge[*e.ID] = make(map[ElementID]struct{})
	g.elements[*e.ID] = e
}

func (g *Graph) validateSpec(element *Element) error {
	if element.Spec == nil {
		return errors.New("missing spec")
	}
	_, okBlock := g.blockLibrary[*element.Spec]
	_, okSource := g.sourceLibrary[*element.Spec]
	if !(okBlock || okSource) {
		return fmt.Errorf("invalid spec '%s'", *element.Spec)
	}
	return nil
}

func (g *Graph) validateRoutes(element *Element, imported map[ElementID]*Element) error {
	if element.Routes != nil {
		for _, route := range element.Routes {
			_, okImport := imported[*route.ID]
			_, okGraph := g.elements[*route.ID]
			if !(okImport || okGraph) {
				return fmt.Errorf("invalid route '%s'", *route.ID)
			}
		}
	}
	return nil
}

func (g *Graph) validateChildren(element *Element, imported map[ElementID]*Element) error {
	if element.Children != nil {
		for _, child := range element.Children {
			_, okImport := imported[*child.ID]
			if okImport {
				if !(*imported[*child.ID].Type == BLOCK ||
					*imported[*child.ID].Type == SOURCE ||
					*imported[*child.ID].Type == GROUP) {
					return fmt.Errorf("invalid child '%s' not a node (block, source, group)", *child.ID)
				}
			}
			_, okGraph := g.elements[*child.ID]
			if okGraph {
				if !g.elements[*child.ID].isNode() {
					fmt.Errorf("invalid child '%s' not a node (block, sourc, group)", *child.ID)
				}
			}

			if !(okImport || okGraph) {
				return fmt.Errorf("invalid child '%s' does not exist", *child.ID)
			}
		}
	}
	return nil
}

func (g *Graph) validateSources(element *Element, imported map[ElementID]*Element) error {
	var sourceSource *core.SourceType
	var targetSource *core.SourceType
	if _, ok := g.elements[*element.SourceID]; ok {
		sourceSource = g.elements[*element.SourceID].Source
	}
	if _, ok := g.elements[*element.TargetID]; ok {
		targetSource = g.elements[*element.TargetID].Source
	}
	if _, ok := imported[*element.SourceID]; ok {
		sourceSource = imported[*element.SourceID].Source
	}
	if _, ok := imported[*element.TargetID]; ok {
		targetSource = imported[*element.TargetID].Source
	}

	if sourceSource == nil && targetSource == nil {
		return fmt.Errorf("link has no source types")
	}

	if *sourceSource != *targetSource {
		return fmt.Errorf("source '%s' is not compatible with target '%s'", *sourceSource, *targetSource)
	}
	return nil
}

// TODO: check to make sure source links are compatible... ( same type of source )
func (g *Graph) validateEdge(element *Element, imported map[ElementID]*Element) error {
	if element.SourceID == nil {
		return errors.New("missing Source ID")
	}

	if element.TargetID == nil {
		return errors.New("missing Target ID")
	}

	_, okImported := imported[*element.SourceID]
	if okImported {
		if *imported[*element.SourceID].Type != ROUTE {
			return fmt.Errorf("invalid source node '%s'", *element.SourceID)
		}
	}

	_, okGraph := g.elements[*element.SourceID]
	if okGraph {
		if *g.elements[*element.SourceID].Type != ROUTE {
			return fmt.Errorf("invalid source node '%s'", *element.SourceID)
		}
	}

	if !(okImported || okGraph) {
		return fmt.Errorf("missing source node '%s'", *element.SourceID)
	}

	_, okImported = imported[*element.TargetID]
	if okImported {
		if *imported[*element.TargetID].Type != ROUTE {
			fmt.Errorf("invalid target node '%s'", *element.TargetID)
		}
	}
	_, okGraph = g.elements[*element.TargetID]
	if okGraph {
		if *g.elements[*element.TargetID].Type != ROUTE {
			return fmt.Errorf("invalid target node '%s'", *element.TargetID)
		}
	}

	if !(okImported || okGraph) {
		return fmt.Errorf("missing target node '%s'", *element.TargetID)
	}

	return nil
}

func (g *Graph) validateRoute(element *Element) error {
	if element.JSONType == nil {
		return errors.New("could not create route, no JSONType found")
	}

	if element.Direction == nil {
		return errors.New("could not create route, no Direction found")
	}

	if element.Name == nil {
		return errors.New("could not create route, no Name")
	}

	return nil
}

func validationError(index int, err error) error {
	return fmt.Errorf("element[%d] has error: %s", index, err.Error())
}

// Add accepts a slice of Elements and a parent and attempts to add them to the graph.
// TODO: validation should be moved into Add
func (g *Graph) Add(elements []*Element, parent *ElementID) ([]*ElementItem, error) {
	oldIDs := make(map[ElementID]*ElementID)
	children := make(map[ElementID]struct{})
	imported := make(map[ElementID]*Element)
	newIDs := []*ElementItem{}

	// if a given id doesn't exist or conflicts with present elements, make a
	// new one.
	for _, element := range elements {
		var id ElementID
		if element.ID == nil {
			id = g.generateID()
		} else {
			if _, ok := g.elements[*element.ID]; ok {
				id = g.generateID()
				oldIDs[*element.ID] = &id
			} else {
				id = *element.ID
			}
		}
		element.ID = &id
		imported[id] = element
	}

	// replace IDs
	for _, element := range elements {
		//update all routes with new IDs
		if element.Routes != nil {
			for index, route := range element.Routes {
				if _, ok := oldIDs[*route.ID]; ok {
					element.Routes[index].ID = oldIDs[*route.ID]
				}
			}
		}

		// update all children with new IDs
		if element.Children != nil {
			for index, child := range element.Children {
				if _, ok := oldIDs[*child.ID]; ok {
					element.Children[index].ID = oldIDs[*child.ID]
				}
				// append to our list of children IDs within this import
				children[*element.Children[index].ID] = struct{}{}
			}
		}

		// update all edges with new route IDs
		if element.SourceID != nil {
			if _, ok := oldIDs[*element.SourceID]; ok {
				element.SourceID = oldIDs[*element.SourceID]
			}
		}

		if element.TargetID != nil {
			if _, ok := oldIDs[*element.TargetID]; ok {
				element.TargetID = oldIDs[*element.TargetID]
			}
		}
	}

	// validate parent
	if parent != nil {
		if _, ok := g.elements[*parent]; !ok {
			return nil, fmt.Errorf("invalid parent: %s", *parent)
		}

		if *g.elements[*parent].Type != GROUP {
			return nil, fmt.Errorf("invalid parent: %s - not a group", *parent)
		}
	}

	// validate imported pattern
	for index, element := range elements {
		if element.Type == nil {
			return nil, validationError(index, errors.New("missing type"))
		}

		switch *element.Type {
		case BLOCK, SOURCE:
			err := g.validateSpec(element)
			if err != nil {
				return nil, validationError(index, err)
			}
			err = g.validateRoutes(element, imported)
			if err != nil {
				return nil, validationError(index, err)
			}
		case GROUP:
			err := g.validateRoutes(element, imported)
			if err != nil {
				return nil, validationError(index, err)
			}
			err = g.validateChildren(element, imported)
			if err != nil {
				return nil, validationError(index, err)
			}
		case ROUTE:
			err := g.validateRoute(element)
			if err != nil {
				return nil, validationError(index, err)
			}
		case LINK:
			err := g.validateEdge(element, imported)
			if err != nil {
				return nil, validationError(index, err)
			}
			err = g.validateSources(element, imported)
			if err != nil {
				return nil, validationError(index, err)
			}
		case CONNECTION:
			err := g.validateEdge(element, imported)
			if err != nil {
				return nil, validationError(index, err)
			}
		default:
			return nil, validationError(index, fmt.Errorf("unknown type %s", *element.Type))
		}
	}

	// add to graph
	for _, element := range elements {
		var ids []*ElementItem
		switch *element.Type {
		case BLOCK:
			ids = g.addBlock(element)
		case SOURCE:
			ids = g.addSource(element)
		case GROUP:
			g.addGroup(element)
		case CONNECTION:
			g.addConnection(element)
		case LINK:
			g.addLink(element)
		case ROUTE:
			g.addRoute(element)
		}

		//g.elements[*element.ID].SetID(*element.ID)
		//g.elements[*element.ID].SetType(*element.Type)

		//if element.Alias != nil {
		//	g.elements[*element.ID].SetAlias(*element.Alias)
		//}

		if g.elements[*element.ID].isNode() {
			if element.Position == nil {
				element.Position = &Position{
					X: 0,
					Y: 0,
				}
			}

			if _, ok := children[*element.ID]; parent != nil && !ok {
				g.addChild(*parent, *element.ID)
			}
		}

		newIDs = append(newIDs, &ElementItem{ID: element.ID})

		if ids != nil {
			newIDs = append(newIDs, ids...)
		}
	}

	// TO BE FIXED
	if parent != nil {
		g.Publish(string(*parent), newIDs)
	}

	return newIDs, nil
}

/*func (g *Graph) recurseGetElements2(id ElementID) map[ElementID]struct{} {
	elements := make(map[ElementID]struct{})

	switch elem := g.elements[id].(type) {
	case *Group:
		for _, child := range elem.Children {
			children := g.recurseGetElements2(child.ID)
			for id, _ := range children {
				elements[id] = struct{}{}
			}
		}
	case Nodes:
		for _, route := range elem.GetRoutes() {
			elements[route.ID] = struct{}{}
			for conn, _ := range g.routeToEdge[route.ID] {
				elements[conn.(Elements).GetID()] = struct{}{}
			}
		}
	}

	elements[id] = struct{}{}
	return elements
}*/

func (g *Graph) recurseGetElements(id ElementID) ([]*Element, map[ElementID]struct{}) {
	elements := []*Element{}
	connections := make(map[ElementID]struct{})

	if *g.elements[id].Type == GROUP {
		for _, child := range g.elements[id].Children {
			childElements, childConnections := g.recurseGetElements(*child.ID)
			elements = append(elements, childElements...)
			for id, _ := range childConnections {
				connections[id] = struct{}{}
			}
		}
	} else if g.elements[id].isNode() {
		for _, route := range g.elements[id].Routes {
			elements = append(elements, g.elements[*route.ID])
			for cid, _ := range g.routeToEdge[*route.ID] {
				connections[cid] = struct{}{}
			}
		}
	}

	elements = append(elements, g.elements[id])
	return elements, connections
}

// retrieve element and all children of element, including routes
// if edgeInclusive is true, only return edges where both source and target are
// present inside the returned set of elements. if edgeInclusive is false,
// return all connected affiliated with the element and the element's children
func (g *Graph) getElement(id ElementID, edgeInclusive bool) []*Element {
	re, rc := g.recurseGetElements(id)
	final := []*Element{}

	// basic lexical sort to ensure that we always have the same ordering
	// for an exported pattern
	connectionIDs := make([]string, len(rc))
	index := 0
	for id, _ := range rc {
		connectionIDs[index] = string(id)
		index += 1
	}
	sort.Strings(connectionIDs)

	for _, e := range re {
		final = append(final, e)
		for i := len(connectionIDs) - 1; i >= 0; i-- {
			id := ElementID(connectionIDs[i])
			source := *g.elements[id].SourceID
			target := *g.elements[id].TargetID
			sFound := false
			tFound := false
			for _, e2 := range final {
				if *e2.ID == source {
					sFound = true
				}
				if *e2.ID == target {
					tFound = true
				}
			}
			if edgeInclusive && sFound && tFound || !edgeInclusive {
				final = append(final, g.elements[id])
				connectionIDs = append(connectionIDs[:i], connectionIDs[i+1:]...)
			}
		}

	}
	return final
}

func (g *Graph) Get(ids ...ElementID) ([]*Element, error) {
	elements := []*Element{}

	if len(ids) == 0 {
		nullParents := []ElementID{}
		for k, _ := range g.elements {
			found := false
			for id, _ := range g.elementParent {
				if id == k {
					found = true
					break
				}
			}
			if g.elements[k].isNode() && !found {
				nullParents = append(nullParents, k)
			}
		}

		for _, id := range nullParents {
			elements = append(elements, g.getElement(id, true)...)
		}
	} else {
		for _, id := range ids {
			elements = append(elements, g.getElement(id, true)...)
		}
	}

	return elements, nil
}

func (g *Graph) SetState(id ElementID, state interface{}) error {
	return nil
}

func (g *Graph) GetState(id ElementID) (interface{}, error) {
	return struct{}{}, nil
}

func (g *Graph) validateIDs(ids ...ElementID) error {
	for _, id := range ids {
		if _, ok := g.elements[id]; !ok {
			return errors.New(fmt.Sprintf("error could not find %s", id))
		}
	}
	return nil
}

func (g *Graph) Update(id ElementID, update *UpdateElement) error {
	err := g.validateIDs(id)
	if err != nil {
		return err
	}

	if *g.elements[id].Type == ROUTE && update.Value != nil {
		g.elements[id].Value = update.Value
	}

	return nil
}

func (g *Graph) UpdateGroupRoute(id ElementID, routeID ElementID, update *UpdateElement) error {
	err := g.validateIDs(id, routeID)
	if err != nil {
		return err
	}

	group := g.elements[id]
	route, ok := group.GetRoute(routeID)
	if !ok {
		return errors.New(fmt.Sprintf("could not find route '%s' on group '%s'", routeID, id))
	}

	if update.Alias != nil {
		route.Alias = update.Alias
	}

	if update.Hidden != nil {
		route.Hidden = update.Hidden
		if _, ok := g.elementParent[id]; ok {
			if *route.Hidden == true {
				g.deleteRouteAscending(g.elementParent[id], routeID)
			} else {
				g.addRouteAscending(g.elementParent[id], routeID)
			}
		}
	}

	return nil
}

func (g *Graph) BatchTranslate(ids []ElementID, xOffset int, yOffset int) error {
	err := g.validateIDs(ids...)
	if err != nil {
		return err
	}

	for _, id := range ids {
		node := g.elements[id]
		position := node.Position
		position.X += xOffset
		position.Y += yOffset
	}

	return nil
}

func (g *Graph) BatchUngroup(ids []ElementID) error {
	err := g.validateIDs(ids...)
	if err != nil {
		return err
	}

	return nil
}

func (g *Graph) deleteBlock(id ElementID) {

}

func (g *Graph) deleteGroup(id ElementID) {

}

func (g *Graph) deleteSource(id ElementID) {

}

func (g *Graph) deleteLink(id ElementID) {

}

func (g *Graph) deleteConnection(id ElementID) {

}

func (g *Graph) BatchDelete(ids []ElementID) error {
	err := g.validateIDs(ids...)
	if err != nil {
		return err
	}

	deleteIDs := make(map[ElementID]struct{})

	for _, id := range ids {
		elements := g.getElement(id, false)
		for _, e := range elements {
			deleteIDs[*e.ID] = struct{}{}
		}
	}

	for id, _ := range deleteIDs {
		switch *g.elements[id].Type {
		case GROUP:
			for _, child := range g.elements[id].Children {
				delete(g.elementParent, *child.ID)
			}
			fallthrough
		case BLOCK, SOURCE:
			if p, ok := g.elementParent[id]; ok {
				g.deleteChild(p, id)
			}
		case CONNECTION, LINK:
			delete(g.routeToEdge[*g.elements[id].TargetID], id)
			delete(g.routeToEdge[*g.elements[id].SourceID], id)
		}
		delete(g.elements, id)
	}
	return nil
}

func (g *Graph) BatchReset(ids []ElementID) error {
	err := g.validateIDs(ids...)
	if err != nil {
		return err
	}

	return nil
}
