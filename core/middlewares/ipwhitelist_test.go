package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIPWhitelist(t *testing.T) {
	// 192.0.2.0/24 is "TEST-NET" in RFC 5737 for use solely in
	// documentation and example source code and should not be
	// used publicly.
	testIPWhitelist(t, map[string]bool{"192.0.2.1": true}, http.StatusOK)
	testIPWhitelist(t, map[string]bool{"127.0.0.1": true}, http.StatusForbidden)
	testIPWhitelist(t, map[string]bool{"192.0.2.1": false}, http.StatusForbidden)
	testIPWhitelist(t, map[string]bool{"192.0.2.2/8": true}, http.StatusOK)
	testIPWhitelist(t, map[string]bool{"192.0.2.2/8": false}, http.StatusForbidden)
	testIPWhitelist(t, map[string]bool{"192.0.2.2/8": true, "192.0.2.1": false}, http.StatusForbidden)
	testIPWhitelist(t, map[string]bool{"192.0.2.2/8": true, "192.0.2.1": true}, http.StatusOK)
	testIPWhitelist(t, nil, http.StatusForbidden)
}

func testIPWhitelist(t *testing.T, whitelist map[string]bool, expectedStatusCode int) {
	mw := IPWhitelist(whitelist)
	assert.NotNil(t, mw)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)

	mw(c)
	res := w.Result()
	defer res.Body.Close()
	assert.Equal(t, expectedStatusCode, res.StatusCode)

	mw(c)
	assert.Equal(t, expectedStatusCode, res.StatusCode)

	mw(c)
	assert.Equal(t, expectedStatusCode, res.StatusCode)
}
