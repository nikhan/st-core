package stserver

import "testing"

func ref(c string) *string {
	return &c
}

func TestAdd(t *testing.T) {
	g := NewGraph()

	// should error: element has no type
	_, err := g.Add([]*CreateElement{&CreateElement{}}, nil)
	if err == nil {
		t.Error("expected error")
	}

	// should error: block as no spec
	_, err = g.Add([]*CreateElement{&CreateElement{
		Type: ref(BLOCK),
	}}, nil)
	if err == nil {
		t.Error("expected error")
	}

	// should error: source has no spec
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

	// should error: create connection no source or target
	_, err = g.Add([]*CreateElement{&CreateElement{
		Type: ref(CONNECTION),
	}}, nil)
	if err == nil {
		t.Error("expected error ")
	}

	// should error: create connection no source or target
	_, err = g.Add([]*CreateElement{&CreateElement{
		Type: ref(LINK),
	}}, nil)
	if err == nil {
		t.Error("expected error ")
	}

	// connect routes
	source := ElementID("12")
	target := ElementID("11")
	_, err = g.Add([]*CreateElement{&CreateElement{
		Type:     ref(CONNECTION),
		SourceID: &source,
		TargetID: &target,
	}}, nil)
	if err != nil {
		t.Error("error adding connection")
	}

	// link routes
	_, err = g.Add([]*CreateElement{&CreateElement{
		Type: ref(BLOCK),
		Spec: ref("valueGet"),
	}}, nil)
	if err != nil {
		t.Error("error adding block")
	}

	_, err = g.Add([]*CreateElement{&CreateElement{
		Type: ref(SOURCE),
		Spec: ref("value"),
	}}, nil)
	if err != nil {
		t.Error("expected error")
	}

	source = ElementID("22")
	target = ElementID("24")
	_, err = g.Add([]*CreateElement{&CreateElement{
		Type:     ref(LINK),
		SourceID: &source,
		TargetID: &target,
	}}, nil)
	if err != nil {
		t.Error("error add link")
	}

	// print graph status
	//resp, _ := g.Get()
	//o, _ := json.Marshal(resp)
	//fmt.Println(string(o))
}
