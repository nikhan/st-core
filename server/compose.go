package server

import (
	"fmt"

	"github.com/nytlabs/st-core/core"
)

func (p *Pattern) components() ([]*Pattern, error) {
	blocks := make([]BlockLedger, len(p.Blocks))
	copy(blocks, p.Blocks)
	components := []*Pattern{}

	var connected func(int) []int
	connected = func(id int) []int {
		ids := []int{}
		for _, c := range p.Connections {
			if c.Source.Id == id {
				ids = append(ids, connected(c.Target.Id)...)
			}
			if c.Target.Id == id {
				ids = append(ids, connected(c.Target.Id)...)
			}
		}
		return ids
	}
	fmt.Println(blocks)
	for len(blocks) > 0 {
		cp := &Pattern{}
		start, blocks := blocks[len(blocks)-1], blocks[:len(blocks)-1]
		cp.Blocks = append(cp.Blocks, start)

		ids := connected(start.Id)
		newBlocks := []BlockLedger{}
		for _, b := range blocks {
			remove := false
			for _, id := range ids {
				if id == b.Id {
					cp.Blocks = append(cp.Blocks, b)
					remove = true
				}
			}
			if remove {
				fmt.Println("REMOVED")
				continue
			}
			newBlocks = append(newBlocks, b)
		}
		blocks = newBlocks
		components = append(components, cp)
	}
	return components, nil
}

func (p *Pattern) Spec() (*core.Spec, error) {

	_, err := p.components()
	if err != nil {
		return nil, err
	}

	return nil, nil
}
