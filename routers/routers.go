/***********************************************
        File Name: routers
        Author: Abby Cin
        Mail: abbytsing@gmail.com
        Created Time: 9/14/19 3:50 PM
***********************************************/

package routers

import (
	"blog/session"
	"net/http"
)

type J map[string]interface{}

type Handler func(c *Context)

func (h Handler) Serve(c *Context) {
	h(c)
}

type Router struct {
	trie         map[string]*node
	lru          *LRU
	mgr          *session.SessionManager
	proxyMode    bool
	PanicHandler func(http.ResponseWriter, *http.Request, interface{})
	NotFound     func(w http.ResponseWriter, r *http.Request)
	callbacks    []IHandler
}

func New(mgr *session.SessionManager, proxyMode bool) *Router {
	return &Router{
		trie:      make(map[string]*node),
		lru:       NewLRU(100, 50<<20), // cache max 100 items total 50MB
		mgr:       mgr,
		proxyMode: proxyMode,
		callbacks: make([]IHandler, 0),
		NotFound:  http.NotFound,
	}
}

func (r *Router) AddCb(cb IHandler) {
	r.callbacks = append(r.callbacks, cb)
}

func (r *Router) dispatch(w http.ResponseWriter, req *http.Request) {

	root, ok := r.trie[req.Method]
	if !ok || root == nil {
		http.NotFound(w, req)
		return
	}

	param := &paramImpl{}
	handlers := root.search(req.URL.Path, param)
	if handlers == nil {
		r.NotFound(w, req)
	} else {
		ctx := &Context{req, w, param, r.mgr, true, r.proxyMode}
		for _, cb := range r.callbacks {
			cb.Serve(ctx)
			if !ctx.next {
				break
			}
		}
		for _, h := range handlers {
			h.Serve(ctx)
			if !ctx.next {
				break
			}
		}
	}
}

func (r *Router) AddRoute(method, path string, handler ...IHandler) {
	if len(path) < 1 || path[0] != '/' {
		panic("path must begin with '/'")
	}

	root := r.trie[method]
	if root == nil {
		root = new(node)
		r.trie[method] = root
	}
	root.addNode(path, handler...)
}

func (r *Router) catcher(w http.ResponseWriter, req *http.Request) {
	if rcv := recover(); rcv != nil {
		r.PanicHandler(w, req, rcv)
	}
}

func (r *Router) ServeFile(path, local string, mimes map[string]string) {
	r.AddRoute(http.MethodGet, path, newStatic(r.lru, path, local, 3<<20, mimes))
}

func (r *Router) GET(path string, handler ...IHandler) {
	r.AddRoute(http.MethodGet, path, handler...)
}

func (r *Router) HEAD(path string, handler ...IHandler) {
	r.AddRoute(http.MethodHead, path, handler...)
}

func (r *Router) POST(path string, handler ...IHandler) {
	r.AddRoute(http.MethodPost, path, handler...)
}

func (r *Router) PUT(path string, handler ...IHandler) {
	r.AddRoute(http.MethodPut, path, handler...)
}

func (r *Router) OPTIONS(path string, handler ...IHandler) {
	r.AddRoute(http.MethodOptions, path, handler...)
}

func (r *Router) DELETE(path string, handler ...IHandler) {
	r.AddRoute(http.MethodDelete, path, handler...)
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if r.PanicHandler != nil {
		defer r.catcher(w, req)
	}

	r.dispatch(w, req)
}
