package tfgen

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
)

// TODO: support resource blocks
type Resource struct {
	name    string
	resType string
}

type ResourceOptions func(r *Resource)

func NewResource(resType, name string, opts ...ResourceOptions) *Resource {
	r := &Resource{name: name, resType: resType}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (r Resource) GetHCLBlock() (*hclwrite.Block, error) {
	b := NewBlock(moduleBlockType, r.name)
	hb, err := b.GetHCLBlock()
	if err != nil {
		return nil, err
	}
	return hb, nil
}
