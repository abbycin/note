/***********************************************
        File Name: statics
        Author: Abby Cin
        Mail: abbytsing@gmail.com
        Created Time: 9/22/19 8:03 AM
***********************************************/

package routers

import (
	"errors"
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	path2 "path"
	"path/filepath"
	"strconv"
	"strings"
)

type statics struct {
	wild  bool
	key   string
	path  string
	etag  string
	mime  string
	data  []byte
	limit int64 // pre-item size limit, if size exceed limit, it will not be cached
	mimes map[string]string
	lru   *LRU
	os.FileInfo
}

func (s *statics) Write(c *Context) {
	h := c.Resp.Header()
	h.Add("Content-Type", s.mime)
	h.Add("ETag", s.etag)
	if s.etag == c.Req.Header.Get("If-None-Match") {
		c.Resp.WriteHeader(http.StatusNotModified)
		c.Resp.Write(nil)
		return
	}

	c.Resp.WriteHeader(http.StatusOK)
	h.Add("Content-Length", strconv.FormatInt(int64(s.Size()), 10))
	c.Resp.Write(s.data)
}

func (s *statics) Update(args ...interface{}) (interface{}, error) {
	info, err := os.Stat(s.path)
	if err != nil {
		return false, err
	}
	s.FileInfo = info
	if info.IsDir() {
		return false, errors.New("can't be directory")
	}
	//s.mime = http.DetectContentType(s.data)
	ext := strings.ToLower(filepath.Ext(s.Name()))
	s.mime = ""
	if len(ext) > 0 {
		ok := true
		s.mime, ok = s.mimes[ext[1:]]
		if !ok {
			s.mime = mime.TypeByExtension(ext)
		}
	}

	if s.mime == "" {
		s.mime = "application/octet-stream"
	}
	if err != nil || info.IsDir() {
		return false, err
	}

	tmp := fmt.Sprintf("%v", info.ModTime().Unix())

	if tmp != s.etag {
		s.data, err = ioutil.ReadFile(s.path)
		if err != nil {
			return false, err
		}
		s.etag = tmp
		return true, nil
	} else {
		return false, nil
	}
}

func (s *statics) Name() string {
	return s.path
}

func (s *statics) Size() int64 {
	return s.FileInfo.Size()
}

func (s *statics) MaxSize() int64 {
	return s.limit
}

func (s *statics) Data() []byte {
	return s.data
}

func (s *statics) Serve(c *Context) {
	value := s.key
	if s.wild {
		value = c.GetParam(s.key)
	}
	path := path2.Join(s.path, value)

	data := s.lru.Get(path)
	if data == nil {
		data = newStatic(nil, s.key, path, s.limit, s.mimes)
		_, err := data.Update()
		if err != nil {
			// write log
			http.NotFound(c.Resp, c.Req)
			return
		}
		if data.Size() < data.MaxSize() {
			s.lru.Add(data)
		}
	} else {
		// if not modified, response header
		updated, err := data.Update()
		if err != nil {
			c.Resp.WriteHeader(http.StatusNotFound)
			return
		}
		if updated.(bool) {
			s.lru.Update(data)
		}
	}
	data.Write(c)
}

func newStatic(lru *LRU, path, local string, limit int64, mimes map[string]string) *statics {
	if strings.Index(path, ":") >= 0 {
		panic("placeholder ':' is not allowed for serving static files, " + path)
	}

	wild := false
	if idx := strings.Index(path, "*"); idx >= 0 {
		if idx == 0 || path[idx-1] != '/' {
			panic("wildcard doesn't after a '/', " + path)
		}
		if path[len(path)-1] == '/' {
			panic("wildcard must at the end of path, " + path)
		}
		wild = true
	}

	tmp := strings.Split(path, "/")
	key := ""
	if wild {
		key = tmp[len(tmp)-1][1:] // for example: path is `/foo/*key`, then key is `key`
	} else {
		key = tmp[len(tmp)-1] // for example: path is `/foo/bar/x.js`, then key is `x.js`
	}

	r := &statics{
		wild:  wild,
		key:   key,
		path:  local,
		lru:   lru,
		limit: limit,
		mimes: mimes,
	}
	return r
}
