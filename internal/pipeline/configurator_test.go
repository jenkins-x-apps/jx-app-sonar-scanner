package pipeline

import (
	"github.com/jenkins-x/jx/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/udhos/equalfile"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestMetaPipelineConfigurator_ConfigurePipeline(t *testing.T) {
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

			e := &MetaPipelineConfigurator{
				sourceDir:     dir,
				context:       tt.fields.context,
				sqServer:      tt.fields.sqServer,
				apiKey:        tt.fields.apiKey,
				scanonpreview: tt.fields.scanonpreview,
				scanonrelease: tt.fields.scanonrelease,
			}
			if err := e.ConfigurePipeline(); (err != nil) != tt.wantErr {
				t.Errorf("MetaPipelineConfigurator.ConfigurePipeline() error = %v, wantErr %v", err, tt.wantErr)
			}
			equal, err := cmp.CompareFile(filepath.Join(dir, "jenkins-x-effective.yml"), filepath.Join(dir, "jenkins-x-effective.gold.yml"))
			assert.NoError(t, err)
			assert.Equal(t, equal, true, "pipeline files don't match")
		})
	}
}
