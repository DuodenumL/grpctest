package testsuite

import (
	"context"
	"log"

	"github.com/projecteru2/grpctest/pbreflect"
	"github.com/stretchr/testify/assert"
)

// Testsuite .
type Testsuite struct {
	Method  string
	Request string
	*Assertion
}

// New .
func New(method, request, assertion string) *Testsuite {
	return &Testsuite{
		Method:    method,
		Request:   request,
		Assertion: MustNewAssertion(assertion),
	}
}

// Run .
func (c Testsuite) Run(t assert.TestingT, service *pbreflect.Service) bool {
	responses, err := service.Send(context.TODO(), c.Method, c.Request)
	if err != nil {
		log.Fatalf("failed to send request %s: %+v", c.Request, err)
	}

	contents, errs := []string{}, []string{}
	for response := range responses {
		contents = append(contents, response.Content)
		errs = append(errs, response.Err)
		if !c.assertEach(t, c.Request, response.Content, response.Err) {
			return false
		}
	}
	return c.assertCompletion(t, c.Request, contents, errs)
}
