package stserver

type ElementType string
type ElementID string

type Elements interface {
	GetID()
	GetType()
}

type ID struct {
	ID ElementID `json:"id"`
}

func (id *ID) GetID() ElementID {
	return id.ID
}

type Element struct {
	ID
	Type  ElementType `json:"type"`
	Alias string      `json:"alias"`
}

type Node struct {
	Parent *Group `json:"-"`
}

func (n *Node) GetParent() *Group {
	return n.Parent
}

func (n *Node) SetParent(g *Group) {
	n.Parent = g
}

func (e *Element) GetType() ElementType {
	return e.Type
}

type Spec struct {
	Spec string `json:"spec"`
}

type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type Group struct {
	Element
	Node
	Position `json:"position"`
	Routes   []struct {
		ID
		Hidden bool   `json:"hidden"`
		Alias  string `json:"alias"`
	} `json:"routes"`
	Children []*ID `json:"children"`
}

type Block struct {
	Element
	Spec
	Node
	Position `json:"position"`
	Routes   []ID `json:"routes"`
}

type Source struct {
	Element
	Spec
	Node
	Position `json:"position"`
	Routes   []ID `json:"routes"`
}

type Link struct {
	Element
	SourceID ElementID `json:"source_id"`
	TargetID ElementID `json:"target_id"`
}

type Connection struct {
	Element
	SourceID ElementID `json:"source_id"`
	TargetID ElementID `json:"target_id"`
}

type Route struct {
	Element
	Name      string      `json:"name"`
	Value     interface{} `json:"value"`
	Direction string      `json:"direction"`
	Source    string      `json:"source"`
}

type CreateElement struct {
	Type     *string   `json:"type"`
	Spec     *string   `json:"spec"`
	Alias    *string   `json:"alias"`
	Position *Position `json:"position"`
	Routes   []struct {
		ID
		Hidden *bool   `json:"hidden"`
		Alias  *string `json:"alias"`
	} `json:"routes"`
	Children []*ID      `json:"children"`
	SourceID *ElementID `json:"source_id"`
	TargetID *ElementID `json:"target_id"`
	Parent   *ElementID `json:"parent_id"`
}

type UpdateElement struct {
	Alias    *string     `json:"alias"`
	Position *Position   `json:"position"`
	Value    interface{} `json:"value"`
	Hidden   *bool       `json:"hidden"`
}

type BatchElement []ID
