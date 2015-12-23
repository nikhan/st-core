package stserver

import "github.com/nytlabs/st-core/core"

const (
	BLOCK      = "block"
	GROUP      = "group"
	SOURCE     = "source"
	CONNECTION = "connection"
	LINK       = "link"
	ROUTE      = "route"
	INPUT      = "input"
	OUTPUT     = "output"
)

//type ElementType string
type ElementID string

type Elements interface {
	SetID(ElementID)
	SetType(string)
	SetAlias(string)
	GetType() string
	GetID() ElementID
}

type Nodes interface {
	SetPosition(Position)
	GetPosition() Position
	GetRoutes() []ID
	SetParent(*ElementID)
	GetParent() *ElementID
}

type Element struct {
	ID    ElementID `json:"id"`
	Type  string    `json:"type"`
	Alias string    `json:"alias"`
}

func (e *Element) SetID(id ElementID) {
	e.ID = id
}

func (e *Element) SetType(t string) {
	e.Type = t
}

func (e *Element) SetAlias(alias string) {
	e.Alias = alias
}

func (e *Element) GetType() string {
	return e.Type
}

func (e *Element) GetID() ElementID {
	return e.ID
}

func (e *Element) GetAlias() string {
	return e.Alias
}

type ID struct {
	ID ElementID `json:"id"`
}

type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

func (p *Position) SetPosition(pos Position) {
	p.X = pos.X
	p.Y = pos.Y
}

func (p *Position) GetPosition() Position {
	return *p
}

type GroupRoute struct {
	ID     ElementID `json:"id"`
	Hidden bool      `json:"hidden"`
	Alias  string    `json:"alias"`
}

type Group struct {
	Element
	Parent   *ElementID `json:"-"`
	Position `json:"position"`
	Routes   []GroupRoute `json:"routes"`
	Children []ID         `json:"children"`
}

func (g *Group) GetParent() *ElementID {
	return g.Parent
}

func (g *Group) SetParent(parent *ElementID) {
	g.Parent = parent
}

func (g *Group) GetRoutes() []ID {
	ids := make([]ID, len(g.Routes))
	for i, route := range g.Routes {
		ids[i] = ID{route.ID}
	}
	return ids
}

type Block struct {
	Element
	Parent   *ElementID `json:"-"`
	Spec     string     `json:"spec"`
	Position `json:"position"`
	Routes   []ID `json:"routes"`
}

func (b *Block) GetParent() *ElementID {
	return b.Parent
}

func (b *Block) SetParent(parent *ElementID) {
	b.Parent = parent
}

func (b *Block) GetRoutes() []ID {
	return b.Routes
}

type Source struct {
	Element
	Parent   *ElementID `json:"-"`
	Spec     string     `json:"spec"`
	Position `json:"position"`
	Routes   []ID `json:"routes"`
}

func (s *Source) GetParent() *ElementID {
	return s.Parent
}

func (s *Source) SetParent(parent *ElementID) {
	s.Parent = parent
}

func (s *Source) GetRoutes() []ID {
	return s.Routes
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
	Name      string        `json:"name"`
	Value     interface{}   `json:"value"`
	Direction string        `json:"direction"`
	Source    string        `json:"source"`
	JSONType  core.JSONType `json:"json_type"`
}

type CreateElement struct {
	ID        *ElementID     `json:"id"`
	Type      *string        `json:"type"`
	JSONType  *core.JSONType `json:"json_type"`
	Direction *string        `json:"direction"`
	Name      *string        `json:"name"`
	Source    *string        `json:"source"`
	Spec      *string        `json:"spec"`
	Alias     *string        `json:"alias"`
	Position  *Position      `json:"position"`
	Routes    []struct {
		ID     *ElementID `json:"id"`
		Hidden *bool      `json:"hidden"`
		Alias  *string    `json:"alias"`
	} `json:"routes"`
	Children []ID
	SourceID *ElementID `json:"source_id"`
	TargetID *ElementID `json:"target_id"`
}

type UpdateElement struct {
	Alias    *string     `json:"alias"`
	Position *Position   `json:"position"`
	Value    interface{} `json:"value"`
	Hidden   *bool       `json:"hidden"`
}
