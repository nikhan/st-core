package server

import (
	"fmt"

	"github.com/nytlabs/st-core/core"
)

// components returns connected subgraphs of the pattern as patterns
// only uses blocks, graphs and ignores groups, sources, links
func (p *Pattern) components() []*Pattern {
	var connected func(BlockLedger) ([]BlockLedger, []ConnectionLedger)
	components := []*Pattern{}
	blocks := make(map[int]BlockLedger)
	connections := make(map[int]ConnectionLedger)

	for _, b := range p.Blocks {
		blocks[b.Id] = b
	}

	for _, c := range p.Connections {
		connections[c.Id] = c
	}

	// traverses graph head and tail at the same time
	connected = func(block BlockLedger) ([]BlockLedger, []ConnectionLedger) {
		cblocks := []BlockLedger{block}
		cconns := []ConnectionLedger{}

		delete(blocks, block.Id)

		for _, c := range p.Connections {
			if c.Source.Id == block.Id {
				if _, ok := blocks[c.Target.Id]; !ok {
					continue
				}
				cconns = append(cconns, c)
				delete(connections, c.Id)
				tb, tc := connected(blocks[c.Target.Id])
				cblocks = append(cblocks, tb...)
				cconns = append(cconns, tc...)
			}
			if c.Target.Id == block.Id {
				if _, ok := blocks[c.Source.Id]; !ok {
					continue
				}
				cconns = append(cconns, c)
				delete(connections, c.Id)
				tb, tc := connected(blocks[c.Source.Id])
				cblocks = append(cblocks, tb...)
				cconns = append(cconns, tc...)
			}
		}
		return cblocks, cconns
	}

	for len(blocks) > 0 {
		for _, b := range blocks {
			pb, pc := connected(b)
			components = append(components, &Pattern{
				Blocks:      pb,
				Connections: pc,
			})

		}
	}

	return components
}

func (p *Pattern) Spec() (*core.Spec, error) {

	c := p.components()
	fmt.Println(c)

	return nil, nil
}
