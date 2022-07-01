package testsuite

import (
	"context"
	"encoding/json"

	"github.com/projecteru2/grpctest/pbreflect"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// Testsuite .
type Testsuite struct {
	Method     string `json:"method"`
	Prepare    string `json:"prepare"`
	Request    string `json:"request"`
	*Assertion `json:"assertion"`
	Success    bool  `json:"success"`
	Error      error `json:"errors"`
}

// New .
func New(method, prepare, request, assertion string) *Testsuite {
	return &Testsuite{
		Method:    method,
		Prepare:   prepare,
		Request:   request,
		Assertion: MustNewAssertion(assertion),
	}
}

// Run .
func (c *Testsuite) Run(t assert.TestingT, service *pbreflect.Service) bool {
	defer func() {
		c.Success = c.Error == nil
	}()
	responses, err := service.Send(context.TODO(), c.Method, c.Request)
	if err != nil {
		log.Errorf("failed to send request %s: %+v", c.Request, err)
		c.Error = err
		return false
	}

	contents, errs := []string{}, []string{}
	for response := range responses {
		contents = append(contents, response.Content)
		errs = append(errs, response.Err)
		if err := c.assertEach(t, c.Request, response.Content, response.Err); err != nil {
			c.Error = err
			return false
		}
	}
	if err := c.assertCompletion(t, c.Request, contents, errs); err != nil {
		c.Error = err
		return false
	}
	return true
}

// String .
func (c Testsuite) String() string {
	body, _ := json.MarshalIndent(c, "", "\t")
	return string(body)
}
