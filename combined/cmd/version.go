package cmd

import (
	"fmt"

	"github.com/netbirdio/netbird/version"
	"github.com/spf13/cobra"
)

// Build-time injected values. Kept in a dedicated file so upstream merges
// stay conflict-free: upstream combined/cmd has no version command and this
// file never touches root.go / main.go.
//
// Build with:
//
//	go build -ldflags "-X github.com/netbirdio/netbird/version.version=v0.76.3 \
//	  -X github.com/netbirdio/netbird/combined/cmd.commit=<hash> \
//	  -X github.com/netbirdio/netbird/combined/cmd.date=<RFC3339> \
//	  -X github.com/netbirdio/netbird/combined/cmd.builtBy=<builder>"
var (
	commit  = "unknown"
	date    = "unknown"
	builtBy = "unknown"
)

func init() {
	// The root command marks --config as required (persistent flag), which
	// would force `version` to need a config file. Shadow it with a local,
	// optional --config so the version command works standalone. Only this
	// command is affected — every other command still inherits the required
	// persistent flag.
	versionCmd.Flags().StringVarP(&configPath, "config", "c", "", "path to YAML configuration file (not used by version)")
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the server version, upstream tag and build commit",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("Upstream tag: %s\n", version.NetbirdVersion())
		fmt.Printf("Commit:       %s\n", commit)
		fmt.Printf("Built:        %s\n", date)
		fmt.Printf("BuiltBy:      %s\n", builtBy)
		return nil
	},
}
