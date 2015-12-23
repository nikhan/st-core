package stserver

import "testing"

func ref(c string) *string {
	return &c
}

func TestAdd(t *testing.T) {
	g := NewGraph()

	// should fail: element has no type
	_, err := g.Add([]*CreateElement{&CreateElement{}}, nil)
	if err == nil {
		t.Error("expected error")
	}

	// should fail: block as no spec
	_, err = g.Add([]*CreateElement{&CreateElement{
		Type: ref(BLOCK),
	}}, nil)
	if err == nil {
		t.Error("expected error")
	}

	// should fail: source has no spec
	_, err = g.Add([]*CreateElement{&CreateElement{
		Type: ref(SOURCE),
	}}, nil)
	if err == nil {
		t.Error("expected error")
	}

	// create group
	_, err = g.Add([]*CreateElement{&CreateElement{
		Type: ref(GROUP),
	}}, nil)
	if err != nil {
		t.Error("error adding group")
	}

	// create + block
	_, err = g.Add([]*CreateElement{&CreateElement{
		Type: ref(BLOCK),
		Spec: ref("+"),
	}}, nil)
	if err != nil {
		t.Error("error adding block")
	}

	// create + block with position
	_, err = g.Add([]*CreateElement{&CreateElement{
		Type: ref(BLOCK),
		Spec: ref("+"),
		Position: &Position{
			X: 99,
			Y: 99,
		},
	}}, nil)
	if err != nil {
		t.Error("error adding block")
	}

	// create + block with position and alias
	_, err = g.Add([]*CreateElement{&CreateElement{
		Alias: ref("TEST"),
		Type:  ref(BLOCK),
		Spec:  ref("+"),
		Position: &Position{
			X: 99,
			Y: 99,
		},
	}}, nil)
	if err != nil {
		t.Error("error adding block")
	}

	// create value source
	_, err = g.Add([]*CreateElement{&CreateElement{
		Type: ref(SOURCE),
		Spec: ref("value"),
	}}, nil)
	if err != nil {
		t.Error("error adding source")
	}

}
