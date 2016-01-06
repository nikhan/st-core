package stserver

/*func ref(c string) *string {
	return &c
}

func printGraph(g *Graph) {
	resp, _ := g.Get()
	o, _ := json.Marshal(resp)
	fmt.Println(string(o))
}

func TestAdd(t *testing.T) {
	g := NewGraph()

	// should error: element has no type
	_, err := g.Add([]*CreateElement{&CreateElement{}}, nil)
	if err == nil {
		t.Error("expected error")
	}
	fmt.Println(err)

	// should error: block as no spec
	_, err = g.Add([]*CreateElement{&CreateElement{
		Type: ref(BLOCK),
	}}, nil)
	if err == nil {
		t.Error("expected error")
	}
	fmt.Println(err)

	// should error: source has no spec
	_, err = g.Add([]*CreateElement{&CreateElement{
		Type: ref(SOURCE),
	}}, nil)
	if err == nil {
		t.Error("expected error")
	}
	fmt.Println(err)

	// create group
	_, err = g.Add([]*CreateElement{&CreateElement{
		Type: ref(GROUP),
	}}, nil)
	if err != nil {
		t.Error("error adding group")
	}
	fmt.Println(err)

	// create + block
	_, err = g.Add([]*CreateElement{&CreateElement{
		Type: ref(BLOCK),
		Spec: ref("+"),
	}}, nil)
	if err != nil {
		t.Error("error adding block")
	}
	fmt.Println(err)

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
	fmt.Println(err)

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
	fmt.Println(err)

	// should error: create connection no source or target
	_, err = g.Add([]*CreateElement{&CreateElement{
		Type: ref(CONNECTION),
	}}, nil)
	if err == nil {
		t.Error("expected error ")
	}
	fmt.Println(err)

	// should error: create connection no source or target
	_, err = g.Add([]*CreateElement{&CreateElement{
		Type: ref(LINK),
	}}, nil)
	if err == nil {
		t.Error("expected error ")
	}
	fmt.Println(err)

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
	fmt.Println(err)

	// link routes
	_, err = g.Add([]*CreateElement{&CreateElement{
		Type: ref(BLOCK),
		Spec: ref("valueGet"),
	}}, nil)
	if err != nil {
		t.Error("error adding block")
	}
	fmt.Println(err)

	_, err = g.Add([]*CreateElement{&CreateElement{
		Type: ref(SOURCE),
		Spec: ref("value"),
	}}, nil)
	if err != nil {
		t.Error("error adding source")
	}
	fmt.Println(err)

	source = ElementID("25")
	target = ElementID("23")
	_, err = g.Add([]*CreateElement{&CreateElement{
		Type:     ref(LINK),
		SourceID: &source,
		TargetID: &target,
	}}, nil)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(err)

	// should error: create connection with bad ids
	badID := ElementID("foo")
	_, err = g.Add([]*CreateElement{&CreateElement{
		Type:     ref(CONNECTION),
		SourceID: &badID,
		TargetID: &badID,
	}}, nil)
	if err == nil {
		t.Error("expected error ")
	}
	fmt.Println(err)

	// should error: create group with bad children
	_, err = g.Add([]*CreateElement{&CreateElement{
		Type:     ref(GROUP),
		Children: []ID{ID{badID}},
	}}, nil)
	if err == nil {
		t.Error("expected error")
	}
	fmt.Println(err)

	// should error: create group with bad routes
	_, err = g.Add([]*CreateElement{&CreateElement{
		Type: ref(GROUP),
		Routes: []GroupRoute{GroupRoute{
			ID: badID,
		}},
	}}, nil)
	if err == nil {
		t.Error("expected error")
	}
	fmt.Println(err)

	// should error: create source with bad routes
	_, err = g.Add([]*CreateElement{&CreateElement{
		Type: ref(SOURCE),
		Spec: ref("value"),
		Routes: []GroupRoute{GroupRoute{
			ID: badID,
		}},
	}}, nil)
	if err == nil {
		t.Error("expected error")
	}
	fmt.Println(err)

	// should error: create block with bad routes
	_, err = g.Add([]*CreateElement{&CreateElement{
		Type: ref(BLOCK),
		Spec: ref("+"),
		Routes: []GroupRoute{GroupRoute{
			ID: badID,
		}},
	}}, nil)
	if err == nil {
		t.Error("expected error")
	}
	fmt.Println(err)

	elements, err := g.Get()
	if err != nil {
		t.Error(err)
	}
	ids := make([]ElementID, len(elements))
	for i, e := range elements {
		ids[i] = e.GetID()
	}
	err = g.BatchDelete(ids)
	if err != nil {
		t.Error(err)
	}
}

func TestParent(t *testing.T) {

}

func TestPattern(t *testing.T) {
	g := NewGraph()

	first := ElementID("100")
	latch := ElementID("200")
	_, err := g.Add([]*CreateElement{
		&CreateElement{
			Type: ref(BLOCK),
			Spec: ref("first"),
			ID:   &first,
		},
		&CreateElement{
			Type: ref(BLOCK),
			Spec: ref("latch"),
			ID:   &latch,
		},
		&CreateElement{
			Type:     ref(GROUP),
			Children: []ID{ID{first}, ID{latch}},
		},
	}, nil)

	if err != nil {
		t.Error("error creating blocks")
	}

	printGraph(g)
}*/
