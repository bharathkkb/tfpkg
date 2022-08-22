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
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/bharathkkb/tfpkg/pkg/tfgen"
	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/iancoleman/strcase"
)

type DownloadedModules struct {
	Modules []DownloadedModule `json:"Modules"`
}

type DownloadedModule struct {
	Key     string `json:"Key"`
	Source  string `json:"Source"`
	Version string `json:"Version,omitempty"`
	Dir     string `json:"Dir"`
}

// downloadModule downloads all modules in sources to dir.
func downloadModule(source, version, dir string) error {
	// skip download for debugging
	if os.Getenv("tfpkg_SKIP_TF_DOWNLOAD") != "" {
		return nil
	}
	// validate tf installed
	tfPath, err := exec.LookPath("terraform")
	if err != nil {
		return fmt.Errorf("terraform not installed: %v", err)
	}

	// To download a module, we make a temporary config referencing the modules we want.
	// Then we terraform init to let TF download them to the .terraform folder within dir.
	blocks := make([]tfgen.HCLBlock, 0, 1)

	randBlockName := b64.RawStdEncoding.EncodeToString([]byte(source))
	opts := []tfgen.ModuleOptions{tfgen.ModuleWithSource(source)}
	if version != "" {
		opts = append(opts, tfgen.ModuleWithVersion(version))
	}
	tmpBlock := tfgen.NewModule(randBlockName, opts...)
	blocks = append(blocks, tmpBlock)

	tmpConfig := tfgen.NewRootModule(tfgen.RootModuleWithBlocks(map[string][]tfgen.HCLBlock{"tmp.tf": blocks}))
	err = tmpConfig.Write(dir)
	if err != nil {
		return fmt.Errorf("error writing tmp config in %s: %v", dir, err)
	}

	tfe, err := tfexec.NewTerraform(dir, tfPath)
	if err != nil {
		return fmt.Errorf("error running initializing Terraform: %v", err)
	}
	err = tfe.Init(context.Background(), tfexec.Upgrade(true))
	if err != nil {
		return fmt.Errorf("error running running terraform init: %v", err)
	}
	return nil
}

// getDownloadedModules inspects the downloaded TF modules to generate metadata.
func getDownloadedModules(dir string) DownloadedModules {
	modDir := path.Join(dir, ".terraform", "modules", "modules.json")
	file, err := ioutil.ReadFile(modDir)
	if err != nil {
		log.Fatalf("error reading %s: %v", modDir, err)
	}
	data := DownloadedModules{}
	err = json.Unmarshal([]byte(file), &data)
	if err != nil {
		log.Fatalf("error unmarshalling %s: %v", modDir, err)
	}
	return data
}

// filterDownloadedModules filters set of modules to only public modules skipping any internal modules.
func filterDownloadedModules(all []DownloadedModule) []DownloadedModule {
	filtered := []DownloadedModule{}
	for _, m := range all {
		if strings.HasPrefix(m.Source, "registry") {
			filtered = append(filtered, m)
		}
	}
	return filtered
}

// sourceToPkgName converts a registry source to an equivalent go package name.
// For example terraform-google-modules/project-factory/google becomes project_factory package.
func sourceToPkgName(source string) string {
	source = strings.TrimSuffix(source, "/google")
	modName := strings.Split(source, "/")[len(strings.Split(source, "/"))-1]
	return strcase.ToSnake(modName)
}
