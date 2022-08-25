package tfgen

import "github.com/hashicorp/hcl/v2/hclwrite"

const providerBlockType = "provider"

// Provider represents an provider config/instantiation block.
type Provider struct {
	name    string
	version string
}

type ProviderOptions func(p *Provider)

func ProviderWithVersion(v string) ProviderOptions {
	return func(p *Provider) {
		p.version = v
	}
}

func NewProvider(name string, opts ...ProviderOptions) *Provider {
	p := &Provider{name: name}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

func (p Provider) GetHCLBlock() (*hclwrite.Block, error) {
	b := NewBlock(providerBlockType, p.name)
	return b.GetHCLBlock()
}
