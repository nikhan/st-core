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

type ElementID string
type ElementType string

type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type ElementItem struct {
	ID     *ElementID `json:"id"`
	Hidden *bool      `json:"hidden,omitempty"`
	Alias  *string    `json:"alias,omitempty"`
}

type ByID []*ElementItem

func (a ByID) Len() int           { return len(a) }
func (a ByID) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByID) Less(i, j int) bool { return *a[i].ID < *a[j].ID }

type Element struct {
	ID        *ElementID       `json:"id"`
	Type      *string          `json:"type,omitempty"`
	JSONType  *core.JSONType   `json:"json_type,omitempty"`
	Direction *string          `json:"direction,omitempty"`
	Name      *string          `json:"name,omitempty"`
	Source    *core.SourceType `json:"source,omitempty"`
	Spec      *string          `json:"spec,omitempty"`
	Alias     *string          `json:"alias,omitempty"`
	Position  *Position        `json:"position,omitempty"`
	Routes    []*ElementItem   `json:"routes,omitempty"`
	Children  []*ElementItem   `json:"children,omitempty"`
	SourceID  *ElementID       `json:"source_id,omitempty"`
	TargetID  *ElementID       `json:"target_id,omitempty"`
	Value     *core.InputValue `json:"value,omitempty"`
}

func (e *Element) GetRoute(id ElementID) (*ElementItem, bool) {
	for _, r := range e.Routes {
		if *r.ID == id {
			return r, true
		}
	}
	return nil, false
}

func (e *Element) isNode() bool {
	switch *e.Type {
	case BLOCK, GROUP, SOURCE:
		return true
	}
	return false
}

type Update struct {
	ID       *ElementID       `json:"id,omitempty"`
	Action   *string          `json:"action"`
	Data     []*Element       `json:"data,omitempty"`
	Alias    *string          `json:"alias,omitempty"`
	Position *Position        `json:"position,omitempty"`
	Value    *core.InputValue `json:"value,omitempty"`
	Hidden   *bool            `json:"hidden,omitempty"`
}
