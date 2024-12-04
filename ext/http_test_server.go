package ext

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"path"
	"sync"
	"testing"
)

type HTTPTestHandler interface {
	http.Handler
	GetResources() map[string][]byte
}

type testHandler struct {
	t             *testing.T
	resources     map[string][]byte
	resourcesLock sync.RWMutex
}

var _ HTTPTestHandler = (*testHandler)(nil)

func NewHTTPTestHandler(t *testing.T) HTTPTestHandler {
	return &testHandler{
		t:         t,
		resources: map[string][]byte{},
	}
}

func (r *testHandler) GetResources() map[string][]byte {
	return r.resources
}

func (r *testHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	t := r.t
	res.Header().Set("content-type", "application/json; charset=utf-8")
	_res := jsoniter.NewEncoder(res)
	key := path.Clean(req.URL.Path)
	switch req.Method {
	case "PUT":
		r.resourcesLock.Lock()
		if body, err := io.ReadAll(req.Body); err != nil {
			res.WriteHeader(500)
			require.NoError(t, _res.Encode(err.Error()))
		} else {
			r.resources[key] = body
			res.WriteHeader(200)
			require.NoError(t, _res.Encode("ok"))
		}
		r.resourcesLock.Unlock()
	case "GET":
		r.resourcesLock.RLock()
		if body, found := r.resources[key]; !found {
			res.WriteHeader(404)
			require.NoError(t, _res.Encode("not-found"))
		} else {
			res.WriteHeader(200)
			_, err := res.Write(body)
			require.NoError(t, err)
		}
		r.resourcesLock.RUnlock()
	case "DELETE":
		r.resourcesLock.Lock()
		if _, found := r.resources[key]; !found {
			res.WriteHeader(404)
			require.NoError(t, _res.Encode("not-found"))
		} else {
			delete(r.resources, key)
			res.WriteHeader(200)
			require.NoError(t, _res.Encode("ok"))
		}
		r.resourcesLock.Unlock()
	default:
		res.WriteHeader(500)
		require.NoError(t, _res.Encode(fmt.Sprintf("unsupported method: %s", req.Method)))
	}
}
