package stserver

import (
	"errors"
	"fmt"
	"strconv"
	"sync"

	"github.com/nytlabs/st-core/core"
)

type Graph struct {
	sync.Mutex
	elements      map[ElementID]Elements
	elementParent map[ElementID]ElementID
	Changes       chan interface{}
	index         int64
	blockLibrary  map[string]core.Spec
	sourceLibrary map[string]core.SourceSpec
}

func NewGraph() *Graph {
	return &Graph{
		elements:      make(map[ElementID]Elements),
		elementParent: make(map[ElementID]ElementID),
		Changes:       make(chan interface{}),
		index:         0,
		blockLibrary:  core.GetLibrary(),
		sourceLibrary: core.GetSources(),
	}
}

func (g *Graph) generateID() ElementID {
	g.index += 1
	return ElementID(strconv.FormatInt(g.index, 10))
}

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

func (g *Graph) addBlock(e *CreateElement) ([]ID, error) {
	var newIDs []ID
	b := &Block{
		Element: Element{},
	}

	if e.Spec == nil {
		return nil, errors.New("block has no spec!")
	}

	b.Spec = *e.Spec
	spec, ok := g.blockLibrary[b.Spec]
	if !ok {
		return nil, errors.New(fmt.Sprintf("could not create spec %s: does not exist"))
	}

	if e.Routes == nil {
		// no routes were sent with this block
		// that means we need to create them.
		inputs, err := g.addRoutesFromPins(spec.Inputs, INPUT)
		if err != nil {
			return nil, err
		}
		outputs, err := g.addRoutesFromPins(spec.Outputs, OUTPUT)
		if err != nil {
			return nil, err
		}
		newIDs = append(inputs, outputs...)
		b.Routes = newIDs
	} else {
		b.Routes = make([]ID, len(e.Routes))
		for i, route := range e.Routes {
			b.Routes[i] = ID{*route.ID}
		}
	}

	g.elements[*e.ID] = b
	return newIDs, nil
}

