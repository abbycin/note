/***********************************************
        File Name: home
        Author: Abby Cin
        Mail: abbytsing@gmail.com
        Created Time: 10/28/19 9:45 PM
***********************************************/

package service

import (
	"blog/conf"
	"blog/dbutil"
	"blog/logging"
	"blog/model"
	"blog/routers"
	"fmt"
	"html/template"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"
)

type Post struct {
	Date  time.Time
	Link  template.URL
	Title string
}

type home struct {
	dao  *dbutil.Dao
	h    *Home
	text []byte
	etag string
}

func (h *home) Write(c *routers.Context) {
	if len(h.text) == 0 {
		_, err := h.Update()
		if err != nil {
			c.Text(http.StatusNotFound, "not found")
			return
		}
	}

	header := c.Resp.Header()
	header.Add("ETag", h.etag)
	if h.etag == c.Req.Header.Get("If-None-Match") {
		c.Resp.WriteHeader(http.StatusNotModified)
		c.Resp.Write(nil)
	} else {
		c.Resp.WriteHeader(http.StatusOK)
		header.Add("Content-Length", strconv.FormatInt(int64(len(h.text)), 10))
		c.Resp.Write(h.text)
	}
}

func (h *home) Makelink(p *model.Post) template.URL {
	r := path.Join(h.h.article, fmt.Sprintf("%v", p.Id))
	return template.URL(r)
}

func (h *home) Update() (interface{}, error) {
	data, err := h.dao.GetPosts()
	if err != nil {
		return nil, err
	}
	navis, err := h.dao.GetNavis()
	if err != nil {
		return nil, err
	}

	res := make([]Post, 0)
	for _, p := range data.Posts {
		if p.Hidden {
			continue
		}
		t := Post{
			Date:  time.Time(p.CreateTime),
			Link:  h.Makelink(&p),
			Title: p.Title,
		}
		res = append(res, t)
	}

	h.text, err = h.h.build(res, navis)
	if err != nil {
		return nil, err
	}
	h.etag = fmt.Sprintf("%v", time.Now().Unix())
	return nil, nil
}

type Home struct {
	model   model.Model
	cfg     *conf.Config
	article string
	cache   *home
}

func NewHome(cfg *conf.Config, dao *dbutil.Dao, r *routers.Router) *Home {
	h := &Home{
		model:   model.NewDefaultModel(path.Join(cfg.Model.TmplRoot, cfg.Model.Home.Tmpl)),
		cfg:     cfg,
		article: cfg.Model.Article.Api[:strings.Index(cfg.Model.Article.Api, ":")],
		cache:   &home{dao: dao},
	}

	h.cache.h = h

	r.GET(cfg.Model.Home.Api, h)
	return h
}

func (h *Home) build(posts []Post, navis *model.NaviData) ([]byte, error) {
	h.model.Funcs(template.FuncMap{"formatTime": h.formatTime})
	return h.model.Parse(map[string]interface{}{
		"Title": h.cfg.Model.Title,
		"Posts": posts,
		"Navis": navis,
	})
}

func (h *Home) Serve(c *routers.Context) {
	h.cache.Write(c)
}

func (h *Home) formatTime(arg interface{}) string {
	return arg.(time.Time).Format("2006 01-02 15:04:05")
}

func (h *Home) Update(arg interface{}) {
	_, err := h.cache.Update()
	if err != nil {
		logging.Error("upate home cache: %v", err)
	}
}
