package pipeline

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/jenkins-x-apps/jx-app-sonar-scanner/internal/logging"
	"github.com/jenkins-x-apps/jx-app-sonar-scanner/internal/util"
	"github.com/jenkins-x-apps/jx-app-sonar-scanner/internal/version"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

const (
	userOverridesFile string = ".jx-app-sonar-scanner.yaml"
)

var (
	logger = logging.AppLogger().WithFields(log.Fields{"component": "meta-pipeline-extender"})
)

// Patcher is responsible for injecting a new application step into the pipeline.
type Patcher struct {
	sourceDir     string
	context       string
	sqServer      string
	apiKey        string
	scanonpreview bool
	scanonrelease bool
	debug         bool
}

// UserOverrides represents a user supplied set of UserOverrides values
type UserOverrides struct {
	Verbose     bool      `yaml:"verbose,omitempty"`
	Skip        bool      `yaml:"skip,omitempty"`
	PullRequest BuildStep `yaml:"pullRequest,omitempty"`
	Release     BuildStep `yaml:"release,omitempty"`
}

// BuildStep represents the stage and step after which we should insert the scan
type BuildStep struct {
	Stage string `yaml:"stage,omitempty"`
	Step  string `yaml:"step,omitempty"`
}

// NewPatcher creates a new instance of Patcher.
func NewPatcher(sourceDir string, context string, sqServer string, apiKey string, scanonpreview bool, scanonrelease bool) Patcher {
	return Patcher{
		sourceDir:     sourceDir,
		context:       context,
		sqServer:      sqServer,
		apiKey:        apiKey,
		scanonpreview: scanonpreview,
		scanonrelease: scanonrelease,
		debug:         false,
	}
}

// ConfigurePipeline configures the Jenkins-X pipeline.
func (e *Patcher) ConfigurePipeline() error {
	log.SetFormatter(&log.TextFormatter{
		ForceColors: true,
	})

	logger.Debugf("Processing directory: %s", e.sourceDir)

	if !util.IsDirectory(e.sourceDir) {
		return errors.Errorf("specified directory '%s' does not exist", e.sourceDir)
	}

	effectiveConfig := "jenkins-x-effective.yml"
	if e.context != "" {
		effectiveConfig = fmt.Sprintf("jenkins-x-%s-effective.yml", e.context)
	}

	userOverrides, err := e.getUserOverrides(userOverridesFile)
	if err != nil {
		return errors.Wrap(err, "unable to get user Overrides")
	}

	if userOverrides.Verbose {
		e.debug = true
		log.SetLevel(log.DebugLevel)
	}
	if userOverrides.Skip {
		log.WithFields(log.Fields{
			"sonarscanskip": true,
		}).Warn("Skipping sonar scan due to developer UserOverrides")
		return nil
	}

	pipelineConfigPath := filepath.Join(e.sourceDir, effectiveConfig)
	if !util.Exists(pipelineConfigPath) {
		return errors.Errorf("unable to find effective pipeline config in '%s'", e.sourceDir)
	}

	logger.Debugf("Pipeline config file: %s", pipelineConfigPath)

	content, err := ioutil.ReadFile(pipelineConfigPath)
	if err != nil {
		return errors.Errorf("unable to open pipeline config '%s'", pipelineConfigPath)
	}

	if e.debug {
		dumpInput(content)
	}

	lines := strings.Split(string(content), "\n")
	lines = lines[:len(lines)-1] // trim the additional line introduced by strings.Split
	if len(lines) == 0 {
		return errors.Errorf("empty pipeline")
	}

	if e.scanonpreview {
		lines, err = e.insertApplicationStep(lines, "pullRequest", userOverrides)
		if err != nil {
			return errors.Wrap(err, "unable to enhance preview pipeline with sonar-scanner configuration")
		}
	}

	if e.scanonrelease {
		lines, err = e.insertApplicationStep(lines, "release", userOverrides)
		if err != nil {
			return errors.Wrap(err, "unable to enhance release pipeline with sonar-scanner configuration")
		}
	}

	err = e.writeProjectConfig(lines, pipelineConfigPath)
	if err != nil {
		return errors.Wrap(err, "unable to write modified project config")
	}

	if e.debug {
		dumpOutput(pipelineConfigPath)
	}
	return nil
}

