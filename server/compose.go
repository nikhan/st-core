package server

import (
	"errors"

	"github.com/nytlabs/st-core/core"
)

const (
	IN  = -1
	OUT = 1
)

// components returns connected subgraphs of the pattern as patterns
// only uses blocks, connections and ignores groups, sources, links
func (p *Pattern) components() []*Pattern {
	components := []*Pattern{}

	// caches to mark what has already been added to a subgraph/pattern
	blocks := make(map[int]BlockLedger)
	connections := make(map[int]ConnectionLedger)

	for _, b := range p.Blocks {
		blocks[b.Id] = b
	}

	for _, c := range p.Connections {
		connections[c.Id] = c
	}

	// traverses graph head and tail at the same time
	var connected func(BlockLedger) ([]BlockLedger, []ConnectionLedger)
	connected = func(block BlockLedger) ([]BlockLedger, []ConnectionLedger) {
		cblocks := []BlockLedger{block}
		cconns := []ConnectionLedger{}

		delete(blocks, block.Id)

		for _, c := range p.Connections {
			var traverseId int

			// if this connection includes our block
			if c.Source.Id == block.Id {
				traverseId = c.Target.Id
			} else if c.Target.Id == block.Id {
				traverseId = c.Source.Id
			} else {
				continue
			}

			// if we've already seen this block don't traverse
			if _, ok := blocks[traverseId]; !ok {
				continue
			}

			// get block's neighbors
			cconns = append(cconns, c)
			delete(connections, c.Id)
			tb, tc := connected(blocks[traverseId])
			cblocks = append(cblocks, tb...)
			cconns = append(cconns, tc...)
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

// composableSource checks to see if the incoming pattern can be composed
// based on its source. returns source type
func (p *Pattern) composableSource() (*core.SourceType, error) {
	library := core.GetLibrary()
	var source core.SourceType
	source = core.NONE

	if len(p.Sources) > 0 || len(p.Links) > 0 {
		return nil, errors.New("can only compose blocks and connections")
	}

	for _, b := range p.Blocks {
		spec, ok := library[b.Type]
		if !ok {
			return nil, errors.New("can only compose core blocks, must decompose compositions first")
		}

		if source == core.NONE && spec.Source != core.NONE {
			source = spec.Source
		} else if source != core.NONE && source != spec.Source {
			return nil, errors.New("composed blocks may only use one type of source")
		}
	}

	return &source, nil
}

func (p *Pattern) zerodegree(w int) []BlockLedger {
	blocks := make(map[int]BlockLedger)
	for _, b := range p.Blocks {
		blocks[b.Id] = b
	}

	for _, c := range p.Connections {
		for k, _ := range blocks {
			switch w {
			case IN:
				if c.Target.Id == k {
					delete(blocks, k)
				}
			case OUT:
				if c.Source.Id == k {
					delete(blocks, k)
				}
			}
		}
	}

	bs := make([]BlockLedger, len(blocks))
	i := 0
	for _, v := range blocks {
		bs[i] = v
		i++
	}
	return bs
}

func (p *Pattern) sinks() []BlockLedger {
	return p.zerodegree(OUT)

}

func (p *Pattern) sources() []BlockLedger {
	return p.zerodegree(IN)
}

func (p *Pattern) Spec() (*core.Spec, error) {
	_, err := p.composableSource()
	if err != nil {
		return nil, err
	}

	_ = p.components()

	return nil, nil
}
