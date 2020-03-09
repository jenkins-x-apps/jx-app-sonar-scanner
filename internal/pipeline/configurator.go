package pipeline

import (
	"fmt"
	"path/filepath"

	"github.com/jenkins-x-apps/jx-app-sonar-scanner/internal/logging"
	"github.com/jenkins-x-apps/jx-app-sonar-scanner/internal/util"
	"github.com/jenkins-x-apps/jx-app-sonar-scanner/internal/version"
	"github.com/jenkins-x/jx/pkg/config"
	"github.com/jenkins-x/jx/pkg/jenkinsfile"
	"github.com/jenkins-x/jx/pkg/tekton/syntax"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var (
	logger = logging.AppLogger().WithFields(log.Fields{"component": "meta-pipeline-extender"})
)

// MetaPipelineConfigurator is responsible for injecting a new application step into the pipeline.
type MetaPipelineConfigurator struct {
	sourceDir string
	context   string
}

// NewMetaPipelineConfigurator creates a new instance of MetaPipelineConfigurator.
func NewMetaPipelineConfigurator(sourceDir string, context string) MetaPipelineConfigurator {
	return MetaPipelineConfigurator{
		sourceDir: sourceDir,
		context:   context,
	}
}

// ConfigurePipeline configures the Jenkins-X pipeline.
func (e *MetaPipelineConfigurator) ConfigurePipeline() error {
	log.Infof("processing directory '%s'", e.sourceDir)

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

	log.printf("pipelineConfigPath: %s", pipelineConfigPath)

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
	return nil
}

func (e *MetaPipelineConfigurator) insertApplicationStep(projectConfig *config.ProjectConfig) error {
	// insert us into all pipeline kinds for now
	for _, pipelineKind := range jenkinsfile.PipelineKinds {
		pipeline, err := projectConfig.PipelineConfig.Pipelines.GetPipeline(pipelineKind, false)
		if err != nil {
			return errors.Wrapf(err, "unable to retrieve pipeline for type %s", pipelineKind)
		}

		if pipeline == nil {
			continue
		}

		stages := pipeline.Pipeline.Stages

		applicationStep := e.createApplicationStep()

		lastStage := stages[len(stages)-1]
		steps := lastStage.Steps
		steps = append(steps, applicationStep)

		lastStage.Steps = steps
		stages[len(stages)-1] = lastStage
		pipeline.Pipeline.Stages = stages
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
	step := syntax.Step{
		Name:      "sonar-scanner",
		Image:     version.GetFQImage(),
		Command:   "echo 'INSERTED HERE'",
		Arguments: []string{"create"},
	}
	return step
}
