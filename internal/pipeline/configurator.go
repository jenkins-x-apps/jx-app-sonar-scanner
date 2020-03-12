package pipeline

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"

	"github.com/jenkins-x-apps/jx-app-sonar-scanner/internal/logging"
	"github.com/jenkins-x-apps/jx-app-sonar-scanner/internal/util"
	"github.com/jenkins-x-apps/jx-app-sonar-scanner/internal/version"
	"github.com/jenkins-x/jx/pkg/config"
	"github.com/jenkins-x/jx/pkg/tekton/syntax"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var (
	logger = logging.AppLogger().WithFields(log.Fields{"component": "meta-pipeline-extender"})
)

// MetaPipelineConfigurator is responsible for injecting a new application step into the pipeline.
type MetaPipelineConfigurator struct {
	sourceDir     string
	context       string
	sqServer      string
	apiKey        string
	scanonpreview bool
	scanonrelease bool
}

// NewMetaPipelineConfigurator creates a new instance of MetaPipelineConfigurator.
func NewMetaPipelineConfigurator(sourceDir string, context string, sqServer string, apiKey string, scanonpreview bool, scanonrelease bool) MetaPipelineConfigurator {
	return MetaPipelineConfigurator{
		sourceDir:     sourceDir,
		context:       context,
		sqServer:      sqServer,
		apiKey:        apiKey,
		scanonpreview: scanonpreview,
		scanonrelease: scanonrelease,
	}
}

// ConfigurePipeline configures the Jenkins-X pipeline.
func (e *MetaPipelineConfigurator) ConfigurePipeline() error {
	log.WithFields(log.Fields{
		"dir": e.sourceDir,
	}).Info("Processing directory")

	if !util.IsDirectory(e.sourceDir) {
		return errors.Errorf("specified directory '%s' does not exist", e.sourceDir)
	}

	effectiveConfig := "jenkins-x-effective.yml"
	if e.context != "" {
		effectiveConfig = fmt.Sprintf("jenkins-x-%s-effective.yml", e.context)
	}

	pipelineConfigPath := filepath.Join(e.sourceDir, effectiveConfig)
	if !util.Exists(pipelineConfigPath) {
		return errors.Errorf("unable to find effective pipeline config in '%s'", e.sourceDir)
	}

	log.WithFields(log.Fields{
		"pipelineConfigPath": pipelineConfigPath,
	}).Info("path")

	// Dump pipeline to log to check input format
	fmt.Println("---------------------------INPUT PIPELINE---------------------------")
	content, err := ioutil.ReadFile(pipelineConfigPath)
	if err != nil {
		return errors.Errorf("unable to display pipeline config '%s'", pipelineConfigPath)
	}
	fmt.Println(string(content))
	fmt.Println("--------------------------------------------------------------------")

	projectConfig, err := config.LoadProjectConfigFile(pipelineConfigPath)
	if err != nil {
		return errors.Wrap(err, "unable to load pipeline configuration")
	}

	err = e.insertApplicationStep(projectConfig)
	if err != nil {
		return errors.Wrap(err, "unable to enhance pipeline with sonar-scanner configuration")
	}

	err = e.writeProjectConfig(projectConfig, pipelineConfigPath)
	if err != nil {
		return errors.Wrap(err, "unable to write modified project config")
	}

	// Dump pipeline to log to check output format
	fmt.Println("--------------------------OUTPUT PIPELINE---------------------------")
	content, err = ioutil.ReadFile(pipelineConfigPath)
	if err != nil {
		return errors.Errorf("unable to display pipeline config '%s'", pipelineConfigPath)
	}
	fmt.Println(string(content))
	fmt.Println("--------------------------------------------------------------------")

	return nil
}

