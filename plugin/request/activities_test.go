package request

import (
	"fmt"
	"github.com/cadence-workflow/starlark-worker/ext"
	"github.com/cadence-workflow/starlark-worker/service"
	"github.com/cadence-workflow/starlark-worker/test/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
	"testing"
)

type Suite struct {
	suite.Suite
	server        *httptest.Server
	activitySuite types.StarTestActivitySuite
}

func TestITCadence(t *testing.T) {
	suite.Run(t, &Suite{activitySuite: service.NewCadTestActivitySuite()})
}

func TestITTemporal(t *testing.T) {
	suite.Run(t, &Suite{activitySuite: service.NewTempTestActivitySuite()})
}

func (r *Suite) SetupSuite()    { r.server = httptest.NewServer(ext.NewHTTPTestHandler(r.T())) }
func (r *Suite) TearDownSuite() { r.server.Close() }

func (r *Suite) BeforeTest(_, _ string) {
	r.activitySuite.RegisterActivity(&activities{
		client: http.DefaultClient,
	})
}

func (r *Suite) Test_DoJSON_Get404() {
	resourceURL := fmt.Sprintf("%s/storage/%s.json", r.server.URL, uuid.New().String())
	val, err := r.activitySuite.ExecuteActivity(Activities.DoJSON, JSONRequest{
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
