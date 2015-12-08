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
type Spec string

type Elements interface {
	GetType()
}

type Element struct {
	ID    ElementID   `json:"id"`
	Type  ElementType `json:"type"`
	Alias string      `json:"alias"`
}

func (e *Element) GetType() ElementType {
	return e.Type
}

func (e *Element) GetID() ElementID {
	return e.ID
}

type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type Group struct {
	Element
	Position `json:"position"`
	Routes   []struct {
		ElementID
		Hidden bool   `json:"hidden"`
		Alias  string `json:"alias"`
	} `json:"routes"`
	Children []struct {
		ID ElementID `json:"id"`
	} `json:"children"`
}

type Block struct {
	Element
	Spec     `json:"spec"`
	Position `json:"position"`
	Routes   []struct {
		ID ElementID `json:"id"`
	} `json:"routes"`
}

type Source struct {
	Element
	Spec     `json:"spec"`
	Position `json:"position"`
	Routes   []struct {
		ID ElementID `json:"id"`
	} `json:"routes"`
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
