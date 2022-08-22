package tfgen

import "github.com/hashicorp/hcl/v2/hclwrite"

// Block is a generic HCL block.
type Block struct {
	blockType string
	name      string
}

type HCLBlock interface {
	GetHCLBlock() (*hclwrite.Block, error)
}

type BlockOptions func(p *Block)

func NewBlock(blockType, name string, opts ...BlockOptions) *Block {
	b := &Block{blockType: blockType, name: name}
	for _, opt := range opts {
		opt(b)
	}
	return b
}

func (b *Block) GetHCLBlock() (*hclwrite.Block, error) {
	hb := hclwrite.NewBlock(b.blockType, []string{b.name})
	return hb, nil
}
