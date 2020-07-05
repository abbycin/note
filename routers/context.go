/***********************************************
        File Name: context
        Author: Abby Cin
        Mail: abbytsing@gmail.com
        Created Time: 10/4/19 1:02 PM
***********************************************/

package routers

import (
	"blog/session"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

const (
	tJson = "application/json; charset=utf-8"
	tTxt  = "plain/text"
	tHtml = "plain/html"
)

type HandlerMap struct {
	handler IHandler
	called  bool
}

type Context struct {
	Req       *http.Request
	Resp      http.ResponseWriter
	param     *paramImpl
	mgr       *session.SessionManager
	abort     bool
	proxyMode bool
	handlers  []HandlerMap
}

func newContext(r *Router, param *paramImpl, handlers []IHandler, req *http.Request, resp http.ResponseWriter) *Context {
	ctx := &Context{
		Req:       req,
		Resp:      resp,
		param:     param,
		mgr:       r.mgr,
		abort:     false,
		proxyMode: r.proxyMode,
		handlers:  make([]HandlerMap, 0),
	}

	for _, h := range handlers {
		ctx.handlers = append(ctx.handlers, HandlerMap{
			handler: h,
			called:  false,
		})
	}

	return ctx
}

func (c *Context) simpleResponse(mime string, code int, data []byte) {
	c.Resp.Header().Add("Context-Type", mime)
	c.Resp.Header().Add("Content-Length", strconv.FormatInt(int64(len(data)), 10))
	c.Resp.WriteHeader(code)
	if len(data) > 0 {
		c.Resp.Write(data)
	}
}

func (c *Context) GetParam(k string) string {
	return c.param.get(k)
}

func (c *Context) StartSession() session.ISession {
	return c.mgr.StartSession(c.Resp, c.Req)
}

func (c *Context) ClearSession() {
	c.mgr.Clear()
}

func (c *Context) SessionDestroy() {
	c.mgr.SessionDestroy(c.Resp, c.Req)
}

func (c *Context) Json(code int, j interface{}) {
	data, err := json.Marshal(j)
	if err != nil {
		panic(fmt.Sprintf("Json: %s", err.Error()))
	}
	c.simpleResponse(tJson, code, data)
}

func (c *Context) Text(code int, t string) {
	c.simpleResponse(tTxt, code, []byte(t))
}

func (c *Context) Html(code int, t []byte) {
	c.simpleResponse(tHtml, code, t)
}

func (c *Context) Unmarshal(out interface{}) error {
	data, err := ioutil.ReadAll(c.Req.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, out)
}

func (c *Context) Next() {
	for _, h := range c.handlers {
		if c.abort {
			break
		}
		if h.called {
			continue
		}
		h.called = true
		h.handler.Serve(c)
	}
}

func (c *Context) Abort() {
	c.abort = true
}

func (c *Context) RemoteAddr() string {
	if c.proxyMode {
		return c.Req.Header.Get("X-Real-IP")
	} else {
		return strings.Split(c.Req.RemoteAddr, ":")[0]
	}
}
