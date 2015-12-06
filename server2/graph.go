package stserver

import "sync"

type Graph struct {
	sync.Mutex
	elements map[string]Elements
	Changes  chan interface{}
}

func (g *Graph) Add(e []CreateElement, parent *ElementID) error {
	var err error

	//if e.Type == nil {
	//	return errors.New("cannot create element: no type")
	//}

	/*switch *e.Type {
	case BLOCK:
		_ = &Block{
			Spec:     e.Spec,
			Alias:    e.Alias,
			Position: e.Position,
		}
	case SOURCE:
		_ = &Source{
			Spec:     e.Spec,
			Alias:    e.Alias,
			Position: e.Position,
		}
	case GROUP:
		_ = &Group{
			Alias:    e.Alias,
			Position: e.Position,
			Children: e.Children,
			Routes:   e.Routes,
		}
	case CONNECTION:
		_ = &Connection{
			Alias:    e.Alias,
			SourceID: e.SourceID,
			TargetID: e.TargetID,
		}
	case LINK:
		_ = &Link{
			Alias:    e.Alias,
			SourceID: e.SourceID,
			TargetID: e.TargetID,
		}
	default:

	}*/

	return err
}

func (g *Graph) Get(ids ...ElementID) (*Element, error) {
	return nil, nil
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
