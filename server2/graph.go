package stserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"sync"

	"github.com/nytlabs/st-core/core"
)

type Graph struct {
	sync.Mutex
	elements      map[ElementID]Elements
	elementParent map[ElementID]ElementID
	routeToEdge   map[ElementID]map[Elements]struct{}
	Changes       chan interface{}
	index         int64
	blockLibrary  map[string]core.Spec
	sourceLibrary map[string]core.SourceSpec
}

// NewGraph returns a reference to an initialized Graph
func NewGraph() *Graph {
	return &Graph{
		elements:      make(map[ElementID]Elements),
		elementParent: make(map[ElementID]ElementID),
		routeToEdge:   make(map[ElementID]map[Elements]struct{}),
		Changes:       make(chan interface{}),
		index:         0,
		blockLibrary:  core.GetLibrary(),
		sourceLibrary: core.GetSources(),
	}
}

// generateID generates a new unique ID for the running graph as a string.
func (g *Graph) generateID() ElementID {
	g.index += 1
	return ElementID(strconv.FormatInt(g.index, 10))
}

// addRouteFromPins accepts a slice of core.Spec pins, a direction, and adds them to the graph.
func (g *Graph) addRoutesFromPins(pins []core.Pin, direction string) ([]ID, error) {
	routes := make([]*CreateElement, len(pins))
	elementType := ROUTE
	for i, _ := range pins {
		routes[i] = &CreateElement{
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
func (g *Graph) addBlock(e *CreateElement) []ID {
	var newIDs []ID
	b := &Block{
		Element: Element{},
		Spec:    *e.Spec,
	}

	spec := g.blockLibrary[b.Spec]

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
			elements := []*CreateElement{&CreateElement{
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

		b.Routes = newIDs
	} else {
		b.Routes = make([]ID, len(e.Routes))
		for i, route := range e.Routes {
			b.Routes[i] = ID{route.ID}
		}
	}

	g.elements[*e.ID] = b
	return newIDs
}

// addSource initializes a core.Source and adds a Source to the graph.
// TODO: addSource always adds a route to the graph regardless of whether or not the imported Source has one.
// addSource should function similar to addBlock, and only create a new route if the imported Source does not have one.
func (g *Graph) addSource(e *CreateElement) []ID {
	var newIDs []ID
	s := &Source{
		Element: Element{},
		Spec:    *e.Spec,
	}

	spec := g.sourceLibrary[s.Spec]

	if e.Routes == nil {
		elementType := ROUTE
		elementDirection := OUTPUT
		elementJSONType := core.ANY

		elements := []*CreateElement{&CreateElement{
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

		s.Routes = newIDs
	} else {
		s.Routes = make([]ID, len(e.Routes))
		for i, route := range e.Routes {
			s.Routes[i] = ID{route.ID}
		}
	}

	g.elements[*e.ID] = s
	return newIDs
}

func (g *Graph) addRouteAscending(parent ElementID, route ElementID) error {
	err := g.validateIDs(parent, route)
	if err != nil {
		return err
	}

	group, ok := g.elements[parent].(*Group)
	if !ok {
		return errors.New(fmt.Sprintf("addRouteAscending: %s not a group", parent))
	}

	hidden := false

	// check to see if this group already has this route added
	groupRoute, ok := group.GetRoute(route)
	if !ok {
		group.Routes = append(group.Routes, GroupRoute{
			ID:     route,
			Hidden: hidden,
			Alias:  "",
		})
	} else {
		hidden = groupRoute.Hidden
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

	group, ok := g.elements[parent].(*Group)
	if !ok {
		return errors.New(fmt.Sprintf("deleteRouteAscending: %s not a group", parent))
	}

	index := -1
	for i, r := range group.Routes {
		if r.ID == route {
			index = i
		}
	}

	if index == -1 {
		return errors.New(fmt.Sprintf("deleteRouteAscending: %s does not have route %s", parent, route))
	}

	group.Routes = append(group.Routes[:index], group.Routes[index+1:]...)

	if parentID, ok := g.elementParent[parent]; ok {
		return g.deleteRouteAscending(parentID, route)
	}

	return nil
}

// addChild adds a child to a group.
// addChild automatically adds all routes from that child to that group and its parents.
// TODO: addChild should update g.elementParent
func (g *Graph) addChild(parent ElementID, child ElementID) {
	group := g.elements[parent].(*Group)
	node := g.elements[child].(Nodes)
	group.Children = append(group.Children, ID{child})
	g.elementParent[child] = parent

	for _, route := range node.GetRoutes() {
		g.addRouteAscending(parent, route.ID)
	}
}

// deleteChild removes a child from a group.
// deleteChild automatically removes all routes from that group and its parents.
// TODO: deleteChild should update g.elementParent
func (g *Graph) deleteChild(parent ElementID, child ElementID) {
	group := g.elements[parent].(*Group)
	node := g.elements[child].(Nodes)

	index := -1
	for i, c := range group.Children {
		if c.ID == child {
			index = i
		}
	}

	group.Children = append(group.Children[:index], group.Children[index+1:]...)

	for _, route := range node.GetRoutes() {
		g.deleteRouteAscending(parent, route.ID)
	}
}

// addGroup adds a group to the graph.
// If the added group has children, we automatically add those children to the new group.
func (g *Graph) addGroup(e *CreateElement) {
	group := &Group{
		Element:  Element{},
		Children: []ID{},
		Routes:   []GroupRoute{},
	}

	g.elements[*e.ID] = group

	if e.Children != nil {
		for _, child := range e.Children {
			g.addChild(*e.ID, child.ID)
		}
	}
}

// addConnection creates a connection and adds it to the graph.
func (g *Graph) addConnection(e *CreateElement) {
	c := &Connection{
		Element: Element{},
	}

	c.SourceID = ElementID(*e.SourceID)
	c.TargetID = ElementID(*e.TargetID)

	g.routeToEdge[*e.SourceID][c] = struct{}{}
	g.routeToEdge[*e.TargetID][c] = struct{}{}
	fmt.Println(g.routeToEdge[*e.SourceID])
	g.elements[*e.ID] = c
}

// addLink creates a link and adds it to the graph.
func (g *Graph) addLink(e *CreateElement) {
	l := &Link{
		Element: Element{},
	}

	l.SourceID = ElementID(*e.SourceID)
	l.TargetID = ElementID(*e.TargetID)

	g.routeToEdge[*e.SourceID][l] = struct{}{}
	g.routeToEdge[*e.TargetID][l] = struct{}{}

	g.elements[*e.ID] = l
}

// addRoute creates a route and adds it to the graph.
func (g *Graph) addRoute(e *CreateElement) {
	r := &Route{
		Element: Element{},
	}

	r.Direction = *e.Direction
	r.JSONType = *e.JSONType
	r.Name = *e.Name
	if e.Source != nil {
		r.Source = *e.Source
	}

	g.routeToEdge[r.ID] = make(map[Elements]struct{})
	g.elements[*e.ID] = r
}

func (g *Graph) validateSpec(element *CreateElement) error {
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

func (g *Graph) validateRoutes(element *CreateElement, imported map[ElementID]*CreateElement) error {
	if element.Routes != nil {
		for _, route := range element.Routes {
			_, okImport := imported[route.ID]
			_, okGraph := g.elements[route.ID]
			if !(okImport || okGraph) {
				return fmt.Errorf("invalid route '%s'", route.ID)
			}
		}
	}
	return nil
}

func (g *Graph) validateChildren(element *CreateElement, imported map[ElementID]*CreateElement) error {
	if element.Children != nil {
		for _, child := range element.Children {
			_, okImport := imported[child.ID]
			if okImport {
				if !(*imported[child.ID].Type == BLOCK ||
					*imported[child.ID].Type == SOURCE ||
					*imported[child.ID].Type == GROUP) {
					return fmt.Errorf("invalid child '%s' not a node (block, source, group)", child.ID)
				}
			}
			_, okGraph := g.elements[child.ID]
			if okGraph {
				if _, ok := g.elements[child.ID].(Nodes); !ok {
					fmt.Errorf("invalid child '%s' not a node (block, sourc, group)", child.ID)
				}
			}

			if !(okImport || okGraph) {
				return fmt.Errorf("invalid child '%s' does not exist", child.ID)
			}
		}
	}
	return nil
}

func (g *Graph) validateSources(element *CreateElement, imported map[ElementID]*CreateElement) error {
	var sourceSource *core.SourceType
	var targetSource *core.SourceType
	if _, ok := g.elements[*element.SourceID]; ok {
		sourceSource = &g.elements[*element.SourceID].(*Route).Source
	}
	if _, ok := g.elements[*element.TargetID]; ok {
		targetSource = &g.elements[*element.TargetID].(*Route).Source
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
		return fmt.Errorf("source '%s' is not compatible with target '%s'", sourceSource, targetSource)
	}
	return nil
}

// TODO: check to make sure source links are compatible... ( same type of source )
func (g *Graph) validateEdge(element *CreateElement, imported map[ElementID]*CreateElement) error {
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
		if _, ok := g.elements[*element.SourceID].(*Route); !ok {
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
		if _, ok := g.elements[*element.TargetID].(*Route); !ok {
			return fmt.Errorf("invalid target node '%s'", *element.TargetID)
		}
	}

	if !(okImported || okGraph) {
		return fmt.Errorf("missing target node '%s'", *element.TargetID)
	}

	return nil
}

func (g *Graph) validateRoute(element *CreateElement) error {
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

// Add accepts a slice of CreateElements and a parent and attempts to add them to the graph.
// TODO: validation should be moved into Add
func (g *Graph) Add(elements []*CreateElement, parent *ElementID) ([]ID, error) {
	oldIDs := make(map[ElementID]*ElementID)
	children := make(map[ElementID]struct{})
	imported := make(map[ElementID]*CreateElement)
	newIDs := []ID{}

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
				if _, ok := oldIDs[route.ID]; ok {
					element.Routes[index].ID = *oldIDs[route.ID]
				}
			}
		}

		// update all children with new IDs
		if element.Children != nil {
			for index, child := range element.Children {
				if _, ok := oldIDs[child.ID]; ok {
					element.Children[index].ID = *oldIDs[child.ID]
				}
				// append to our list of children IDs within this import
				children[element.Children[index].ID] = struct{}{}
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

		if _, ok := g.elements[*parent].(*Group); !ok {
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
		var ids []ID
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

		g.elements[*element.ID].SetID(*element.ID)
		g.elements[*element.ID].SetType(*element.Type)

		if element.Alias != nil {
			g.elements[*element.ID].SetAlias(*element.Alias)
		}

		if n, ok := g.elements[*element.ID].(Nodes); ok {
			if element.Position != nil {
				n.SetPosition(*element.Position)
			} else {
				n.SetPosition(Position{
					X: 0,
					Y: 0,
				})
			}

			if _, ok := children[*element.ID]; parent != nil && !ok {
				g.addChild(*parent, *element.ID)
			}
		}

		newIDs = append(newIDs, ID{*element.ID})

		if ids != nil {
			newIDs = append(newIDs, ids...)
		}
	}

	return newIDs, nil
}

func (g *Graph) recurseGetElements(id ElementID) ([]Elements, map[Elements]struct{}) {
	elements := []Elements{}
	connections := make(map[Elements]struct{})

	if group, ok := g.elements[id].(*Group); ok {
		for _, child := range group.Children {
			childElements, childConnections := g.recurseGetElements(child.ID)
			elements = append(elements, childElements...)
			for conn, _ := range childConnections {
				connections[conn] = struct{}{}
			}
		}
	} else if node, ok := g.elements[id].(Nodes); ok {
		for _, id := range node.GetRoutes() {
			elements = append(elements, g.elements[id.ID])
			fmt.Println(g.routeToEdge[id.ID], "????")
			for conn, _ := range g.routeToEdge[id.ID] {
				connections[conn] = struct{}{}
			}
		}
	}

	elements = append(elements, g.elements[id])
	return elements, connections
}

func (g *Graph) Get(ids ...ElementID) ([]Elements, error) {
	elements := []Elements{}

	if len(ids) == 0 {
		for _, e := range g.elements {
			elements = append(elements, e)
		}
	} else {
		for _, id := range ids {
			e, c := g.recurseGetElements(id)
			elements = append(elements, e...)
			fmt.Println("HELLO, ", c)
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

	if route, ok := g.elements[id].(*Route); update.Value != nil && ok {
		route.Value = update.Value
	}

	return nil
}

func (g *Graph) UpdateGroupRoute(id ElementID, routeID ElementID, update *UpdateElement) error {
	err := g.validateIDs(id, routeID)
	if err != nil {
		return err
	}

	group := g.elements[id].(*Group)
	route, ok := group.GetRoute(routeID)
	if !ok {
		return errors.New(fmt.Sprintf("could not find route '%s' on group '%s'", routeID, id))
	}

	if update.Alias != nil {
		route.Alias = *update.Alias
	}

	if update.Hidden != nil {
		route.Hidden = *update.Hidden
	}

	return nil
}

func (g *Graph) BatchTranslate(ids []ElementID, xOffset int, yOffset int) error {
	err := g.validateIDs(ids...)
	if err != nil {
		return err
	}

	for _, id := range ids {
		node := g.elements[id].(Nodes)
		position := node.GetPosition()
		position.X += xOffset
		position.Y += yOffset
		node.SetPosition(position)
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

func (g *Graph) BatchDelete(ids []ElementID) error {
	err := g.validateIDs(ids...)
	if err != nil {
		return err
	}

	for _, id := range ids {
		switch g.elements[id].(type) {
		case *Block:
			fmt.Println("deleting block")
		case *Group:
			fmt.Println("deleting group")
		case *Source:
			fmt.Println("deleting source")
		case *Connection:
			fmt.Println("deleting conection")
		case *Link:
			fmt.Println("deleting link")
		}
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
