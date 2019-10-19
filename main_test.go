package main

import (
	"testing"
	"github.com/valyala/fasthttp"
)

func TestPeekLocation(t *testing.T) {
	req := fasthttp.AcquireRequest()
	res := fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseResponse(res)
		fasthttp.ReleaseRequest(req)
	}()

	req.Header.SetMethod("HEAD")
	req.SetRequestURI("https://coolman.site")
	if err := fasthttp.Do(req, res); nil != err {
		t.Error(err)
	}
	location := res.Header.Peek("location")
	if nil != location {
		t.Errorf("%#v", location)
	}
}
