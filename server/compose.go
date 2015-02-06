package server

import (
	"fmt"

	"github.com/nytlabs/st-core/core"
)

// components returns an array of array of ids
// each array of ids is a single connected component in the pattern graph
func (p *Pattern) components() [][]int {
	var connected func(int) []int
	components := [][]int{}
	blocks := make(map[int]BlockLedger)

	for _, b := range p.Blocks {
		blocks[b.Id] = b
	}

	// traverses graph head and tail at the same time
	// returns a list of block ids connected to a single block id
	connected = func(id int) []int {
		delete(blocks, id)
		ids := []int{id}
		for _, c := range p.Connections {
			if c.Source.Id == id {
				if _, ok := blocks[c.Target.Id]; !ok {
					continue
				}
				ids = append(ids, connected(c.Target.Id)...)
			}
			if c.Target.Id == id {
				if _, ok := blocks[c.Source.Id]; !ok {
					continue
				}
				ids = append(ids, connected(c.Source.Id)...)
			}
		}
		return ids
	}

	for len(blocks) > 0 {
		for k, _ := range blocks {
			components = append(components, connected(k))
		}
	}

	return components
}

func (p *Pattern) Spec() (*core.Spec, error) {

	c := p.components()
	fmt.Println(c)

	return nil, nil
}