func (e *MetaPipelineConfigurator) insertApplicationStep(projectConfig *config.ProjectConfig) error {
	// Insert into Preview pipeline if enabled
	if e.scanonpreview && projectConfig.PipelineConfig.Pipelines.PullRequest != nil {

		log.WithFields(log.Fields{
			"pipelineKind": "PullRequest",
		}).Info("pipeline")

		stages := projectConfig.PipelineConfig.Pipelines.PullRequest.Pipeline.Stages

		applicationStep := e.createApplicationStep()
		found := false
		// Iterate over all stages that may be present
		for s, stg := range stages {
			log.WithFields(log.Fields{
				"stage": stg.Name,
			}).Info("stage")

			steps := stg.Steps
			// Parse through the steps looking for the one that compiles the code
			for i, step := range steps {
				log.WithFields(log.Fields{
					"step": step.Name,
				}).Info("step")

				if step.Name == "build-make-linux" {
					// Insert step j after this step i
					j := i + 1

					thisStage := &projectConfig.PipelineConfig.Pipelines.PullRequest.Pipeline.Stages[s]
					thisStage.Steps = append(thisStage.Steps, syntax.Step{})
					copy(thisStage.Steps[j+1:], thisStage.Steps[j:])
					thisStage.Steps[j] = applicationStep

					log.WithFields(log.Fields{
						"step": step.Name,
					}).Info("matched")

					// Done
					found = true
					break
				}
			}
		}

		if found == false {
			log.Warn("Failed to find a build step in PullRequest pipeline to insert scannner after")
		}
	}

	// Insert into Release pipeline if enabled
	if e.scanonrelease && projectConfig.PipelineConfig.Pipelines.Release != nil {

		log.WithFields(log.Fields{
			"pipelineKind": "Release",
		}).Info("pipeline")

		stages := projectConfig.PipelineConfig.Pipelines.Release.Pipeline.Stages

		applicationStep := e.createApplicationStep()
		found := false
		// Iterate over all stages that may be present
		for s, stg := range stages {
			log.WithFields(log.Fields{
				"stage": stg.Name,
			}).Info("stage")

			steps := stg.Steps
			// Parse through the steps looking for the one that compiles the code
			for i, step := range steps {
				log.WithFields(log.Fields{
					"step": step.Name,
				}).Info("step")

				if step.Name == "build-make-build" {
					// Insert step j after this step i
					j := i + 1

					thisStage := &projectConfig.PipelineConfig.Pipelines.Release.Pipeline.Stages[s]
					thisStage.Steps = append(thisStage.Steps, syntax.Step{})
					copy(thisStage.Steps[j+1:], thisStage.Steps[j:])
					thisStage.Steps[j] = applicationStep

					log.WithFields(log.Fields{
						"step": step.Name,
					}).Info("matched")

					// Done
					found = true
					break
				}
			}
		}

		if found == false {
			log.Warn("Failed to find a build step in Release pipeline to insert scannner after")
		}
	}
	return nil
}

func (e *MetaPipelineConfigurator) writeProjectConfig(projectConfig *config.ProjectConfig, pipelineConfigPath string) error {
	err := util.CopyFile(pipelineConfigPath, pipelineConfigPath+".sonar-scanner.orig")
	if err != nil {
		return errors.Wrapf(err, "unable to backup '%s'", pipelineConfigPath)
	}

	logger.Infof("writing '%s'", pipelineConfigPath)
	err = projectConfig.SaveConfig(pipelineConfigPath)
	if err != nil {
		return errors.Wrapf(err, "unable to write '%s'", pipelineConfigPath)
	}
	return nil
}

func (e *MetaPipelineConfigurator) createApplicationStep() syntax.Step {
	args := []string{}
	if e.sqServer != "" {
		args = append(args, "-s "+e.sqServer)
	}
	if e.apiKey != "" {
		args = append(args, "-k "+e.apiKey)
	}
	args = append(args, "-r "+strconv.FormatBool(e.scanonrelease))
	args = append(args, "-p "+strconv.FormatBool(e.scanonpreview))

	step := syntax.Step{
		Name:      "sonar-scanner",
		Image:     version.GetFQImage(),
		Command:   "exec-sonar-scanner.sh",
		Arguments: args,
	}
	return step
}
