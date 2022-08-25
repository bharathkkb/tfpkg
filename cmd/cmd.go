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
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	"github.com/spf13/cobra"
)

var flags struct {
	genDir           string
	tmpDir           string
	genModuleVersion string
}

var rootCmd = &cobra.Command{
	Use:   "tfpkg",
	Short: "Generate go packages representing TF modules.",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("expected single module name as arg, got %+v", args)
		}
		return nil
	},
	RunE: func(_ *cobra.Command, args []string) error {
		if flags.tmpDir == "" {
			flags.tmpDir = os.TempDir()
		}
		return generateBlueprintPkg(flags.genDir, flags.tmpDir, args[0], flags.genModuleVersion)
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVar(&flags.genDir, "gen-dir", "generated", "Dir to generate packages.")
	rootCmd.Flags().StringVar(&flags.tmpDir, "tmp-dir", "", "Temp dir for downloading TF blueprints. Defaults to system temp directory.")
	rootCmd.Flags().StringVar(&flags.genModuleVersion, "version", "", "Optional version of module to download. Defaults to latest.")
}

// generateBlueprintPkg generates a go package for a given module src/version in genDir.
func generateBlueprintPkg(genDir, tmpDir, modSrc, modVersion string) error {
	// downoad module
	err := downloadModule(modSrc, modVersion, tmpDir)
	if err != nil {
		return fmt.Errorf("error downloading modules: %w", err)
	}
	// parse downloaded modules and filter out any internal modules
	downloadedMods := getDownloadedModules(tmpDir)
	filteredMods := filterDownloadedModules(downloadedMods.Modules)

	for _, m := range filteredMods {
		module, diags := tfconfig.LoadModule(path.Join(tmpDir, m.Dir))
		if diags.HasErrors() {
			return fmt.Errorf("error loading modules: %w", diags.Err())
		}

		pkgName := sourceToPkgName(m.Source)
		data := generatePkgFromMod(pkgName, modSrc, modVersion, module)
		if err := writeFile(path.Join(genDir, pkgName, fmt.Sprintf("%s.go", pkgName)), data); err != nil {
			return fmt.Errorf("error writing generated pkg: %w", err)
		}
	}
	return nil
}

// writeFile writes content to file path
func writeFile(p string, content string) error {
	err := os.MkdirAll(path.Dir(p), os.ModePerm)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(p, []byte(content), os.ModePerm)
}
