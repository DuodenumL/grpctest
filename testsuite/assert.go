package testsuite

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
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
		Run        []string `json:"run"`
	} `json:"after_completion"`
}

// MustNewAssertion .
func MustNewAssertion(a string) *Assertion {
	asser := &Assertion{}
	if err := json.Unmarshal([]byte(a), asser); err != nil {
		log.Fatalf("failed to load assertion: %v, err: %+v", a, err)
	}
	return asser
}

func (a Assertion) assertEach(t assert.TestingT, req, resp, err string) error {
	env := []string{
		"req=" + req,
		"resp=" + resp,
		"err=" + err,
	}
	for _, equal := range a.ForEach.Equals {
		if err := a.equal(t, equal, env); err != nil {
			return errors.Wrap(err, "failed to assert equal in assert each")
		}
	}
	for _, command := range a.ForEach.RunSuccess {
		if err := a.runSuccess(t, command, env); err != nil {
			return errors.Wrap(err, "failed to exec run success in assert each")
		}
	}
	return nil
}

func (a Assertion) assertCompletion(t assert.TestingT, req string, resps, errs []string) error {
	env := []string{
		"req=" + req,
		"resps=" + strings.Join(resps, "\n"),
		"errs=" + strings.Join(errs, "\n"),
	}
	defer func() {
		for _, command := range a.AfterCompletion.Run {
			output, err := execCommand(command, append(os.Environ(), env...))
			log.Infof("command: %s, output: %s, err: %+v", command, output, err)
		}
	}()
	for _, equal := range a.AfterCompletion.Equals {
		if err := a.equal(t, equal, env); err != nil {
			return errors.Wrap(err, "failed to assert equal in assert completion")
		}
	}
	for _, command := range a.AfterCompletion.RunSuccess {
		if err := a.runSuccess(t, command, env); err != nil {
			return errors.Wrap(err, "failed to exec run success in assert completion")
		}
	}
	// TODO@zc: too many duplicated code between assertEach and assertCompletion
	return nil
}

func (a Assertion) equal(t assert.TestingT, equal Equal, env []string) error {
	envString := strings.Join(env, " ")
	env = append(os.Environ(), env...)
	actualOut, e := execCommand(equal.Actual, env)
	if e != nil {
		log.Errorf("failed to exec command %v: err %+v, output %s", equal.Actual, e, actualOut)
		return errors.Errorf("failed to exec command %v: err %+v, output %s", equal.Actual, e, actualOut)
	}
	expectedOut, e := execCommand(equal.Expected, env)
	if e != nil {
		log.Errorf("failed to exec command %v: err %+v, output %s", equal.Expected, e, expectedOut)
		return errors.Errorf("failed to exec command %v: err %+v, output %s", equal.Expected, e, expectedOut)
	}
	if expectedOut != actualOut {
		log.Errorf("not equal: %s %s", envString, equal.Actual)
		return errors.Errorf("not equal: %s %s", envString, equal.Actual)
	}
	return nil
}

func (a Assertion) runSuccess(t assert.TestingT, command string, env []string) error {
	output, e := execCommand(command, append(os.Environ(), env...))
	log.Infof("command: %s, output: %s, err: %+v", command, output, e)
	if e != nil {
		return errors.Errorf("failed to exec command %v: err %+v, output %s", command, e, output)
	}
	return nil
}