// getUserOverrides returns a set of user defined properties if one exists in the source code
func (e *Patcher) getUserOverrides(file string) (UserOverrides, error) {
	userOverrides := UserOverrides{}
	userOverridesFilePath := filepath.Join(e.sourceDir, file)
	if util.Exists(userOverridesFilePath) {
		// User has set local UserOverrides
		userOverridesFileContent, err := ioutil.ReadFile(userOverridesFilePath)
		if err != nil {
			return userOverrides, errors.Errorf("failed to open '%s'", userOverridesFilePath)
		}

		err = yaml.Unmarshal(userOverridesFileContent, &userOverrides)
		if err != nil {
			return userOverrides, errors.Errorf("unable to parse '%s'", userOverridesFilePath)
		}
	}
	return userOverrides, nil
}

// insertApplicationStep inserts a new step into the pipeline to trigger the scanner
func (e *Patcher) insertApplicationStep(lines []string, pipeline string, userOverrides UserOverrides) ([]string, error) {

	bpm := map[string]map[string]BuildStep{
		"go":                     {"pullRequest": {Stage: "build", Step: "build-make-linux"}, "release": {Stage: "build", Step: "build-make-build"}},
		"gradle":                 {"pullRequest": {Stage: "build", Step: "build-gradle-build"}, "release": {Stage: "build", Step: "build-gradle-build"}},
		"javascript":             {"pullRequest": {Stage: "build", Step: "build-npm-test"}, "release": {Stage: "build", Step: "build-npm-test"}},
		"maven":                  {"pullRequest": {Stage: "build", Step: "build-mvn-install"}, "release": {Stage: "build", Step: "build-mvn-deploy"}},
		"ml-python-gpu-service":  {"pullRequest": {Stage: "build", Step: "build-testing"}, "release": {Stage: "build", Step: "build-testing"}},
		"ml-python-gpu-training": {"pullRequest": {Stage: "build", Step: "testing"}, "release": {Stage: "build", Step: "flake8"}},
		"ml-python-service":      {"pullRequest": {Stage: "build", Step: "build-testing"}, "release": {Stage: "build", Step: "build-testing"}},
		"ml-python-training":     {"pullRequest": {Stage: "build", Step: "build-training"}, "release": {Stage: "build", Step: "build-training"}},
		"python":                 {"pullRequest": {Stage: "build", Step: "build-python-unittest"}, "release": {Stage: "build", Step: "build-python-unittest"}},
		"scala":                  {"pullRequest": {Stage: "build", Step: "build-sbt-assembly"}, "release": {Stage: "build", Step: "build-sbt-assembly"}},
		"typescript":             {"pullRequest": {Stage: "build", Step: "build-npm-test"}, "release": {Stage: "build", Step: "build-npm-test"}},
	}

	buildPack := getBuildPack(lines)
	logger.Infof("Detected buildpack %s\n", buildPack)

	var stagename string
	var stepname string
	if pipeline == "pullRequest" && userOverrides.PullRequest.Stage != "" {
		stagename = userOverrides.PullRequest.Stage
		stepname = userOverrides.PullRequest.Step
		logger.Infof("Overriding %s config\n", pipeline)
	} else if pipeline == "release" && userOverrides.Release.Stage != "" {
		stagename = userOverrides.Release.Stage
		stepname = userOverrides.Release.Step
		logger.Infof("Overriding %s config\n", pipeline)
	} else {
		stagename = bpm[buildPack][pipeline].Stage
		stepname = bpm[buildPack][pipeline].Step
	}
	logger.Debugf("Looking for Stage: %s Step: %s\n", stagename, stepname)

	if buildPack == "" || stagename == "" || stepname == "" {
		// We have found a pipeline that lacks a buildPack that we recognise
		// Fail without breaking the build
		log.Warnf("unable to recognise buildPack: %s\n", buildPack)
		log.Warnf("skipping scan on pipeline: %s [1]\n", pipeline)
		return lines, nil
	}

	logger.Infof("build: %s", pipeline)

	// Identify the subset of this configuration that represents the desired pipeline
	targetPipelineStart, err := indexOfNamedPipeline(lines, pipeline)
	if err != nil {
		return nil, errors.Wrap(err, "finding pipeline")
	}

	pipelineIndent := countLeadingSpace(lines[targetPipelineStart])
	targetPipelineEnd, err := indexOfEndOfPipeline(lines, targetPipelineStart+1, pipelineIndent)
	if err != nil {
		return nil, errors.Wrap(err, "finding end of pipeline")
	}

	logger.Debugf("targetPipelineStart: %d - %s\n", targetPipelineStart, lines[targetPipelineStart])
	logger.Debugf("targetPipelineEnd: %d - %s\n", targetPipelineEnd, lines[targetPipelineEnd])
	logger.Debugf("size: %d\n", len(lines))

	targetPipeline := lines[targetPipelineStart : targetPipelineEnd+1] // This creates an offset that we need to account for later

	// Identify the point where we should insert environment variables
	createEnv := false
	envInsertPoint, err := indexOfEnv(targetPipeline)
	if err != nil {
		// No env section in containerOptions: so insert in pipelineConfig:
		pipelineConfigStart, err2 := indexOfPipelineConfig(lines)
		if err2 != nil {
			return nil, errors.Wrap(err, "finding pipelineConfig")
		}
		logger.Debugf("pipelineConfigStart: %d - %s\n", pipelineConfigStart, lines[pipelineConfigStart])
		pipelinesStart, err3 := indexOfPipelines(lines)
		if err3 != nil {
			return nil, errors.Wrap(err, "finding pipelines")
		}
		logger.Debugf("pipelinesStart: %d - %s\n", pipelinesStart, lines[pipelinesStart])
		indexEnv, err4 := indexOfEnv(lines[:pipelinesStart])
		if err4 != nil {
			// No existing env: section in pipelineConfig either
			envInsertPoint = pipelineConfigStart + 1
			createEnv = true
		} else {
			envInsertPoint = indexEnv + 1
		}
	} else {
		envInsertPoint = targetPipelineStart + envInsertPoint + 1
	}

	// Identify the subset of this configuration that represents the desired stage
	targetStagesStart, err := indexOfStages(lines[targetPipelineStart : targetPipelineEnd+1])
	if err != nil {
		return nil, errors.Wrap(err, "finding stages:")
	}
	targetStagesStart = targetStagesStart + targetPipelineStart
	logger.Debugf("targetStagesStart: %d - %s\n", targetStagesStart, lines[targetStagesStart])
	somewhereInTargetStage, err := indexOfNamedStage(lines[targetStagesStart:targetPipelineEnd+1], stagename)
	if err != nil {
		// We have found a pipeline that lacks a build stage that we recognise
		// Fail without breaking the build
		logger.Debugf("unable to find named Stage: %s\n", stagename)
		logger.Debugf("skipping scan on pipeline: %s [2]\n", pipeline)
		return lines, nil
	}
	somewhereInTargetStage = somewhereInTargetStage + targetStagesStart // realign to absolute offset

	// scan from next line after name: until we find the start of the next steps: object
	currentStage, err := indexOfCurrentStage(lines, somewhereInTargetStage)
	if err != nil {
		return nil, errors.Wrap(err, "unable to find start of stage")
	}
	logger.Debugf("currentStage: %d - %s\n", currentStage, lines[currentStage])
	targetStepsStart, err := indexOfNextSteps(lines[currentStage : targetPipelineEnd+1])
	if err != nil {
		return nil, errors.Wrap(err, "unable to find steps:")
	}
	targetStepsStart = targetStepsStart + currentStage // realign to absolute offset
	logger.Debugf("targetStepsStart: %d - %s\n", targetStepsStart, lines[targetStepsStart])

	// find end of steps: section
	targetStepsEnd, err := indexOfEndOfSteps(lines[:targetPipelineEnd+1], targetStepsStart, countLeadingSpace(lines[targetStepsStart]))
	if err != nil {
		return nil, errors.Wrap(err, "unable to find end of steps:")
	}
	logger.Debugf("targetStepsEnd: %d - %s\n", targetStepsEnd, lines[targetStepsEnd])

	targetSteps := lines[targetStepsStart : targetStepsEnd+1] // This creates an offset that we need to account for later

	// Identify the line relating to the step we wish to insert after

	somewhereInTargetStep, err := indexOfNamedStep(targetSteps, stepname)
	if err != nil {
		// We have found a pipeline that lacks a build step that we recognise
		// Fail without breaking the build
		logger.Debugf("unable to find named Step: %s\n", stepname)
		logger.Debugf("skipping scan on pipeline: %s [3]\n", pipeline)
		return lines, nil
	}
	somewhereInTargetStep = somewhereInTargetStep + targetStepsStart // realign to absolute offset

	logger.Debugf("somewhereInTargetStep: %d - %s\n", somewhereInTargetStep, lines[somewhereInTargetStep])

	// scan from next line after name: until we find the start of the next step
	nextStep, err := indexOfNextStep(lines[somewhereInTargetStep+1 : targetPipelineEnd+1])
	if err != nil {
		return nil, errors.Wrap(err, "unable to find next step")
	}
	nextStep = nextStep + somewhereInTargetStep // realign to absolute offset

	currentStep, err := indexOfCurrentStep(lines, somewhereInTargetStep)
	if err != nil {
		return nil, errors.Wrap(err, "unable to find start of step")
	}
	stepIndent := countLeadingSpace(lines[currentStep])
	envIndent := countLeadingSpace(lines[envInsertPoint])
	logger.Debugf("nextStep: %d - %s\n", nextStep, lines[nextStep])

	// resolve the offsets
	absoluteInsertPoint := nextStep + 1

	logger.Debugf("absoluteInsertPoint: %d\n", absoluteInsertPoint)

	applicationStep := e.createApplicationStep(stepIndent)

	lines = append(lines, applicationStep...)                                           // make the slice bigger by the size of the new step
	copy(lines[absoluteInsertPoint+len(applicationStep):], lines[absoluteInsertPoint:]) // move the subsequent lines down
	copy(lines[absoluteInsertPoint:], applicationStep)                                  // insert the new step

	if !envExists(lines, envInsertPoint) {
		envEntry := e.createEnvEntry(envIndent, buildPack, createEnv)
		lines = append(lines, envEntry...)                                 // make the slice bigger by the size of the environment variables
		copy(lines[envInsertPoint+len(envEntry):], lines[envInsertPoint:]) // move the subsequent lines down
		copy(lines[envInsertPoint:], envEntry)                             // insert the new step
	}
	return lines, nil
}