func (g *Graph) addSource(e *CreateElement) ([]ID, error) {
	s := &Source{
		Element: Element{},
	}

	if e.Spec == nil {
		return nil, errors.New("source has no spec!")
	}

	s.Spec = *e.Spec
	spec, ok := g.sourceLibrary[s.Spec]
	if !ok {
		return nil, errors.New(fmt.Sprintf("could not create spec %s: does not exist"))
	}

	elementType := ROUTE
	elementDirection := OUTPUT
	elementJSONType := core.ANY
	elements := []*CreateElement{&CreateElement{
		Type:      &elementType,
		Name:      &spec.Name,
		JSONType:  &elementJSONType,
		Direction: &elementDirection,
		Source:    &spec.Name,
	}}

	newIDs, err := g.Add(elements, nil)
	if err != nil {
		return nil, err
	}

	s.Routes = newIDs
	g.elements[*e.ID] = s
	return newIDs, nil
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

func (g *Graph) addChild(parent ElementID, child ElementID) error {
	err := g.validateIDs(parent, child)
	if err != nil {
		return err
	}

	group, ok := g.elements[parent].(*Group)
	if !ok {
		return errors.New(fmt.Sprintf("%s not a group", parent))
	}

	node, ok := g.elements[child].(Nodes)
	if !ok {
		return errors.New(fmt.Sprintf("could not add child %s, not a node", child))
	}

	group.Children = append(group.Children, ID{child})
	g.elementParent[child] = parent

	for _, route := range node.GetRoutes() {
		g.addRouteAscending(parent, route.ID)
	}
	return nil
}

func (g *Graph) deleteChild(parent ElementID, child ElementID) error {
	err := g.validateIDs(parent, child)
	if err != nil {
		return err
	}

	group, ok := g.elements[parent].(*Group)
	if !ok {
		return errors.New(fmt.Sprintf("%s not a group", parent))
	}

	node, ok := g.elements[child].(Nodes)
	if !ok {
		return errors.New(fmt.Sprintf("could not add child %s, not a node", child))
	}

	index := -1
	for i, c := range group.Children {
		if c.ID == child {
			index = i
		}
	}

	if index == -1 {
		return errors.New(fmt.Sprintf("group %s does not have child %s", parent, child))
	}

	group.Children = append(group.Children[:index], group.Children[index+1:]...)

	for _, route := range node.GetRoutes() {
		g.deleteRouteAscending(parent, route.ID)
	}

	return nil
}

func (g *Graph) addGroup(e *CreateElement) error {
	group := &Group{
		Element:  Element{},
		Children: []ID{},
		Routes:   []GroupRoute{},
	}

	g.elements[*e.ID] = group

	if e.Children != nil {
		for _, child := range e.Children {
			err := g.addChild(*e.ID, child.ID)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (g *Graph) validateEdge(e *CreateElement) error {
	if e.SourceID == nil || e.TargetID == nil {
		return errors.New("connection missing source or target")
	}

	return g.validateIDs(ElementID(*e.SourceID), ElementID(*e.TargetID))
}

func (g *Graph) addConnection(e *CreateElement) error {
	err := g.validateEdge(e)
	if err != nil {
		return err
	}

	c := &Connection{
		Element: Element{},
	}

	c.SourceID = ElementID(*e.SourceID)
	c.TargetID = ElementID(*e.TargetID)

	g.elements[*e.ID] = c
	return nil
}

func (g *Graph) addLink(e *CreateElement) error {
	err := g.validateEdge(e)
	if err != nil {
		return err
	}

	l := &Link{
		Element: Element{},
	}

	l.SourceID = ElementID(*e.SourceID)
	l.TargetID = ElementID(*e.TargetID)

	g.elements[*e.ID] = l
	return nil
}

func (g *Graph) addRoute(e *CreateElement) error {
	r := &Route{
		Element: Element{},
	}

	if e.JSONType == nil {
		return errors.New("could not create route, no JSONType found")
	}

	if e.Direction == nil {
		return errors.New("could not create route, no Direction found")
	}

	if e.Name == nil {
		return errors.New("could not create route, no Name")
	}

	r.Direction = *e.Direction
	r.JSONType = *e.JSONType
	r.Name = *e.Name
	if e.Source != nil {
		r.Source = *e.Source
	}

	g.elements[*e.ID] = r
	return nil
}

func (g *Graph) Add(elements []*CreateElement, parent *ElementID) ([]ID, error) {
	oldIDs := make(map[ElementID]*ElementID)
	children := make(map[ElementID]struct{})
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
				if _, ok := oldIDs[child.ID]; ok {
					element.Children[index].ID = *oldIDs[child.ID]
				}
				// append to our list of children IDs within this import
				children[*oldIDs[child.ID]] = struct{}{}
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

	// add to graph
	for _, element := range elements {
		if element.Type == nil {
			return nil, errors.New("unable to import: element has no type")
		}

		var err error
		var ids []ID
		switch *element.Type {
		case BLOCK:
			ids, err = g.addBlock(element)
		case SOURCE:
			ids, err = g.addSource(element)
		case GROUP:
			err = g.addGroup(element)
		case CONNECTION:
			err = g.addConnection(element)
		case LINK:
			err = g.addLink(element)
		case ROUTE:
			err = g.addRoute(element)
		default:
			err = errors.New(fmt.Sprintf("unable to import unknown type %s", *element.Type))
		}

		if err != nil {
			return nil, err
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
			// if:
			// element is a node,
			// element's ID is not in the children set for this import,
			// we are given a parent ID
			// then: make this node a child of given parent
			if _, ok := children[*element.ID]; parent != nil && !ok {
				err := g.addChild(*parent, *element.ID)
				if err != nil {
					return nil, err
				}
			}
		}

		newIDs = append(newIDs, ID{*element.ID})

		if ids != nil {
			newIDs = append(newIDs, ids...)
		}
	}

	return newIDs, nil
}

func (g *Graph) Get(ids ...ElementID) ([]Elements, error) {
	elements := []Elements{}

	if len(ids) == 0 {
		for _, e := range g.elements {
			elements = append(elements, e)
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
		return errors.New(fmt.Sprintf("could not find route %s on group %s", id, routeID))
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
