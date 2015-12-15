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

type Group struct {
	Element
	Position `json:"position"`
	Routes   []struct {
		ID     ElementID `json:"id"`
		Hidden bool      `json:"hidden"`
		Alias  string    `json:"alias"`
	} `json:"routes"`
	Children []ID `json:"children"`
}

type Block struct {
	Element
	Spec     string `json:"spec"`
	Position `json:"position"`
	Routes   []ID `json:"routes"`
}

type Source struct {
	Element
	Spec     string `json:"spec"`
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
	Name      string        `json:"name"`
	Value     interface{}   `json:"value"`
	Direction string        `json:"direction"`
	Source    string        `json:"source"`
	JSONType  core.JSONType `json:"json_type":`
}

type CreateElement struct {
	ID        *ElementID     `json:"id"`
	Type      *string        `json:"type"`
	JSONType  *core.JSONType `json:"json_type"`
	Direction *string        `json:"direction"`
	Name      *string        `json:"name"`
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