func (e *Patcher) writeProjectConfig(lines []string, pipelineConfigPath string) error {
	err := util.MoveFile(pipelineConfigPath, pipelineConfigPath+".sonar-scanner.orig")
	if err != nil {
		return errors.Wrapf(err, "unable to backup '%s'", pipelineConfigPath)
	}

	logger.Debugf("writing '%s'", pipelineConfigPath)

	file, err := os.OpenFile(pipelineConfigPath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return errors.Wrapf(err, "unable to create new pipeline file '%s'", pipelineConfigPath)
	}
	defer file.Close()
	sep := "\n"
	for _, line := range lines {
		if _, err = file.WriteString(line + sep); err != nil {
			return errors.Wrapf(err, "unable to write '%s'", pipelineConfigPath)
		}
	}
	return nil
}

func (e *Patcher) createApplicationStep(indent int) []string {
	// set correct whitespace for indent
	ws := nspaces(indent)

	// build the set of arguments for the script
	args := []string{}
	if e.sqServer != "" {
		args = append(args, "-s "+e.sqServer)
	}
	if e.apiKey != "" {
		args = append(args, "-k "+e.apiKey)
	}
	args = append(args, "-r "+strconv.FormatBool(e.scanonrelease))
	args = append(args, "-p "+strconv.FormatBool(e.scanonpreview))
	if e.debug {
		args = append(args, "-v "+strconv.FormatBool(e.debug))
	}

	// construct the pipeline syntax for the step
	step := []string{}
	step = append(step, ws+"- command: /usr/local/bin/exec-sonar-scanner.sh")
	step = append(step, ws+"  args:")
	for _, arg := range args {
		step = append(step, ws+"  - "+arg)
	}
	step = append(step, ws+"  image: "+version.GetFQImage())
	step = append(step, ws+"  name: sonar-scanner")
	return step
}

