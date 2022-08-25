// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"strings"
	"testing"

	"github.com/dave/jennifer/jen"
	"github.com/iancoleman/strcase"
)

func Test_getStructDef(t *testing.T) {

	tests := []struct {
		name   string
		inType string
		want   string
	}{
		{
			name:   "simple",
			inType: "bool",
			want:   "Simple bool",
		},
		{
			name:   "testnum",
			inType: "number",
			want:   "Testnum int",
		},
		{
			name:   "listnum",
			inType: "list(number)",
			want:   "Listnum []string",
		},
		{
			name:   "mapnum",
			inType: "map(number)",
			want:   "Mapnum map[string]string",
		},
		{
			name:   "maplistnum",
			inType: "map(list(string))",
			want:   "Maplistnum map[string][]string",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			varStructName := strcase.ToCamel(tt.name)
			j := jen.Id(varStructName)
			getGoTypeFromTFType(tt.inType, j)
			gotWithStruct := jen.Type().Id("Test").Struct(j).GoString()
			converted := strings.TrimSpace(strings.Split(gotWithStruct, "\n")[1])
			if converted != tt.want {
				t.Errorf("getStructDef() got = %v, want %v", converted, tt.want)
			}
		})
	}
}
