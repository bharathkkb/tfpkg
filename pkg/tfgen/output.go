package tfgen

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

const outputBlockType = "output"

// Output represents an output block.
type Output struct {
	name string
	val  string
	desc string
}

type OutputOptions func(o *Output)

func OutputWithDesc(d string) OutputOptions {
	return func(o *Output) {
		o.desc = d
	}
}

func NewOutput(name string, val string, opts ...OutputOptions) *Output {
	o := &Output{name: name, val: val}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

func (o Output) GetHCLBlock() (*hclwrite.Block, error) {
	b := NewBlock(outputBlockType, o.name)
	hb, err := b.GetHCLBlock()
	if err != nil {
		return nil, err
	}
	hb.Body().SetAttributeValue("value", cty.StringVal(o.val))
	if o.desc != "" {
		hb.Body().SetAttributeValue("description", cty.StringVal(o.desc))
	}
	return hb, nil
}
