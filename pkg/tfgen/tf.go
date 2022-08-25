package tfgen

import (
	"fmt"
	"os"
	"path"
	"reflect"
	"regexp"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

// RootModule represents a Terraform root configuration.
type RootModule struct {
	// map of blocks and filepath where the block should be written
	blocks  map[string][]HCLBlock
	vars    []*Variable
	outputs []*Output
}

type RootOptions func(r *RootModule)

func RootModuleWithBlocks(b map[string][]HCLBlock) RootOptions {
	return func(r *RootModule) {
		r.blocks = b
	}
}

func RootModuleWithVars(v []*Variable) RootOptions {
	return func(r *RootModule) {
		r.vars = v
	}
}

func RootModuleWithOutputs(o []*Output) RootOptions {
	return func(r *RootModule) {
		r.outputs = o
	}
}

func NewRootModule(opts ...RootOptions) *RootModule {
	r := &RootModule{}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// Write writes a root module to a directory
func (r *RootModule) Write(dir string) error {
	allBlocks := map[string][]HCLBlock{}
	for k, v := range r.blocks {
		allBlocks[k] = v
	}
	// write any vars to variables.tf
	allBlocks["variables.tf"] = make([]HCLBlock, 0)
	for _, v := range r.vars {
		allBlocks["variables.tf"] = append(allBlocks["variables.tf"], v)
	}
	// write any outputs to outputs.tf
	allBlocks["outputs.tf"] = make([]HCLBlock, 0)
	for _, o := range r.outputs {
		allBlocks["outputs.tf"] = append(allBlocks["outputs.tf"], o)
	}
	// ensure dir exists
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}

	for pth, blocks := range allBlocks {
		writePath := path.Join(dir, pth)
		f := hclwrite.NewEmptyFile()
		rootBody := f.Body()
		for _, b := range blocks {
			h, err := b.RenderHCLBlock()
			if err != nil {
				return err
			}
			rootBody.AppendBlock(h)
			// newlines btw blocks
			rootBody.AppendNewline()

		}
		hclFile, err := os.Create(writePath)
		if err != nil {
			return err
		}
		defer hclFile.Close()
		hclFile.Write(processTF(f.Bytes()))
	}
	return nil
}

var refParserRegexp = regexp.MustCompile(`\"\$\$\{(.*?)\}\"`)

// any post rendering processing
func processTF(b []byte) []byte {
	// format HCL
	b = hclwrite.Format(b)
	// hack for refs we wrap in "$${ref}"
	// $$ is escape in HCL
	// https://github.com/hashicorp/hcl/blob/2ebbb5d2dcc67881774d2b81d1497a55452e890f/hclsyntax/spec.md#template-literals
	return refParserRegexp.ReplaceAll(b, []byte(`$1`))
}

// toCtyVal types given value to a cty.Value for writing to HCL
func toCtyVal(v interface{}) (cty.Value, error) {
	if IsZeroOfUnderlyingType(v) {
		return cty.NilVal, nil
	}
	reflectV := reflect.ValueOf(v)
	switch reflectV.Kind() {
	case reflect.String:
		return cty.StringVal(reflectV.String()), nil
	case reflect.Bool:
		return cty.BoolVal(reflectV.Bool()), nil
	case reflect.Int:
		return cty.NumberIntVal(reflectV.Int()), nil
	case reflect.Slice:
		ctyVals := []cty.Value{}
		for i := 0; i < reflectV.Len(); i++ {
			vv := reflectV.Index(i)
			ctyVal, err := toCtyVal(vv.Interface())
			if err != nil {
				return cty.Value{}, err
			}
			ctyVals = append(ctyVals, ctyVal)
		}
		if len(ctyVals) > 0 {
			return cty.ListVal(ctyVals), nil
		}
		return cty.ListValEmpty(cty.String), nil
	case reflect.Map:
		m := map[string]cty.Value{}
		for _, k := range reflectV.MapKeys() {
			vv := reflectV.MapIndex(k)
			ctyVal, err := toCtyVal(vv.Interface())
			if err != nil {
				return cty.Value{}, err
			}
			m[k.String()] = ctyVal
		}
		return cty.ObjectVal(m), nil

	}
	return cty.Value{}, fmt.Errorf("unknown val %s", v)
}

func IsZeroOfUnderlyingType(x interface{}) bool {
	return reflect.DeepEqual(x, reflect.Zero(reflect.TypeOf(x)).Interface())
}
