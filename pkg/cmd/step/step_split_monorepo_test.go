package step_test

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	step2 "github.com/jenkins-x/jx/pkg/cmd/opts/step"

	"github.com/jenkins-x/jx/pkg/cmd/step"

	"github.com/jenkins-x/jx/pkg/cmd/opts"
	"github.com/jenkins-x/jx/pkg/log"
	"github.com/jenkins-x/jx/pkg/tests"
	"github.com/stretchr/testify/assert"
)

func TestStepSplitMonorepo(t *testing.T) {
	t.Parallel()
	testData := filepath.Join("test_data", "split_monorepo")

	tempDir, err := ioutil.TempDir("", "test_split_monorepo")
	assert.NoError(t, err)

	options := &step.StepSplitMonorepoOptions{
		StepOptions: step2.StepOptions{
			CommonOptions: &opts.CommonOptions{},
		},
		Organisation: "dummy",
		Glob:         "*",
		Dir:          testData,
		OutputDir:    tempDir,
		NoGit:        true,
	}

	err = options.Run()
	assert.NoError(t, err, "Failed to run split monorepo on source %s output %s", testData, tempDir)

	tests.Debugf("Generated split repos in: %s\n", tempDir)
	log.Logger().Infof("Generated split repos in: %s", tempDir)

	tests.AssertFilesExist(t, true, filepath.Join(tempDir, "bar"), filepath.Join(tempDir, "foo"))
	tests.AssertFilesExist(t, false, filepath.Join(tempDir, "kubernetes"))

	tests.AssertFilesExist(t, true,
		filepath.Join(tempDir, "bar", "charts", "bar", "Chart.yaml"),
		filepath.Join(tempDir, "bar", "charts", "bar", "templates", "deployment.yaml"))
}

func TestStepSplitMonorepoGetLastGitCommit(t *testing.T) {
	t.Parallel()
	testData := filepath.Join("test_data", "split_monorepo")

	tempDir, err := ioutil.TempDir("", "test_split_monorepo")
	assert.NoError(t, err)

	options := &step.StepSplitMonorepoOptions{
		StepOptions: step2.StepOptions{
			CommonOptions: &opts.CommonOptions{},
		},
		Organisation: "dummy",
		Glob:         "*",
		Dir:          testData,
		OutputDir:    tempDir,
		RepoName:     "test",
		NoGit:        true,
	}

	err = options.Run()
	assert.NoError(t, err, "Failed to run split monorepo on source %s output %s", testData, tempDir)

	tests.Debugf("Generated split repos in: %s\n", tempDir)
	log.Logger().Infof("Generated split repos in: %s", tempDir)

	tests.AssertFilesExist(t, true, filepath.Join(tempDir, "bar"), filepath.Join(tempDir, "foo"))
	tests.AssertFilesExist(t, false, filepath.Join(tempDir, "kubernetes"))

	tests.AssertFilesExist(t, true,
		filepath.Join(tempDir, "bar", "charts", "bar", "Chart.yaml"),
		filepath.Join(tempDir, "bar", "charts", "bar", "templates", "deployment.yaml"))
}
