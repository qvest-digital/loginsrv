package login

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tarent/loginsrv/model"
)

const (
	endpointPath = "/claims"
	token        = "token"
	timeout      = time.Second
)

var aUserInfo = model.UserInfo{
	Sub:    "test@example.com",
	Origin: "origin",
	Domain: "example.com",
}

func Test_newUserClaimsProvider_ValidatesURL(t *testing.T) {
	_, err := newUserClaimsProvider("@#$%^&*(", "auth", time.Minute)

	assert.Error(t, err)
}

func Test_userClaimsProvider_Claims(t *testing.T) {
	mock := createMockServer(
		mockResponse{
			url:    endpointPath,
			status: http.StatusOK,
			body: `{
	"claims": [
		{ "role": "admin" }
	]
}`,
		},
	)
	defer mock.Close()
	provider, err := newUserClaimsProvider(mock.URL+endpointPath, token, time.Minute)
	require.NoError(t, err)

	claims, err := provider.Claims(model.UserInfo{
		Sub:    "test@example.com",
		Origin: "origin",
		Domain: "example.com",
	})

	require.NoError(t, err)

	assert.Equal(t, 1, len(mock.requests))

	request := mock.requests[0]
	assertQueryValue(t, "sub", "test@example.com", request.URL)
	assertQueryValue(t, "origin", "origin", request.URL)
	assertQueryValue(t, "domain", "example.com", request.URL)

	assert.Equal(t, "Bearer "+token, request.Header.Get("Authorization"))

	assert.Equal(t,
		customClaims{
			"claims": []interface{}{
				map[string]interface{}{
					"role": "admin",
				},
			},
			"domain": "example.com",
			"origin": "origin",
			"sub":    "test@example.com",
		},
		claims,
	)
}

func Test_userClaimsProvider_Claims_NotFound(t *testing.T) {
	mock := createMockServer(
		mockResponse{
			url:    endpointPath,
			status: http.StatusNotFound,
			body:   ``,
		},
	)
	defer mock.Close()
	provider, err := newUserClaimsProvider(mock.URL+endpointPath, token, time.Minute)
	require.NoError(t, err)

	claims, err := provider.Claims(model.UserInfo{
		Sub:    "test@example.com",
		Origin: "origin",
		Domain: "example.com",
	})

	require.NoError(t, err)

	assert.Equal(t,
		customClaims{
			"domain": "example.com",
			"origin": "origin",
			"sub":    "test@example.com",
		},
		claims,
	)
}

func Test_userClaimsProvider_Claims_EndpointNotReachable(t *testing.T) {
	provider, err := newUserClaimsProvider("http://not-exists.example.com", token, time.Millisecond)
	require.NoError(t, err)

	_, err = provider.Claims(aUserInfo)

	assert.Error(t, err)
}

func Test_userClaimsProvider_Claims_Errors(t *testing.T) {
	for _, test := range []struct {
		name   string
		status int
		body   string
	}{
		{
			name:   "invalid json body",
			status: http.StatusOK,
			body:   `}{`,
		},
		{
			name:   "not 200 response code",
			status: http.StatusForbidden,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			mock := createMockServer(
				mockResponse{
					url:    endpointPath,
					status: test.status,
					body:   test.body,
				},
			)
			defer mock.Close()
			provider, err := newUserClaimsProvider(mock.URL+endpointPath, token, time.Minute)
			require.NoError(t, err)

			_, err = provider.Claims(aUserInfo)

			assert.Error(t, err)
		})
	}
}

type mockServer struct {
	*httptest.Server
	requests []*http.Request
}

type mockResponse struct {
	url, body string
	status    int
}

func createMockServer(responses ...mockResponse) *mockServer {
	mux := http.NewServeMux()
	server := &mockServer{
		httptest.NewServer(mux),
		[]*http.Request{},
	}

	for _, response := range responses {
		body := response.body
		mux.HandleFunc(
			response.url,
			func(w http.ResponseWriter, r *http.Request) {
				server.requests = append(server.requests, r)

				w.WriteHeader(response.status)
				w.Write([]byte(body))
			})
	}

	return server
}

func assertQueryValue(t *testing.T, name, expectedValue string, u *url.URL) {
	value, err := url.QueryUnescape(u.Query().Get(name))
	require.NoError(t, err)
	assert.Equal(t, expectedValue, value)
}
