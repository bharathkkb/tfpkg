package tfgen

import (
	"fmt"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

const variableBlockType = "variable"

// Variable represent an input variable block.
type Variable struct {
	name       string
	desc       string
	defaultVal cty.Value
}

type VariableOptions func(v *Variable)

func VariableWithDesc(d string) VariableOptions {
	return func(v *Variable) {
		v.desc = d
	}
}

func VariableWithDefaultVal(dv cty.Value) VariableOptions {
	return func(v *Variable) {
		v.defaultVal = dv
	}
}

func NewVariable(name string, opts ...VariableOptions) *Variable {
	v := &Variable{name: name}
	for _, opt := range opts {
		opt(v)
	}
	return v
}

func (v Variable) Ref() string {
	return fmt.Sprintf("${var.%s}", v.name)
}

func (v Variable) RenderHCLBlock() (*hclwrite.Block, error) {
	b := NewBlock(variableBlockType, v.name)
	hb, err := b.RenderHCLBlock()
	if err != nil {
		return nil, err
	}
	hb.Body().SetAttributeValue("description", cty.StringVal(v.desc))
	if v.defaultVal != cty.NilVal {
		hb.Body().SetAttributeValue("default", v.defaultVal)
	}
	return hb, nil
}
