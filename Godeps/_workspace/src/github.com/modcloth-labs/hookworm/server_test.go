package hookworm

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"runtime"
	"strings"
	"testing"

	"github.com/codegangsta/martini"
)

var (
	serverTestConfig = &HandlerConfig{
		GithubPath: "/github-test",
		TravisPath: "/travis-test",
		Debug:      true,
	}
	here               = ""
	serverTestHandlers = map[string]string{
		"00-echo.py": `#!/usr/bin/env python
import sys

if sys.argv[1] != 'configure':
    sys.stdout.write(sys.stdin.read())
sys.exit(0)
`,
		"10-eats-json.py": `#!/usr/bin/env python
import sys
import json

if sys.argv[1] != 'configure':
    payload = json.load(sys.stdin)
    json.dump(payload, sys.stdout)
sys.exit(0)
`,
		"15-skipped.sh": `#!/bin/bash
exit 1`,
		"20-github-only.py": `#!/usr/bin/env python
import sys

if sys.argv[1] == 'configure':
    sys.exit(0)
elif sys.argv[1:3] == ['handle', 'github']:
    sys.exit(0)
else:
    sys.exit(78)
`,
		"30-travis-only.py": `#!/usr/bin/env python
import sys

if sys.argv[1] == 'configure':
    sys.exit(0)
elif sys.argv[1:3] == ['handle', 'travis']:
    sys.exit(0)
else:
    sys.exit(78)
`,
		".hidden.py": `#!/usr/bin/env python
import sys

sys.exit(1)
`,
	}
)

func init() {
	debug = true
	setHere()
	createServerTestWormDir()
	fillInServerTestHandlers()
}

func setHere() {
	_, filename, _, _ := runtime.Caller(0)
	here = path.Dir(filename)
}

func createServerTestWormDir() {
	serverTestConfig.WormDir = path.Join(os.TempDir(), "hookworm-test-worm.d")
	os.RemoveAll(serverTestConfig.WormDir)
	err := os.MkdirAll(serverTestConfig.WormDir, 0750)
	if err != nil {
		panic(err)
	}
}

func fillInServerTestHandlers() {
	if len(serverTestConfig.WormDir) < 1 {
		panic("No worm dir, nerds!")
	}

	for filename, stringContent := range serverTestHandlers {
		content := []byte(strings.TrimSpace(stringContent))
		ioutil.WriteFile(path.Join(serverTestConfig.WormDir, filename), content, 0750)
	}
}

func getPayload(kind, name string) string {
	parts := []string{here, "sampledata"}
	if kind == "github" {
		parts = append(parts, "github-payloads")
	} else if kind == "travis" {
		parts = append(parts, "travis-payloads")
	}
	parts = append(parts, name+".json")
	bytes, err := ioutil.ReadFile(path.Join(parts...))
	if err != nil {
		panic(err)
	}
	return string(bytes)
}

func getPayloadJSONReader(kind, name string) io.Reader {
	return strings.NewReader(getPayload(kind, name))
}

func getPayloadFormReader(kind, name string) io.Reader {
	payload := getPayload(kind, name)
	v := url.Values{"payload": []string{payload}}
	return strings.NewReader(v.Encode())
}

func setupServer() (*httptest.ResponseRecorder, *martini.ClassicMartini) {
	m, err := NewServer(serverTestConfig)
	if err != nil {
		panic(err)
	}
	hr := httptest.NewRecorder()
	m.MapTo(hr, (*http.Handler)(nil))
	return hr, m
}

func getResponse(verb, path, ctype string, body io.Reader) *httptest.ResponseRecorder {
	hr, m := setupServer()

	req, err := http.NewRequest(verb, path, body)
	if err != nil {
		panic(err)
	}

	if ctype != "" {
		req.Header.Set("Content-Type", ctype)
	}

	m.ServeHTTP(hr, req)
	return hr
}

func TestServerRespondsToIndex(t *testing.T) {
	if getResponse("GET", "/", "", nil).Code != 200 {
		t.Fail()
	}
}

func TestServerRespondsToTestPage(t *testing.T) {
	if getResponse("GET", "/debug/test", "", nil).Code != 200 {
		t.Fail()
	}
}

func TestServerRespondsToGithubJSON(t *testing.T) {
	if getResponse("POST", "/github-test", "application/json",
		getPayloadJSONReader("github", "valid")).Code != 204 {
		t.Fail()
	}
}

func TestServerRespondsToGithubForm(t *testing.T) {
	if getResponse("POST", "/github-test", "application/x-www-form-urlencoded",
		getPayloadFormReader("github", "valid")).Code != 204 {
		t.Fail()
	}
}

func TestServerRespondsToTravisJSON(t *testing.T) {
	if getResponse("POST", "/travis-test", "application/json",
		getPayloadJSONReader("travis", "valid")).Code != 204 {
		t.Fail()
	}
}

func TestServerRespondsToTravisForm(t *testing.T) {
	if getResponse("POST", "/travis-test", "application/x-www-form-urlencoded",
		getPayloadFormReader("travis", "valid")).Code != 204 {
		t.Fail()
	}
}
