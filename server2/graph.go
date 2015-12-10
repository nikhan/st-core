package stserver

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
)

type Graph struct {
	sync.Mutex
	elements      map[ElementID]Elements
	elementParent map[ElementID]ElementID
	Changes       chan interface{}
	index         int64
}

func NewGraph() *Graph {
	return &Graph{
		elements:      make(map[ElementID]Elements),
		elementParent: make(map[ElementID]ElementID),
		Changes:       make(chan interface{}),
		index:         0,
	}
}

func (g *Graph) generateID() ElementID {
	g.index += 1
	return ElementID(strconv.FormatInt(g.index, 10))
}

func (g *Graph) addBlock(e *CreateElement) error {
	b := &Block{
		Element: Element{},
	}

	if e.Spec == nil {
		return errors.New("block has no spec!")
	}

	b.Spec = Spec(*e.Spec)

	if *e.Position == nil {
		b.Position = Position{
			x: 0,
			y: 0,
		}
	} else {
		b.Position = &e.Position
	}

	if *e.Position.Routes == nil {
		fmt.Println("wjp carws!")
	} else {
		b.Routes = &e.Routes
	}

	g.elements[*e.ID] = b

	/*if e.Position != nil {
		block.SetPosition(*e.Position)
	}

	if e.Alias != nil {
		block.SetAlias(*e.Alias)
	}*/

	return nil
}
func (g *Graph) addGroup(e *CreateElement) error {
	return nil
}
func (g *Graph) addSource(e *CreateElement) error {
	return nil
}
func (g *Graph) addConnection(e *CreateElement) error {
	return nil
}
func (g *Graph) addLink(e *CreateElement) error {
	return nil
}
func (g *Graph) addRoute(e *CreateElement) error {
	return nil
}

func (g *Graph) Add(elements []*CreateElement, parent *ElementID) error {
	oldIDs := make(map[ElementID]*ElementID)

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
				if _, ok := oldIDs[*child.ID]; ok {
					element.Children[index].ID = oldIDs[*child.ID]
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

		var err error
		switch *element.Type {
		case BLOCK:
			err = g.addBlock(element)
		case GROUP:
			err = g.addGroup(element)
		case SOURCE:
			err = g.addSource(element)
		case CONNECTION:
			err = g.addConnection(element)
		case LINK:
			err = g.addLink(element)
		case ROUTE:
			err = g.addRoute(element)
		default:
			err = errors.New(fmt.Sprintf("unable to import unknown type %s", *element.Type))
		}

		g.elements[*element.ID].Element.ID = *element.ID
		g.elements[*element.ID].Element.Type = *element.Type

		if *element.Alias != nil {
			g.elements[*element.ID].Element.Alias = *element.Alias
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func (g *Graph) Get(ids ...ElementID) ([]*Elements, error) {
	elements := []*Elements{}

	if len(ids) == 0 {
		for _, e := range g.elements {
			elements = append(elements, &e)
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