func (e *Patcher) createEnvEntry(indent int, buildPack string, create bool) []string {
	// set correct whitespace for indent
	ws := nspaces(indent)

	// construct the pipeline syntax for the environment variable
	env := []string{}
	if create {
		env = append(env, ws+"env:")
	}
	env = append(env, ws+"- name: BUILDPACK_NAME")
	env = append(env, ws+"  value: "+buildPack)
	return env
}

func nspaces(n int) string {
	s := make([]byte, n)
	for i := 0; i < n; i++ {
		s[i] = ' '
	}
	return string(s)
}

// indexOfNamedStage finds the first instance of a named stage with the given pipeline
func indexOfNamedStage(lines []string, name string) (int, error) {
	stageIndent := countLeadingSpace(lines[0])
	for l, line := range lines {
		isTopLevel, err := hasMatchingIndent(line, stageIndent)
		if err != nil {
			return 0, errors.Wrapf(err, "parsing match: '%s'", line)
		}
		isFirstIndent, err := hasMatchingIndent(line, stageIndent+2)
		if err != nil {
			return 0, errors.Wrapf(err, "parsing match: '%s'", line)
		}
		if isTopLevel || isFirstIndent {
			if isNamedStage(line, name) {
				return l, nil
			}
		}
	}
	return 0, errors.Errorf("unable to find stage '%s'", name)
}

