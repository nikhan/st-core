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
	library       map[string]core.Spec
}

func NewGraph() *Graph {
	return &Graph{
		elements:      make(map[ElementID]Elements),
		elementParent: make(map[ElementID]ElementID),
		Changes:       make(chan interface{}),
		index:         0,
		library:       core.GetLibrary(),
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
	spec, ok := core.GetLibrary()[b.Spec]
	if !ok {
		return nil, errors.New(fmt.Sprintf("could not create spec %s: does not exist"))
	}

	if e.Position == nil {
		b.Position = Position{
			X: 0,
			Y: 0,
		}
	} else {
		b.Position = *e.Position
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

	return []ID{}, nil
}
func (g *Graph) addGroup(e *CreateElement) error {
	return nil
}
func (g *Graph) addConnection(e *CreateElement) error {
	return nil
}
func (g *Graph) addLink(e *CreateElement) error {
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

	g.elements[*e.ID] = r
	return nil
}

func (g *Graph) Add(elements []*CreateElement, parent *ElementID) ([]ID, error) {
	oldIDs := make(map[ElementID]*ElementID)
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

	// replace IDs and add to graph.
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

func (g *Graph) Update(id ElementID, update *UpdateElement) error {
	return nil
}

func (g *Graph) UpdateGroupRoute(id ElementID, routeID ElementID, update *UpdateElement) error {
	return nil
}

func (g *Graph) BatchTranslate(ids []ElementID, xOffset int, yOffset int) error {
	return nil
}

func (g *Graph) BatchUngroup(ids []ElementID) error {
	return nil
}

func (g *Graph) BatchDelete(ids []ElementID) error {
	return nil
}

func (g *Graph) BatchReset(ids []ElementID) error {
	return nil
}
