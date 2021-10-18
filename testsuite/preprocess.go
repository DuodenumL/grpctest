package testsuite

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pkg/errors"
)

func Preprocess(suiteYaml string) (stdout, stderr io.Reader, err error) {
	python := os.Getenv("PYTHON")
	if python == "" {
		python = "python"
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}
	pythonScript := filepath.Join(cwd, "testsuite", "gen.py")
	cmd := exec.Command(python, pythonScript, suiteYaml)
	sout, serr := &bytes.Buffer{}, &bytes.Buffer{}
	cmd.Stdout, cmd.Stderr = sout, serr
	return sout, serr, errors.WithStack(cmd.Run())
}
