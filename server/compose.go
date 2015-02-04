package server

import "github.com/nytlabs/st-core/core"

func (p *Pattern) findComponents() ([]*Pattern, error) {

	return []*Pattern{}, nil
}

func (p *Pattern) findLastNode() (*BlockLedger, error) {

	return &BlockLedger{}, nil
}

func (p *Pattern) tail(b *BlockLedger) []*BlockLedger {

	return []*BlockLedger{}
}

func (p *Pattern) head() (*BlockLedger, error) {
	return &BlockLedger{}, nil
}

func (p *Pattern) Spec() (*core.Spec, error) {

	patterns, err := p.findComponents()
	if err != nil {
		return nil, err
	}

	for i, p := range patterns {
		head, err := p.head()
		if err != nil {
			return nil, err
		}
		_ = head
	}

	return nil, nil
}
