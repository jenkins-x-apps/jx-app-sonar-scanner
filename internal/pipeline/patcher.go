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
	}
}

// ConfigurePipeline configures the Jenkins-X pipeline.
func (e *Patcher) ConfigurePipeline() error {
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
		return errors.Errorf("unable to open pipeline config '%s'", pipelineConfigPath)
	}
	fmt.Println(string(content))
	fmt.Println("--------------------------------------------------------------------")

	lines := strings.Split(string(content), "\n")
	lines = lines[:len(lines)-1] // trim the additional line introduced by strings.Split
	if len(lines) == 0 {
		return errors.Errorf("empty pipeline")
	}

	if e.scanonpreview {
		lines, err = e.insertApplicationStep(lines, "pullRequest")
		if err != nil {
			return errors.Wrap(err, "unable to enhance preview pipeline with sonar-scanner configuration")
		}
	}

	if e.scanonrelease {
		lines, err = e.insertApplicationStep(lines, "release")
		if err != nil {
			return errors.Wrap(err, "unable to enhance release pipeline with sonar-scanner configuration")
		}
	}

	err = e.writeProjectConfig(lines, pipelineConfigPath)
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

func (e *Patcher) insertApplicationStep(lines []string, pipeline string) ([]string, error) {

	bpm := map[string]map[string]string{
		"go":     {"pullRequest": "build-make-linux", "release": "build-make-build"},
		"maven":  {"pullRequest": "build-mvn-install", "release": "build-mvn-deploy"},
		"python": {"pullRequest": "build-python-unittest", "release": "build-python-unittest"},
	}

	builder := getBuildPack(lines)
	stepname := bpm[builder][pipeline]

	if builder == "" || stepname == "" {
		// We have found a pipeline that lacks a builder that we recognise
		// Fail without breaking the build
		fmt.Printf("unable to recognise builder: %s\n", stepname)
		fmt.Printf("skipping scan on pipeline: %s\n", pipeline)
		return lines, nil
	}

	log.WithFields(log.Fields{
		"pipelineKind": pipeline,
	}).Info("pipeline")

	targetPipelineStart, err := indexOfString(lines, pipeline+":")
	if err != nil {
		return nil, errors.Wrap(err, "finding pipeline")
	}

	pipelineIndent := countLeadingSpace(lines[targetPipelineStart])
	targetPipelineEnd, err := indexOfEndOfPipeline(lines, targetPipelineStart+1, pipelineIndent)
	if err != nil {
		return nil, errors.Wrap(err, "finding  end of pipeline")
	}

	fmt.Printf("targetPipelineStart: %d\n", targetPipelineStart)
	fmt.Printf("targetPipelineEnd: %d\n", targetPipelineEnd)
	fmt.Printf("size: %d\n", len(lines))

	targetPipeline := lines[targetPipelineStart:targetPipelineEnd] // This creates an offset that we need to account for later

	somewhereInTargetStep, err := indexOfNamedStep(targetPipeline, stepname)
	if err != nil {
		// We have found a pipeline that lacks a build step that we recognise
		// Fail without breaking the build
		fmt.Printf("unable to find named step: %s\n", stepname)
		fmt.Printf("skipping scan on pipeline: %s\n", targetPipeline)
		return lines, nil
	}
	somewhereInTargetStep = somewhereInTargetStep + targetPipelineStart // realign to absolute offset

	fmt.Printf("somewhereInTargetStep: %d\n", somewhereInTargetStep)

	// scan from next line after name: until we find the start of the next step
	nextStep, err := indexOfNextStep(lines[somewhereInTargetStep+1 : targetPipelineEnd])
	if err != nil {
		return nil, errors.Wrap(err, "unable to find next step")
	}
	nextStep = nextStep + somewhereInTargetStep // realign to absolute offset

	currentStep, err := indexOfCurrentStep(lines, somewhereInTargetStep)
	if err != nil {
		return nil, errors.Wrap(err, "unable to find start of step")
	}
	stepIndent := countLeadingSpace(lines[currentStep])

	fmt.Printf("nextStep: %d\n", nextStep)

	// resolve the offsets
	absoluteInsertPoint := nextStep + 1

	fmt.Printf("absoluteInsertPoint: %d\n", absoluteInsertPoint)

	applicationStep := e.createApplicationStep(stepIndent)

	lines = append(lines, applicationStep...)                                           // make the slice bigger by the size of the new step
	copy(lines[absoluteInsertPoint+len(applicationStep):], lines[absoluteInsertPoint:]) // move the subsequent lines down
	copy(lines[absoluteInsertPoint:], applicationStep)                                  // insert the new step

	return lines, nil
}

func (e *Patcher) writeProjectConfig(lines []string, pipelineConfigPath string) error {
	err := util.MoveFile(pipelineConfigPath, pipelineConfigPath+".sonar-scanner.orig")
	if err != nil {
		return errors.Wrapf(err, "unable to backup '%s'", pipelineConfigPath)
	}

	logger.Infof("writing '%s'", pipelineConfigPath)

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

func nspaces(n int) string {
	s := make([]byte, n)
	for i := 0; i < n; i++ {
		s[i] = ' '
	}
	return string(s)
}

// indexOfNamedStep finds the first instance of a named step with the given name
func indexOfNamedStep(lines []string, stepname string) (int, error) {
	for l, line := range lines {
		if strings.Contains(line, stepname) && strings.Contains(line, "name:") {
			return l, nil
		}
	}
	return 0, errors.Errorf("unable to find step '%s'", stepname)
}

// hasMatchingIndent indicates whether a given string s has an initial indent of length i
func isRootOfStep(s string) (bool, error) {
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
	return 0, errors.Errorf("unable to find step '%s'", s)
}

// indexOfNextStep returns the index of the first line of a new step or the end of the list
func indexOfNextStep(lines []string) (int, error) {
	for l, line := range lines {
		match, err := isRootOfStep(line)
		if err != nil {
			return 0, errors.Wrapf(err, "finding next step: '%s'", line)
		}
		if match {
			return l, nil
		}
	}
	return len(lines) - 1, nil
}

// indexOfCurrentStep returns the index of the first line of the step containing this index
func indexOfCurrentStep(lines []string, start int) (int, error) {
	for l := start; l > 0; l-- {
		match, err := isRootOfStep(lines[l])
		if err != nil {
			return 0, errors.Wrapf(err, "checking root status of line: '%s'", lines[l])
		}
		if match {
			return l, nil
		}
	}
	return 0, errors.Errorf("unable to find start of step containing '%d'", start)
}

// hasMatchingIndent indicates whether a given string s has an initial indent of length i
func hasMatchingIndent(s string, i int) (bool, error) {
	exp := `^\s{` + strconv.Itoa(i) + `}\S`
	return regexp.MatchString(exp, s)
}

// indexOfEndOfPipeline finds the index of the last entry in the current pipeline, or the end of the file if no subsequent pipeline is defined
func indexOfEndOfPipeline(lines []string, start int, indent int) (int, error) {
	if start >= len(lines) {
		return 0, errors.Errorf("start value too big '%d'", start)
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
