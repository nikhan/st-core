package server

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type Pattern struct {
	Blocks      []BlockLedger      `json:"blocks"`
	Connections []ConnectionLedger `json:"connections"`
	Groups      []Group            `json:"groups"`
	Sources     []SourceLedger     `json:"sources"`
	Links       []LinkLedger       `json:"links"`
}

type Node interface {
	GetID() int
	GetParent() *Group
	SetParent(*Group)
}

type Group struct {
	Id       int      `json:"id"`
	Label    string   `json:"label"`
	Children []int    `json:"children"`
	Parent   *Group   `json:"-"`
	Position Position `json:"position"`
}

type ProtoGroup struct {
	Group    int      `json:"group"`
	Children []int    `json:"children"`
	Label    string   `json:"label"`
	Position Position `json:"position"`
}

func (g *Group) GetID() int {
	return g.Id
}

func (g *Group) GetParent() *Group {
	return g.Parent
}

func (g *Group) SetParent(group *Group) {
	g.Parent = group
}

func (s *Server) ListGroups() []Group {
	groups := []Group{}
	for _, g := range s.groups {
		groups = append(groups, *g)
	}
	return groups
}

func (s *Server) GroupIndexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(s.ListGroups()); err != nil {
		panic(err)
	}
}

func (s *Server) DetachChild(g Node) error {
	parent := g.GetParent()
	if parent == nil {
		return errors.New("no parent to detach from")
	}

	id := g.GetID()

	child := -1
	for i, v := range parent.Children {
		if v == id {
			child = i
		}
	}

	if child == -1 {
		return errors.New("could not remove child from group: child does not exist")
	}

	parent.Children = append(parent.Children[:child], parent.Children[child+1:]...)

	update := struct {
		Id     int    `json:"id"`
		Child  int    `json:"child"`
		Action string `json:"string"`
	}{
		parent.GetID(), g.GetID(), DELETE,
	}

	s.websocketBroadcast(Update{Action: DELETE, Type: GROUP, Data: update})
	return nil
}

func (s *Server) AddChildToGroup(id int, n Node) error {
	newParent, ok := s.groups[id]
	if !ok {
		return errors.New("group not found")
	}

	nid := n.GetID()
	for _, v := range newParent.Children {
		if v == nid {
			return errors.New("node already child of this group")
		}
	}

	newParent.Children = append(newParent.Children, nid)
	if n.GetParent() != nil {
		err := s.DetachChild(n)
		if err != nil {
			return err
		}
	}

	n.SetParent(newParent)

	update := struct {
		Id     int    `json:"id"`
		Child  int    `json:"child"`
		Action string `json:"action"`
	}{
		id, nid, CREATE,
	}

	s.websocketBroadcast(Update{Action: UPDATE, Type: GROUP, Data: update})
	return nil
}

