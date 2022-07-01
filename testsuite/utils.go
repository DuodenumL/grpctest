package testsuite

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const pyTemplate = `
import os
import sys
import json

def env(key):
	return os.getenv(key, "")
`

func bash(command string, env []string) (out string, err error) {
	cwd, err := os.Getwd()
	if err != nil {
		return
	}
	cdCwd := fmt.Sprintf("cd %s; ", filepath.Dir(filepath.Dir(cwd)))
	cmd := exec.Command("/bin/bash", "-c", "set -eo pipefail; "+cdCwd+command) // nolint:gosec
	cmd.Env = env
	output, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(output)), err
}

func py(command string, env []string) (out string, err error) {
	python := os.Getenv("PYTHON")
	if python == "" {
		python = "python"
	}
	cmd := exec.Command(python, "-c", pyTemplate+command)
	cmd.Env = env
	output, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(output)), err
}

func execCommand(command string, env []string) (out string, err error) {
	if len(command) == 0 {
		return "", nil
	}
	if strings.HasPrefix(command, "py:") {
		return py(strings.TrimPrefix(strings.TrimPrefix(command, "py:"), " "), env)
	}
	return bash(command, env)
}

func combine(candicates [][]string) (res [][]string) {
	var do func(int, []string)
	do = func(idx int, wip []string) {
		if idx == len(candicates) {
			cp := make([]string, len(wip))
			copy(cp, wip)
			res = append(res, cp)
			return
		}

		wip = append(wip, "")
		for _, s := range candicates[idx] {
			wip[idx] = s
			do(idx+1, wip)
		}
	}

	do(0, []string{})
	return
}
