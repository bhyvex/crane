// Copyright 2013 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"errors"
	"github.com/globocom/tsuru/cmd"
	"io/ioutil"
	"launchpad.net/gocheck"
	"net/http"
	"os"
	"os/exec"
	"testing"
)

type S struct {
	recover []string
}

func (s *S) SetUpSuite(c *gocheck.C) {
	targetFile := os.Getenv("HOME") + "/.tsuru_target"
	_, err := os.Stat(targetFile)
	if err == nil {
		old := targetFile + ".old"
		s.recover = []string{"mv", old, targetFile}
		exec.Command("mv", targetFile, old)
	} else {
		s.recover = []string{"rm", targetFile}
	}
	f, err := os.Create(targetFile)
	c.Assert(err, gocheck.IsNil)
	f.Write([]byte("http://localhost"))
	f.Close()
}

func (s *S) TearDownSuite(c *gocheck.C) {
	exec.Command(s.recover[0], s.recover[1:]...).Run()
}

type transport struct {
	msg    string
	status int
}

type conditionalTransport struct {
	transport
	condFunc func(*http.Request) bool
}

var _ = gocheck.Suite(&S{})
var manager *cmd.Manager

func Test(t *testing.T) { gocheck.TestingT(t) }

func (t *transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	resp = &http.Response{
		Body:       ioutil.NopCloser(bytes.NewBufferString(t.msg)),
		StatusCode: t.status,
	}
	return resp, nil
}

func (t *conditionalTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if !t.condFunc(req) {
		return &http.Response{Body: nil, StatusCode: 500}, errors.New("condition failed")
	}
	return t.transport.RoundTrip(req)
}

func (s *S) SetUpTest(c *gocheck.C) {
	var stdout, stderr bytes.Buffer
	manager = cmd.NewManager("glb", version, "Supported-Crane", &stdout, &stderr, os.Stdin)
}
