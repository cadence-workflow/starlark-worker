package request

import (
	"fmt"
	"github.com/cadence-workflow/starlark-worker/ext"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"go.uber.org/cadence/testsuite"
	"net/http"
	"net/http/httptest"
	"testing"
)

type Suite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
	server *httptest.Server
	env    *testsuite.TestActivityEnvironment
}

func TestIT(t *testing.T) { suite.Run(t, new(Suite)) }

func (r *Suite) SetupSuite()    { r.server = httptest.NewServer(ext.NewHTTPTestHandler(r.T())) }
func (r *Suite) TearDownSuite() { r.server.Close() }

func (r *Suite) BeforeTest(_, _ string) {
	r.env = r.NewTestActivityEnvironment()
	r.env.RegisterActivity(&activities{
		client: http.DefaultClient,
	})
}

func (r *Suite) Test_DoJSON_Get404() {

	resourceURL := fmt.Sprintf("%s/storage/%s.json", r.server.URL, uuid.New().String())
	val, err := r.env.ExecuteActivity(Activities.DoJSON, JSONRequest{
		Method: "GET",
		URL:    resourceURL,
		Assert: Assert{
			StatusCodes: []int{404},
		},
	})
	r.Require().NoError(err)
	r.Require().True(val.HasValue())

	var res any
	r.Require().NoError(val.Get(&res))
	r.Require().Equal("not-found", res)
}
