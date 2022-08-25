package tfgen

import (
	"fmt"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

const moduleBlockType = "module"

// Module represents a module instantiation of an remote or local module.
type Module struct {
	name       string
	source     string
	version    string
	attributes map[string]interface{}
}

type ModuleOptions func(m *Module)

func ModuleWithVersion(v string) ModuleOptions {
	return func(m *Module) {
		m.version = v
	}
}

func ModuleWithSource(s string) ModuleOptions {
	return func(m *Module) {
		m.source = s
	}
}

func ModuleWithAttributes(a map[string]interface{}) ModuleOptions {
	return func(m *Module) {
		m.attributes = a
	}
}

func NewModule(name string, opts ...ModuleOptions) *Module {
	m := &Module{name: name}
	for _, opt := range opts {
		opt(m)
	}
	if m.attributes == nil {
		m.attributes = make(map[string]interface{})
	}
	return m
}
func (m *Module) AddAttribute(k string, v interface{}) {
	if m.attributes == nil {
		m.attributes = make(map[string]interface{})
	}
	m.attributes[k] = v
}

func (m *Module) Ref(r string) string {
	return fmt.Sprintf("${module.%s.%s}", m.name, r)
}

func (m Module) RenderHCLBlock() (*hclwrite.Block, error) {
	b := NewBlock(moduleBlockType, m.name)
	hb, err := b.RenderHCLBlock()
	if err != nil {
		return nil, err
	}
	hb.Body().SetAttributeValue("source", cty.StringVal(m.source))
	if m.version != "" {
		hb.Body().SetAttributeValue("version", cty.StringVal(m.version))
	}
	hb.Body().AppendNewline()
	for k, v := range m.attributes {
		ctyV, err := toCtyVal(v)
		if err != nil {
			return nil, err
		}
		if ctyV.IsNull() {
			continue
		}
		hb.Body().SetAttributeValue(k, ctyV)
	}
	return hb, nil
}
