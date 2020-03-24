package pipeline

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	// "github.com/jenkins-x-apps/jx-app-sonar-scanner/internal/util"
	"github.com/jenkins-x/jx/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/udhos/equalfile"
)

func TestPatcher_ConfigurePipeline(t *testing.T) {
	t.Parallel()

	type fields struct {
		context       string
		sqServer      string
		apiKey        string
		scanonpreview bool
		scanonrelease bool
	}

	// Test data
	testDataLocation := "../../test/"

	// Create a temporary directory for testing and ensure it is cleaned up after
	tmpDirs := make([]string, 0)
	defer func() {
		for _, dir := range tmpDirs {
			err := os.RemoveAll(dir)
			assert.NoError(t, err)
		}
	}()

	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"go", fields{"", "http://jx-sonarqube.sonarqube.svc.cluster.local:9000", "12345", true, true}, false},
		{"go-preview", fields{"", "http://jx-sonarqube.sonarqube.svc.cluster.local:9000", "12345", true, false}, false},
		{"go-release", fields{"", "http://jx-sonarqube.sonarqube.svc.cluster.local:9000", "12345", false, true}, false},
		{"go-none", fields{"", "http://jx-sonarqube.sonarqube.svc.cluster.local:9000", "12345", false, false}, false},
		{"go-no-token", fields{"", "http://jx-sonarqube.sonarqube.svc.cluster.local:9000", "", true, true}, false},
		{"go-no-server", fields{"", "", "12345", true, true}, false},
		{"maven", fields{"", "http://jx-sonarqube.sonarqube.svc.cluster.local:9000", "12345", true, true}, false},
		{"python", fields{"", "http://jx-sonarqube.sonarqube.svc.cluster.local:9000", "12345", true, true}, false},
		{"unknown-step-name", fields{"", "http://jx-sonarqube.sonarqube.svc.cluster.local:9000", "12345", true, true}, false},
		{"unknown-builder", fields{"", "http://jx-sonarqube.sonarqube.svc.cluster.local:9000", "12345", true, true}, false},
	}

	cmp := equalfile.New(nil, equalfile.Options{})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, err := ioutil.TempDir(testDataLocation, "run-"+tt.name)
			tmpDirs = append(tmpDirs, dir)
			assert.NoError(t, err)
			dataTemplateLocation := filepath.Join(testDataLocation, tt.name)
			err = util.CopyDir(dataTemplateLocation, dir, true)
			assert.NoError(t, err)

			e := &Patcher{
				sourceDir:     dir,
				context:       tt.fields.context,
				sqServer:      tt.fields.sqServer,
				apiKey:        tt.fields.apiKey,
				scanonpreview: tt.fields.scanonpreview,
				scanonrelease: tt.fields.scanonrelease,
			}
			if err := e.ConfigurePipeline(); (err != nil) != tt.wantErr {
				t.Errorf("Patcher.ConfigurePipeline() error = %v, wantErr %v", err, tt.wantErr)
			}
			equal, err := cmp.CompareFile(filepath.Join(dir, "jenkins-x-effective.yml"), filepath.Join(dir, "jenkins-x-effective.gold.yml"))
			assert.NoError(t, err)
			assert.Equal(t, equal, true, "pipeline files don't match")
		})
	}
}

func Test_indexOfEndOfPipeline(t *testing.T) {

	// Test data
	testDataLocation := "../../test/"

	type args struct {
		start  int
		indent int
	}
	tests := []struct {
		name    string
		file    string
		args    args
		want    int
		wantErr bool
	}{
		{"indexEndOfPipeline1", "indexEndOfPipeline1.yml", args{11, 4}, 86, false}, // Zero-referenced, start AFTER beginnning of this pipeline
		{"indexEndOfPipeline2", "indexEndOfPipeline1.yml", args{88, 4}, 182, false},
		{"indexEndOfPipeline2", "indexEndOfPipeline2.yml", args{8, 4}, 71, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dataTemplateLocation := filepath.Join(testDataLocation, tt.file)
			content, err := ioutil.ReadFile(dataTemplateLocation)
			assert.NoError(t, err)
			lines := strings.Split(string(content), "\n")
			got, err := indexOfEndOfPipeline(lines, tt.args.start, tt.args.indent)
			if (err != nil) != tt.wantErr {
				t.Errorf("indexOfEndOfPipeline() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("indexOfEndOfPipeline() = %v, want %v", got, tt.want)
			}
		})
	}
}
