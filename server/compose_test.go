package server

import (
	"encoding/json"
	"fmt"
	"testing"
)

type PatternTest struct {
	PatternInput       string
	ExpectedComponents int
}

var tests = []PatternTest{
	PatternTest{
		PatternInput:       `{"blocks":[{"label":"","type":"+","id":1,"inputs":[{"name":"addend","type":"fetch","value":"."},{"name":"addend","type":"fetch","value":"."}],"outputs":[{"name":"sum"}],"position":{"x":0,"y":0}},{"label":"","type":"+","id":2,"inputs":[{"name":"addend","type":"fetch","value":"."},{"name":"addend","type":"fetch","value":"."}],"outputs":[{"name":"sum"}],"position":{"x":0,"y":0}},{"label":"","type":"+","id":3,"inputs":[{"name":"addend","type":"fetch","value":"."},{"name":"addend","type":"fetch","value":"."}],"outputs":[{"name":"sum"}],"position":{"x":0,"y":0}}],"connections":[{"source":{"id":2,"route":0},"target":{"id":1,"route":0},"id":4},{"source":{"id":3,"route":0},"target":{"id":1,"route":0},"id":5}],"groups":[{"id":0,"label":"root","children":[1,2,3],"position":{"x":0,"y":0}}],"sources":null,"links":null}`,
		ExpectedComponents: 1,
	},
	PatternTest{
		PatternInput:       `{"blocks":[{"label":"","type":"+","id":1,"inputs":[{"name":"addend","type":"fetch","value":"."},{"name":"addend","type":"fetch","value":"."}],"outputs":[{"name":"sum"}],"position":{"x":0,"y":0}},{"label":"","type":"+","id":2,"inputs":[{"name":"addend","type":"fetch","value":"."},{"name":"addend","type":"fetch","value":"."}],"outputs":[{"name":"sum"}],"position":{"x":0,"y":0}},{"label":"","type":"+","id":3,"inputs":[{"name":"addend","type":"fetch","value":"."},{"name":"addend","type":"fetch","value":"."}],"outputs":[{"name":"sum"}],"position":{"x":0,"y":0}},{"label":"","type":"+","id":6,"inputs":[{"name":"addend","type":"fetch","value":"."},{"name":"addend","type":"fetch","value":"."}],"outputs":[{"name":"sum"}],"position":{"x":0,"y":0}}],"connections":[{"source":{"id":2,"route":0},"target":{"id":1,"route":0},"id":4},{"source":{"id":3,"route":0},"target":{"id":1,"route":0},"id":5}],"groups":[{"id":0,"label":"root","children":[1,2,3,6],"position":{"x":0,"y":0}}],"sources":null,"links":null}`,
		ExpectedComponents: 2,
	},
	PatternTest{
		PatternInput:       `{"blocks":[{"label":"","type":"+","id":1,"inputs":[{"name":"addend","type":"fetch","value":"."},{"name":"addend","type":"fetch","value":"."}],"outputs":[{"name":"sum"}],"position":{"x":0,"y":0}},{"label":"","type":"+","id":2,"inputs":[{"name":"addend","type":"fetch","value":"."},{"name":"addend","type":"fetch","value":"."}],"outputs":[{"name":"sum"}],"position":{"x":0,"y":0}},{"label":"","type":"+","id":3,"inputs":[{"name":"addend","type":"fetch","value":"."},{"name":"addend","type":"fetch","value":"."}],"outputs":[{"name":"sum"}],"position":{"x":0,"y":0}},{"label":"","type":"+","id":6,"inputs":[{"name":"addend","type":"fetch","value":"."},{"name":"addend","type":"fetch","value":"."}],"outputs":[{"name":"sum"}],"position":{"x":0,"y":0}},{"label":"","type":"+","id":7,"inputs":[{"name":"addend","type":"fetch","value":"."},{"name":"addend","type":"fetch","value":"."}],"outputs":[{"name":"sum"}],"position":{"x":0,"y":0}}],"connections":[{"source":{"id":2,"route":0},"target":{"id":1,"route":0},"id":4},{"source":{"id":3,"route":0},"target":{"id":1,"route":0},"id":5}],"groups":[{"id":0,"label":"root","children":[1,2,3,6,7],"position":{"x":0,"y":0}}],"sources":null,"links":null}`,
		ExpectedComponents: 3,
	},
}

func TestPatternComponents(t *testing.T) {
	for _, p := range tests {
		var po Pattern
		err := json.Unmarshal([]byte(p.PatternInput), &po)
		if err != nil {
			fmt.Println(err)
		}
		c := po.components()
		if len(c) != p.ExpectedComponents {
			t.Fail()
		}
	}
}

func TestPatternSpec(t *testing.T) {
	for _, p := range tests {
		var po Pattern
		err := json.Unmarshal([]byte(p.PatternInput), &po)
		if err != nil {
			fmt.Println(err)
		}

		s, err := po.Spec()
		if err != nil {
			t.Fail()
		}
		fmt.Printf("%+v\n", s)
	}

}