// indexOfNamedStep finds the first instance of a named step with the given stage
func indexOfNamedStep(lines []string, stepname string) (int, error) {
	for l, line := range lines {
		if strings.Contains(line, stepname) && strings.Contains(line, "name:") {
			return l, nil
		}
	}
	return 0, errors.Errorf("unable to find step '%s'", stepname)
}

// indexOfNamedPipeline finds the first instance of a named pipeline
func indexOfNamedPipeline(lines []string, pipeline string) (int, error) {
	for l, line := range lines {
		if strings.Contains(line, pipeline+":") {
			return l, nil
		}
	}
	return 0, errors.Errorf("unable to find pipeline '%s'", pipeline)
}

// isRootOfSection indicates whether a given string s is the root of a section
func isRootOfSection(s string) (bool, error) {
	exp := `^\s+-\s\S+:`
	return regexp.MatchString(exp, s)
}

// indexOfString finds the first instance of the given string in a slice
func indexOfString(lines []string, s string) (int, error) {
	for l, line := range lines {
		if strings.Contains(line, s) {
			return l, nil
		}
	}
	return 0, errors.Errorf("unable to find '%s'", s)
}

// indexOfNextStep returns the index of the first line of a new steps: object
func indexOfNextStep(lines []string) (int, error) {
	for l, line := range lines {
		match, err := isRootOfSection(line)
		if err != nil {
			return 0, errors.Wrapf(err, "finding next Step: '%s'", line)
		}
		if match {
			return l, nil
		}
	}
	return len(lines), nil
}

// indexOfNextSteps returns the index of the first line of a new step or the end of the list
func indexOfNextSteps(lines []string) (int, error) {
	for l, line := range lines {
		if strings.Contains(line, "steps:") {
			return l, nil
		}
	}
	return 0, errors.Errorf("unable to find steps:")
}

// indexOfCurrentSection returns the index of the first line of the section containing this index
func indexOfCurrentSection(lines []string, start int) (int, error) {
	for l := start; l > 0; l-- {
		match, err := isRootOfSection(lines[l])
		if err != nil {
			return 0, errors.Wrapf(err, "checking root status of line: '%s'", lines[l])
		}
		if match {
			return l, nil
		}
	}
	return 0, errors.Errorf("unable to find start of step containing '%d'", start)
}

// indexOfCurrentStage returns the index of the first line of the stage containing this index
func indexOfCurrentStage(lines []string, start int) (int, error) {
	return indexOfCurrentSection(lines, start)
}

// indexOfCurrentStep returns the index of the first line of the step containing this index
func indexOfCurrentStep(lines []string, start int) (int, error) {
	return indexOfCurrentSection(lines, start)
}

// hasMatchingIndent indicates whether a given string s has an initial indent of length i
func hasMatchingIndent(s string, i int) (bool, error) {
	exp := `^\s{` + strconv.Itoa(i) + `}\S`
	return regexp.MatchString(exp, s)
}

