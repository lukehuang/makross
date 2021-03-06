package logger

import (
	"bytes"
	"errors"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/insionng/makross"
	"github.com/stretchr/testify/assert"
)

func TestLogger(t *testing.T) {
	// Note: Just for the test coverage, not a real test.
	e := makross.New()

	req := httptest.NewRequest(makross.GET, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec, func(c *makross.Context) error {
		return c.String("test", makross.StatusOK)
	})
	h := Logger()

	// Status 2xx
	h(c)

	// Status 3xx
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec, func(c *makross.Context) error {
		return c.String("test", makross.StatusTemporaryRedirect)
	})
	h = Logger()
	h(c)

	// Status 4xx
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec, func(c *makross.Context) error {
		return c.String("test", makross.StatusNotFound)
	})
	h = Logger()
	h(c)

	// Status 5xx with empty path
	req = httptest.NewRequest(makross.GET, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec, func(c *makross.Context) error {
		return errors.New("error")
	})
	h = Logger()
	h(c)
}

func TestLoggerIPAddress(t *testing.T) {
	e := makross.New()

	req := httptest.NewRequest(makross.GET, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec, func(c *makross.Context) error {
		return c.String("test", makross.StatusOK)
	})
	buf := new(bytes.Buffer)
	//e.Logger.SetOutput(buf)
	ip := "127.0.0.1"
	h := Logger()

	// With X-Real-IP
	req.Header.Add(makross.HeaderXRealIP, ip)
	h(c)
	assert.Contains(t, ip, buf.String())

	// With X-Forwarded-For
	buf.Reset()
	req.Header.Del(makross.HeaderXRealIP)
	req.Header.Add(makross.HeaderXForwardedFor, ip)
	h(c)
	assert.Contains(t, ip, buf.String())

	buf.Reset()
	h(c)
	assert.Contains(t, ip, buf.String())
}

func TestLoggerTemplate(t *testing.T) {
	buf := new(bytes.Buffer)

	e := makross.New()

	e.Use(LoggerWithConfig(LoggerConfig{
		Format: `{"time":"${time_rfc3339_nano}","id":"${id}","remote_ip":"${remote_ip}","host":"${host}","user_agent":"${user_agent}",` +
			`"method":"${method}","uri":"${uri}","status":${status}, "latency":${latency},` +
			`"latency_human":"${latency_human}","bytes_in":${bytes_in}, "path":"${path}", "referer":"${referer}",` +
			`"bytes_out":${bytes_out},"ch":"${header:X-Custom-Header}",` +
			`"us":"${query:username}", "cf":"${form:username}", "session":"${cookie:session}"}` + "\n",
		Output: buf,
	}))

	e.Get("/", func(c *makross.Context) error {
		return c.String("Header Logged", makross.StatusOK)
	})

	req := httptest.NewRequest(makross.GET, "/?username=apagano-param&password=secret", nil)
	req.RequestURI = "/"
	req.Header.Add(makross.HeaderXRealIP, "127.0.0.1")
	req.Header.Add("Referer", "google.com")
	req.Header.Add("User-Agent", "makross-tests-agent")
	req.Header.Add("X-Custom-Header", "AAA-CUSTOM-VALUE")
	req.Header.Add("X-Request-ID", "6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	req.Header.Add("Cookie", "_ga=GA1.2.000000000.0000000000; session=ac08034cd216a647fc2eb62f2bcf7b810")
	req.Form = url.Values{
		"username": []string{"apagano-form"},
		"password": []string{"secret-form"},
	}

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	cases := map[string]bool{
		"apagano-param":                        true,
		"apagano-form":                         true,
		"AAA-CUSTOM-VALUE":                     true,
		"BBB-CUSTOM-VALUE":                     false,
		"secret-form":                          false,
		"hexvalue":                             false,
		"GET":                                  true,
		"127.0.0.1":                            true,
		"\"path\":\"/\"":                       true,
		"\"uri\":\"/\"":                        true,
		"\"status\":200":                       true,
		"\"bytes_in\":0":                       true,
		"google.com":                           true,
		"makross-tests-agent":                  true,
		"6ba7b810-9dad-11d1-80b4-00c04fd430c8": true,
		"ac08034cd216a647fc2eb62f2bcf7b810":    true,
	}

	for token, present := range cases {
		assert.True(t, strings.Contains(buf.String(), token) == present, "Case: "+token)
	}
}
