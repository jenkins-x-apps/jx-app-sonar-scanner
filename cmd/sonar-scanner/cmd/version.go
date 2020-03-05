package cmd

import (
	"fmt"
	"github.com/jenkins-x-apps/jx-app-sonar-scanner/internal/version"
	"github.com/spf13/cobra"
)

var (
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Displays the version of binary",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version.GetVersion())
		},
	}
)
