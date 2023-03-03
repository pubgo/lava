package whitelists

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
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

	var app = fiber.New()
	app.Get("/", func(ctx *fiber.Ctx) error {
		return mw(ctx)
	})

	assert.NotNil(t, mw)
	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/", nil))
	assert.Nil(t, err)
	data, err := ioutil.ReadAll(resp.Body)
	assert.Nil(t, err)
	t.Log(string(data))
	assert.Equal(t, expectedStatusCode, resp.StatusCode)
}
