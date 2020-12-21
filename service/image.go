/***********************************************
        File Name: image
        Author: Abby Cin
        Mail: abbytsing@gmail.com
        Created Time: 10/20/19 4:24 PM
***********************************************/

package service

import (
	"blog/conf"
	"blog/routers"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

type Image struct {
	api     string
	dstPath string
	prefix  string
}

func NewImage(cfg *conf.Config, r *routers.Router, middleware ...routers.IHandler) *Image {
	i := &Image{
		api:     cfg.Service.Image,
		dstPath: cfg.Common.Images,
		prefix:  cfg.Service.Image[0:strings.Index(cfg.Service.Image, "*")],
	}

	os.MkdirAll(i.dstPath, os.ModePerm)
	info, err := os.Stat(i.dstPath)
	if err != nil {
		log.Panicf("can't create image directory: %s\n", err)
	}
	if !info.IsDir() {
		log.Panicf("%s is not directory\n", i.dstPath)
	}

	r.POST(i.api, append(middleware, i)...)
	r.ServeFile(i.api, cfg.Common.Images, cfg.Mimes)
	return i
}

func (i *Image) Serve(c *routers.Context) {
	switch c.Req.Method {
	case "POST":
		i.Post(c)
	default:
		c.Json(http.StatusBadRequest, newError(-1, "unsupport method"))
	}
}

func (i *Image) Post(c *routers.Context) {
	// 10MB
	if e := c.Req.ParseMultipartForm(10 << 20); e != nil {
		c.Json(http.StatusBadRequest, newError(-1, e.Error()))
		return
	}
	f, _, e := c.Req.FormFile("image")
	if e != nil {
		c.Json(http.StatusBadRequest, newError(-1, e.Error()))
		return
	}
	defer f.Close()

	name := c.Req.URL.Query().Get("filename")
	if name == "" {
		c.Json(http.StatusBadRequest, newError(-1, "no filename found"))
		return
	}

	dst, link := i.getPath(name)
	out, e := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if e != nil {
		c.Json(http.StatusBadRequest, routers.J{
			"code":  -1,
			"error": e.Error(),
		})
		return
	}
	defer out.Close()
	io.Copy(out, f)

	c.Json(http.StatusOK, routers.J{
		"code":  0,
		"error": "",
		"link":  link,
		"name":  name,
	})
}

func (i *Image) getPath(name string) (string, string) {
	now := time.Now().Format("2006/01/02")
	root := path.Join(i.dstPath, now)
	os.MkdirAll(root, os.ModePerm)
	return path.Join(root, name), path.Join(i.prefix, now, name)
}