// indexOfEndOfPipeline finds the index of the last entry in the current pipeline, or the end of the file if no subsequent pipeline is defined
func indexOfEndOfPipeline(lines []string, start int, indent int) (int, error) {
	if start >= len(lines) {
		return 0, errors.Errorf("start value too big '%d' [1]", start)
	}
	for i := start; i < len(lines); i++ {
		match, err := hasMatchingIndent(lines[i], indent)
		if err != nil {
			return 0, errors.Wrapf(err, "whilst parsing line: '%s'", lines[i])
		}
		if match {
			return i - 1, nil // report the index of the previous line
		}
	}
	// end of file
	return len(lines) - 2, nil // account for last line being blank
}

// indexOfEndOfSteps finds the index of the last entry in the current steps:, or the end of the list
func indexOfEndOfSteps(lines []string, start int, indent int) (int, error) {
	if start >= len(lines) {
		return 0, errors.Errorf("start value too big '%d' [2]", start)
	}
	for i := start; i < len(lines); i++ {
		thisindent := countLeadingSpace(lines[i])
		if thisindent < indent {
			return i - 1, nil // report the index of the previous line
		}
	}
	// end of file
	return len(lines) - 1, nil
}

// indexOfEndOfStage finds the index of the last entry in the current stage, or the end of the list if no subsequent stage is defined
func indexOfEndOfStage(lines []string, start int, indent int) (int, error) {
	if start >= len(lines) {
		return 0, errors.Errorf("start value too big '%d' [3]", start)
	}
	for i := start; i < len(lines); i++ {
		match, err := hasMatchingIndent(lines[i], indent)
		if err != nil {
			return 0, errors.Wrapf(err, "whilst parsing line: '%s'", lines[i])
		}
		if match {
			return i - 1, nil // report the index of the previous line
		}
	}
	// end of file
	return len(lines) - 1, nil
}

// countLeadingSpace measures the indent of the given string
func countLeadingSpace(line string) int {
	i := 0
	for _, runeValue := range line {
		if runeValue == ' ' {
			i++
		} else {
			break
		}
	}
	return i
}

func getBuildPack(lines []string) string {
	exp := `^buildPack: (\S+)`
	re := regexp.MustCompile(exp)
	r := re.FindStringSubmatch(lines[0])
	if len(r) == 2 {
		return r[1]
	}
	return ""
}

// indexOfEnv finds the first instance of an env: entry
func indexOfEnv(lines []string) (int, error) {
	return indexOfString(lines, "env:")
}

// indexOfPipelineConfig finds the start of the pipelineConfig: entry
func indexOfPipelineConfig(lines []string) (int, error) {
	return indexOfString(lines, "pipelineConfig:")
}

// indexOfPipelines finds the start of the pipelines: entry
func indexOfPipelines(lines []string) (int, error) {
	return indexOfString(lines, "pipelines:")
}

// indexOfPipeline finds the start of the pipeline: entry
func indexOfPipeline(lines []string) (int, error) {
	return indexOfString(lines, "pipeline:")
}

// indexOfPipeline finds the start of the pipeline: entry
func indexOfStages(lines []string) (int, error) {
	return indexOfString(lines, "stages:")
}

// envExists checks if the buildpack env variable is already set
func envExists(lines []string, index int) bool {
	return strings.Contains(lines[index], "name:") && strings.Contains(lines[index], "BUILDPACK_NAME")
}

// isNamedStage checks if this line contains the name: declaration for stage 'name'
func isNamedStage(line string, name string) bool {
	return strings.Contains(line, "name:") && strings.Contains(line, name)
}

// dumpInput writes pipeline to log to check input format
func dumpInput(content []byte) {
	// Dump pipeline to log to check input format
	fmt.Println("---------------------------INPUT PIPELINE---------------------------")
	fmt.Println(string(content))
	fmt.Println("--------------------------------------------------------------------")
}

// dumpOutput writes pipeline to log to check output format
func dumpOutput(path string) {
	fmt.Println("--------------------------OUTPUT PIPELINE---------------------------")
	content, err := ioutil.ReadFile(path)
	if err != nil {
		logger.Fatalf("unable to display pipeline config '%s'", path)
	}
	fmt.Println(string(content))
	fmt.Println("--------------------------------------------------------------------")
}
