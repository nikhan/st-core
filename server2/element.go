package stserver

const (
	BLOCK      = "block"
	GROUP      = "group"
	SOURCE     = "source"
	CONNECTION = "connection"
	LINK       = "link"
	ROUTE      = "route"
)

type ElementType string
type ElementID string

type Elements interface {
	GetType()
}

type Ident struct {
	ID ElementID `json:"id"`
}

type Element struct {
	Ident
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
		Ident
		Hidden bool   `json:"hidden"`
		Alias  string `json:"alias"`
	} `json:"routes"`
	Children []Ident `json:"children"`
}

type Block struct {
	Element
	Spec
	Node
	Position `json:"position"`
	Routes   []Ident `json:"routes"`
}

type Source struct {
	Element
	Spec
	Node
	Position `json:"position"`
	Routes   []Ident `json:"routes"`
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
	ID       *ElementID `json:"id"`
	Type     *string    `json:"type"`
	Spec     *string    `json:"spec"`
	Alias    *string    `json:"alias"`
	Position *Position  `json:"position"`
	Routes   []struct {
		ID     *ElementID `json:"id"`
		Hidden *bool      `json:"hidden"`
		Alias  *string    `json:"alias"`
	} `json:"routes"`
	Children []struct {
		ID *ElementID `json:"id"`
	} `json:"children"`
	SourceID *ElementID `json:"source_id"`
	TargetID *ElementID `json:"target_id"`
}

type UpdateElement struct {
	Alias    *string     `json:"alias"`
	Position *Position   `json:"position"`
	Value    interface{} `json:"value"`
	Hidden   *bool       `json:"hidden"`
}
