/***********************************************
        File Name: article
        Author: Abby Cin
        Mail: abbytsing@gmail.com
        Created Time: 10/27/19 10:23 PM
***********************************************/

package service

import (
	"blog/conf"
	"blog/dbutil"
	"blog/logging"
	"blog/model"
	"blog/routers"
	"errors"
	"fmt"
	"github.com/russross/blackfriday"
	"html/template"
	"net/http"
	"path"
	"strconv"
	"time"
)

type article struct {
	dao   *dbutil.Dao
	id    int64
	ids   string
	etag  string
	limit int64
	a     *Article
	text  []byte
}

func (c *article) Write(ctx *routers.Context) {
	if len(c.text) == 0 {
		http.NotFound(ctx.Resp, ctx.Req)
		return
	}
	h := ctx.Resp.Header()
	h.Add("ETag", c.etag)
	if c.etag == ctx.Req.Header.Get("If-None-Match") {
		ctx.Resp.WriteHeader(http.StatusNotModified)
		ctx.Resp.Write(nil)
	} else {
		ctx.Html(http.StatusOK, c.text)
	}
}

func (c *article) Name() string {
	return c.ids
}

func (c *article) Size() int64 {
	return int64(len(c.text))
}

func (c *article) Update(args ...interface{}) (interface{}, error) {
	includingHide := false
	if args != nil {
		includingHide = args[0].(bool)
	}
	post, err := c.dao.GetArticle(c.id, includingHide)
	if err != nil {
		return nil, err
	}
	if post == nil {
		c.a.cache.Remove(c.ids)
		return nil, errors.New("post is hide")
	}
	navi, err := c.dao.GetNavis()
	if err != nil {
		return nil, err
	}
	c.text, err = c.a.build(post, navi)
	if err != nil {
		logging.Error("build article cache: %s", err)
		return nil, err
	}
	c.etag = fmt.Sprintf("%v", time.Now().Unix())
	return nil, nil
}

func (c *article) MaxSize() int64 {
	return c.limit
}

func (c *article) Data() []byte {
	return c.text
}

type Article struct {
	model model.Model
	limit int64 // maximum article size which can be cached
	cfg   *conf.Config
	dao   *dbutil.Dao
	cache *routers.LRU
}

func NewArticle(cfg *conf.Config, dao *dbutil.Dao, r *routers.Router, middleware ...routers.IHandler) *Article {
	a := &Article{
		model: model.NewDefaultModel(path.Join(cfg.Model.TmplRoot, cfg.Model.Article.Tmpl)),
		limit: 2 << 20,
		cfg:   cfg,
		dao:   dao,
		cache: routers.NewLRU(100, 20<<20), // 100 items,20MB
	}

	r.GET(cfg.Model.Article.Api, a)
	r.PUT(cfg.Model.Article.Api, a)

	return a
}

func (a *Article) markup(args interface{}) template.HTML {
	s := blackfriday.Run([]byte(args.(string)), blackfriday.WithExtensions(blackfriday.CommonExtensions))
	return template.HTML(s)
}

func (a *Article) build(post *model.ArticleData, navi *model.NaviData) ([]byte, error) {
	a.model.Funcs(template.FuncMap{"markup": a.markup})
	return a.model.Parse(map[string]interface{}{
		"Post":  post,
		"Navis": navi,
	})
}

func (a *Article) prepareUpdate(c *routers.Context) {
	id := c.GetParam("id")
	if id == "" {
		c.Html(http.StatusNotFound, []byte("not found"))
		return
	}
	data := a.cache.Get(id)
	includingHide := false
	if data == nil {
		iid, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			c.Text(http.StatusBadRequest, "invalid id")
			return
		}
		data = &article{dao: a.dao, id: iid, ids: id, limit: a.limit, a: a}
		ss := c.StartSession()
		if ss != nil && ss.Get(a.cfg.Session.AuthKey) == a.cfg.Session.AuthVal {
			includingHide = true
		}
		_, err = data.Update(includingHide)
		if err != nil {
			c.Html(http.StatusNotFound, []byte("not found"))
			return
		}
		a.cache.Add(data)
	}
	data.Write(c)
	if includingHide {
		a.cache.Remove(data.Name())
	}
}

func (a *Article) incrViewCount(c *routers.Context) {
	idStr := c.GetParam("id")

	if idStr == "" {
		c.Html(http.StatusNotFound, []byte("not found"))
		return
	}
	id, err := strconv.ParseInt(idStr, 0, 64)
	if err != nil || id < 1 {
		c.Text(http.StatusBadRequest, "invalid id")
	}
	err, count := a.dao.IncrViewCount(int(id))
	if err != nil {
		c.Text(http.StatusInternalServerError, "action failed")
	} else {
		c.Text(http.StatusOK, strconv.FormatInt(int64(*count), 10))
	}
}

func (a *Article) Serve(c *routers.Context) {
	switch c.Req.Method {
	case "GET":
		a.prepareUpdate(c)
	//case "PUT":
	//	a.incrViewCount(c)
	default:
		c.Json(http.StatusBadRequest, newError(-1, "unsupport request method"))
	}

}

func (a *Article) Update(arg interface{}) {
	if arg == nil {
		a.cache.UpdateAll()
	} else {
		c := a.cache.Get(arg.(string))
		if c == nil {
			return // not visited yet
		}
		a.cache.Update(c)
	}
}
