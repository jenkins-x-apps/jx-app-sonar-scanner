package cmd

import (
	"github.com/jenkins-x-apps/jx-app-sonar-scanner/internal/logging"
	"github.com/jenkins-x-apps/jx-app-sonar-scanner/internal/pipeline"
	sonarutil "github.com/jenkins-x-apps/jx-app-sonar-scanner/internal/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	sqServerOptionName      = "sqServer"
	apiKeyOptionName        = "apiKey"
	scanonpreviewOptionName = "scanonpreview"
	scanonreleaseOptionName = "scanonrelease"
	contextOptionName       = "pipeline-context"
)

var (
	configureCmdLogger = logging.AppLogger().WithFields(log.Fields{"command": "configure"})
	sourceDir          = "."

	configureCmd = &cobra.Command{
		Use:   "configure",
		Short: "configures pom.xml and effective pipeline config",
		Run:   configure,
	}
	sqServer      string
	apiKey        string
	scanonpreview bool
	scanonrelease bool
	context       string
)

func init() {
	configureCmd.Flags().StringVar(&sqServer, sqServerOptionName, "", "The URL of your Sonarqube server instance including protocol and port.")
	_ = viper.BindPFlag(sqServerOptionName, configureCmd.Flags().Lookup(sqServerOptionName))
	viper.SetDefault(sqServerOptionName, "http://jx-sonarqube.sonarqube.svc.cluster.local:9000")

	configureCmd.Flags().StringVar(&apiKey, apiKeyOptionName, "", "The Sonarqube user token, if required by your server instance.")
	_ = viper.BindPFlag(apiKeyOptionName, configureCmd.Flags().Lookup(apiKeyOptionName))

	configureCmd.Flags().BoolVarP(&scanonpreview, scanonpreviewOptionName, "p", true, "Run Sonarqube scans against all preview builds.")
	_ = viper.BindPFlag(scanonpreviewOptionName, configureCmd.Flags().Lookup(scanonpreviewOptionName))

	configureCmd.Flags().BoolVarP(&scanonrelease, scanonreleaseOptionName, "r", true, "Run Sonarqube scans against all release builds.")
	_ = viper.BindPFlag(scanonreleaseOptionName, configureCmd.Flags().Lookup(scanonreleaseOptionName))

	configureCmd.Flags().StringVar(&context, contextOptionName, "", "The build context")
	_ = viper.BindPFlag(contextOptionName, configureCmd.Flags().Lookup(contextOptionName))
	viper.SetDefault(contextOptionName, "")
}

func configure(cmd *cobra.Command, args []string) {
	multiError := verify()
	if !multiError.Empty() {
		for _, err := range multiError.Errors {
			configureCmdLogger.Error(err.Error())
		}

		configureCmdLogger.Fatal("not all required parameters for this command execution specified")
	}

	pipelineExtender := pipeline.NewPatcher(sourceDir, viper.GetString(contextOptionName), sqServer, apiKey, scanonpreview, scanonrelease)
	err := pipelineExtender.ConfigurePipeline()
	if err != nil {
		configureCmdLogger.Fatal(err)
	}
}

func verify() sonarutil.MultiError {
	validationErrors := sonarutil.MultiError{}

	validationErrors.Collect(sonarutil.IsNotEmpty(viper.GetString(sqServerOptionName), sqServerOptionName))
	//	validationErrors.Collect(sonarutil.IsNotEmpty(viper.GetString(apiKeyOptionName), apiKeyOptionName))

	return validationErrors
}
