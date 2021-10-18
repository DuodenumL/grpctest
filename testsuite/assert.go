package testsuite

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/stretchr/testify/assert"
)

// Equal assertion
type Equal struct {
	Actual   string `json:"actual"`
	Expected string `json:"expected"`
}

// Assertion .
type Assertion struct {
	ForEach struct {
		Equals     []Equal  `json:"equals"`
		RunSuccess []string `json:"run_success"`
	} `json:"for_each"`
	AfterCompletion struct {
		Equals     []Equal  `json:"equals"`
		RunSuccess []string `json:"run_success"`
	} `json:"after_completion"`
}

// MustNewAssertion .
func MustNewAssertion(a string) *Assertion {
	asser := &Assertion{}
	if err := json.Unmarshal([]byte(a), asser); err != nil {
		log.Fatalf("failed to load assertion: %+v", err)
	}
	return asser
}

func (a Assertion) assertEach(t assert.TestingT, req, resp, err string) bool {
	env := []string{
		"req=" + req,
		"resp=" + resp,
		"err=" + err,
	}
	for _, equal := range a.ForEach.Equals {
		if !a.equal(t, equal, env) {
			return false
		}
	}
	for _, command := range a.ForEach.RunSuccess {
		if !a.runSuccess(t, command, env) {
			return false
		}
	}
	return true
}

func (a Assertion) assertCompletion(t assert.TestingT, req string, resps, errs []string) bool {
	env := []string{
		"req=" + req,
		"resps=" + strings.Join(resps, "\n"),
		"errs=" + strings.Join(errs, "\n"),
	}
	for _, equal := range a.AfterCompletion.Equals {
		if !a.equal(t, equal, env) {
			return false
		}
	}
	for _, command := range a.AfterCompletion.RunSuccess {
		if !a.runSuccess(t, command, env) {
			return false
		}
	}
	// TODO@zc: too many duplicated code between assertEach and assertCompletion
	return true
}

func (a Assertion) equal(t assert.TestingT, equal Equal, env []string) bool {
	envString := strings.Join(env, " ")
	env = append(os.Environ(), env...)
	actualOut, e := bash(equal.Actual, env)
	if e != nil {
		log.Fatalf("failed to exec bash command: err %+v, output %s, %s %s", e, actualOut, envString, equal.Actual)
	}
	expectedOut, e := bash(equal.Expected, env)
	if e != nil {
		log.Fatalf("failed to exec bash command: err %+v, output %s, %s %s", e, expectedOut, envString, equal.Expected)
	}
	return assert.EqualValues(t, expectedOut, actualOut, fmt.Sprintf("%s %s", envString, equal.Actual))
}

func (a Assertion) runSuccess(t assert.TestingT, command string, env []string) bool {
	output, e := bash(command, append(os.Environ(), env...))
	return assert.NoError(t, e, fmt.Sprintf("output: %s, command: %s", output, command))
}