// CreateGroupHandler responds to a POST request to instantiate a new group and add it to the Server.
// Moves all of the specified children out of the parent's group and into the new group.
func (s *Server) GroupCreateHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(w, Error{"could not read request body"})
		return
	}

	var g ProtoGroup
	err = json.Unmarshal(body, &g)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(w, Error{"could not read JSON"})
		return
	}

	s.Lock()
	defer s.Unlock()

	newGroup, err := s.CreateGroup(g)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(w, Error{err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	writeJSON(w, newGroup)
}

func (s *Server) CreateGroup(g ProtoGroup) (*Group, error) {
	newGroup := &Group{
		Children: g.Children,
		Label:    g.Label,
		Id:       s.GetNextID(),
	}

	if newGroup.Children == nil {
		newGroup.Children = []int{}
	}

	for _, c := range newGroup.Children {
		_, okb := s.blocks[c]
		_, okg := s.groups[c]
		_, oks := s.sources[c]
		if !okb && !okg && !oks {
			return nil, errors.New("could not create group: invalid children")
		}
	}

	s.groups[newGroup.Id] = newGroup
	s.websocketBroadcast(Update{Action: CREATE, Type: GROUP, Data: newGroup})

	err := s.AddChildToGroup(g.Group, newGroup)
	if err != nil {
		return nil, err
	}

	for _, c := range newGroup.Children {
		if cb, ok := s.blocks[c]; ok {
			err = s.AddChildToGroup(newGroup.Id, cb)
		}
		if cg, ok := s.groups[c]; ok {
			err = s.AddChildToGroup(newGroup.Id, cg)
		}
		if cs, ok := s.sources[c]; ok {
			err = s.AddChildToGroup(newGroup.Id, cs)
		}
		if err != nil {
			return nil, err
		}
	}

	return newGroup, nil
}

func (s *Server) DeleteGroup(id int) error {
	group, ok := s.groups[id]
	if !ok {
		return errors.New("could not find group to delete")
	}

	for _, c := range group.Children {
		if _, ok := s.blocks[c]; ok {
			err := s.DeleteBlock(c)
			if err != nil {
				return err
			}
		} else if _, ok := s.groups[c]; ok {
			err := s.DeleteGroup(c)
			if err != nil {
				return err
			}
		}
	}

	update := struct {
		Id int `json:"id"`
	}{
		id,
	}
	s.DetachChild(group)
	delete(s.groups, id)
	s.websocketBroadcast(Update{Action: DELETE, Type: GROUP, Data: update})
	return nil
}

func (s *Server) GroupDeleteHandler(w http.ResponseWriter, r *http.Request) {

	id, err := getIDFromMux(mux.Vars(r))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(w, err)
		return
	}

	s.Lock()
	defer s.Unlock()

	err = s.DeleteGroup(id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(w, Error{err.Error()})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// returns a description of the group - its id and childreen
func (s *Server) GroupHandler(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromMux(mux.Vars(r))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(w, err)
		return
	}
	s.Lock()
	defer s.Unlock()
	group, ok := s.groups[id]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(w, Error{"could not find group"})
		return
	}
	w.WriteHeader(http.StatusOK)
	writeJSON(w, group)
}

