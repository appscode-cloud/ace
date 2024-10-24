/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Community License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Community-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package installer

import (
	"fmt"

	"go.bytebuilders.dev/ace/pkg/cmds/installer/installer_precheck"

	"github.com/spf13/cobra"
)

func newCmdValidate() *cobra.Command {
	var optionsPath string
	cmd := &cobra.Command{
		Use:               "validate",
		Short:             "Validate installer options",
		DisableAutoGenTag: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if optionsPath == "" {
				return fmt.Errorf("options file missing")
			}
			allOK, err := installer_precheck.CheckOptions(optionsPath)
			if err != nil {
				return fmt.Errorf("validation failed. Reason: %w", err)
			}

			if !allOK {
				return fmt.Errorf("Validation failed due to multiple issues. Please check the details and try again.\n")
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&optionsPath, "options", "", "Path of the options file")
	return cmd
}
