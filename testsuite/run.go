package testsuite

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/projecteru2/grpctest/pbreflect"
	log "github.com/sirupsen/logrus"
)

var (
	patExecCommand *regexp.Regexp
	t              = &T{}
)

const resultReportTemplate = `
================================================================================
%s
================================================================================
`

func init() { // nolint:gochecknoinits
	patExecCommand = regexp.MustCompile(`"\$\(.*?\)"[,}\]]`)
}

// Run .
func Run(r io.Reader, service *pbreflect.Service) {
	scanner := bufio.NewScanner(r)
	results := []string{}
	defer func() {
		log.Info(strings.Join(results, "\n"))
	}()
	for {
		if !scanner.Scan() {
			return
		}

		// then there must be 4 lines followed
		method := scanner.Text()
		if !scanner.Scan() {
			log.Fatalf("no request line for %s", method)
		}

		prepare := scanner.Text()
		if !scanner.Scan() {
			log.Fatalf("no prepare line for %s", method)
		}
		if err := json.Unmarshal([]byte(prepare), &prepare); err != nil {
			log.Fatalf("failed to parse prepare line for %s: %+v", method, err)
		}

		request := scanner.Text()
		if !scanner.Scan() {
			log.Fatalf("no assertion line for %s", method)
		}

		assertion := scanner.Text()

		prepareOutput, err := execCommand(prepare, os.Environ())
		if err != nil {
			log.Errorf("method %s failed to execute prepare command: %+v, output: %s, err: %s", method, prepare, prepareOutput, err)
		}
		log.Infof("prepare command: %s, output: %s", prepare, prepareOutput)

		for _, r := range mustRender(request, prepareOutput) {
			testcase := New(method, prepare, r, assertion)
			result := testcase.Run(t, service)
			if !result {
				log.Errorf("testcase %s failed, request: %v", testcase.Method, r)
			}
			results = append(results, fmt.Sprintf(resultReportTemplate, testcase.String()))
		}
	}
}

func mustRender(request string, prepareOutput string) (res []string) {
	matches := patExecCommand.FindAllString(request, -1)
	if len(matches) == 0 {
		return []string{request}
	}

	for idx, match := range matches {
		switch {
		case strings.HasSuffix(match, `)",`):
			matches[idx] = strings.TrimSuffix(match, `,`)
		case strings.HasSuffix(match, `)"]`):
			matches[idx] = strings.TrimSuffix(match, `]`)
		case strings.HasSuffix(match, `)"}`):
			matches[idx] = strings.TrimSuffix(match, `}`)
		}
	}

	replacements := [][]string{}
	for _, match := range matches {
		command := strings.TrimPrefix(match, `"$(`)
		command = strings.TrimSuffix(command, `)"`)
		output, err := execCommand(command, append(os.Environ(), "prepare="+prepareOutput))
		if err != nil {
			log.Fatalf("failed to render request with command %s: %+v, %s", command, err, output)
		}
		output = strings.TrimSpace(output)
		replacements = append(replacements, strings.Split(output, "\n"))
	}

	for _, comb := range combine(replacements) {
		for idx, replace := range comb {
			res = append(res, strings.ReplaceAll(request, matches[idx], replace))
		}
	}
	return // nolint:nakedret
}