func (s *Server) GroupExportHandler(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromMux(mux.Vars(r))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(w, err)
		return
	}

	s.Lock()
	defer s.Unlock()
	p, err := s.Export(id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(w, Error{err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	writeJSON(w, p)
}

func (s *Server) ExportGroup(id int) (*Pattern, error) {
	p := &Pattern{}
	g, ok := s.groups[id]
	if !ok {
		return nil, errors.New("could not find group to export")
	}

	p.Groups = append(p.Groups, *g)
	for _, c := range g.Children {
		if b, ok := s.blocks[c]; ok {
			p.Blocks = append(p.Blocks, *b)
			continue
		}
		if source, ok := s.sources[c]; ok {
			p.Sources = append(p.Sources, *source)
			continue
		}
		if group, ok := s.groups[c]; ok {
			g, err := s.ExportGroup(group.Id)
			if err != nil {
				return nil, err
			}

			p.Blocks = append(p.Blocks, g.Blocks...)
			p.Groups = append(p.Groups, g.Groups...)
			p.Sources = append(p.Sources, g.Sources...)
			continue
		}
	}
	return p, nil
}

func (s *Server) Export(id int) (*Pattern, error) {
	p, err := s.ExportGroup(id)
	if err != nil {
		return nil, err
	}

	ids := make(map[int]struct{})
	for _, b := range p.Blocks {
		ids[b.Id] = struct{}{}
	}

	for _, g := range p.Groups {
		ids[g.Id] = struct{}{}
	}

	for _, source := range p.Sources {
		ids[source.Id] = struct{}{}
	}

	for _, c := range s.connections {
		_, sourceIncluded := ids[c.Source.Id]
		_, targetIncluded := ids[c.Target.Id]
		if sourceIncluded && targetIncluded {
			p.Connections = append(p.Connections, *c)
		}
	}

	for _, l := range s.links {
		_, sourceIncluded := ids[l.Block]
		_, targetIncluded := ids[l.Source]
		if sourceIncluded && targetIncluded {
			p.Links = append(p.Links, *l)
		}
	}

	return p, nil
}

func (s *Server) GroupImportHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(w, Error{"could not read request body"})
		return
	}

	id, err := getIDFromMux(mux.Vars(r))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(w, err)
		return
	}

	var p Pattern
	err = json.Unmarshal(body, &p)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(w, Error{"could not unmarshal value"})
		return
	}

	s.Lock()
	defer s.Unlock()

	err = s.ImportGroup(id, p)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(w, Error{err.Error()})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) ImportGroup(id int, p Pattern) error {
	parents := make(map[int]int) // old child id / old parent id
	newIds := make(map[int]int)  // old id / new id

	if _, ok := s.groups[id]; !ok {
		return errors.New("could not attach to group: does not exist")
	}

	for _, g := range p.Groups {
		ng, err := s.CreateGroup(ProtoGroup{
			Label:    g.Label,
			Position: g.Position,
		})

		if err != nil {
			return err
		}

		newIds[g.Id] = ng.Id

		for _, c := range g.Children {
			parents[c] = g.Id
		}
	}

	for _, b := range p.Blocks {
		nb, err := s.CreateBlock(ProtoBlock{
			Label:    b.Label,
			Position: b.Position,
			Type:     b.Type,
		})

		if err != nil {
			return err
		}

		newIds[b.Id] = nb.Id
	}

	for _, source := range p.Sources {
		ns, err := s.CreateSource(ProtoSource{
			Label:    source.Label,
			Position: source.Position,
			Type:     source.Type,
		})

		if err != nil {
			return err
		}

		newIds[source.Id] = ns.Id
	}

	for _, c := range p.Connections {
		c.Source.Id = newIds[c.Source.Id]
		c.Target.Id = newIds[c.Target.Id]
		_, err := s.CreateConnection(ProtoConnection{
			Source: c.Source,
			Target: c.Target,
		})
		if err != nil {
			return err
		}
	}

	for _, l := range p.Links {
		l.Block = newIds[l.Block]
		l.Source = newIds[l.Source]
		_, err := s.CreateLink(ProtoLink{
			Source: l.Source,
			Block:  l.Block,
		})
		if err != nil {
			return err
		}
	}

	for _, source := range p.Sources {
		err := s.ModifySource(newIds[source.Id], source.Parameters)
		if err != nil {
			return err
		}
	}

	for _, g := range p.Groups {
		for _, c := range g.Children {
			var n Node
			if bn, ok := s.blocks[newIds[c]]; ok {
				n = bn
			}
			if bg, ok := s.groups[newIds[c]]; ok {
				n = bg
			}
			if bs, ok := s.sources[newIds[c]]; ok {
				n = bs
			}
			if n == nil {
				return errors.New("could not add node, node does not exist")
			}

			err := s.AddChildToGroup(newIds[g.Id], n)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Server) GroupModifyLabelHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(w, Error{"could not read request body"})
		return
	}

	id, err := getIDFromMux(mux.Vars(r))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(w, err)
		return
	}

	var l string
	err = json.Unmarshal(body, &l)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(w, Error{"could not unmarshal: " + string(body) + ""})
		return
	}

	s.Lock()
	defer s.Unlock()

	g, ok := s.groups[id]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(w, Error{"no block found"})
		return
	}

	g.Label = l

	update := struct {
		Label string `json:"label"`
		Id    int    `json:"id"`
	}{
		l, id,
	}

	s.websocketBroadcast(Update{Action: UPDATE, Type: GROUP, Data: update})

	w.WriteHeader(http.StatusNoContent)
}
func (s *Server) GroupModifyAllChildrenHandler(w http.ResponseWriter, r *http.Request) {
}
func (s *Server) GroupModifyChildHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	id, err := getIDFromMux(vars)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(w, err)
		return
	}

	childs, ok := vars["node_id"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(w, Error{"no ID supplied"})
		return
	}

	child, err := strconv.Atoi(childs)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(w, Error{err.Error()})
		return
	}

	if id == child {
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(w, Error{"cannot add group as member of itself"})
		return
	}

	s.Lock()
	defer s.Unlock()

	var n Node

	if _, ok := s.groups[id]; !ok {
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(w, Error{"could not find id"})
		return
	}

	if b, ok := s.blocks[child]; ok {
		n = b
	}
	if g, ok := s.groups[child]; ok {
		n = g
	}

	if n == nil {
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(w, Error{"could not find id"})
		return
	}

	err = s.AddChildToGroup(id, n)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(w, Error{err.Error()})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) GroupPositionHandler(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromMux(mux.Vars(r))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(w, err)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(w, Error{"could not read request body"})
		return
	}

	var p Position
	err = json.Unmarshal(body, &p)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(w, Error{"could not read JSON"})
		return
	}

	s.Lock()
	defer s.Unlock()

	g, ok := s.groups[id]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(w, Error{"could not find group"})
		return
	}

	g.Position = p

	update := struct {
		Position
		Id int
	}{
		p,
		id,
	}

	s.websocketBroadcast(Update{Action: UPDATE, Type: GROUP, Data: update})
	w.WriteHeader(http.StatusNoContent)
}
