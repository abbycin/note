/***********************************************
        File Name: article
        Author: Abby Cin
        Mail: abbytsing@gmail.com
        Created Time: 10/20/19 4:23 PM
***********************************************/

package service

import (
	"blog/conf"
	"blog/dbutil"
	"blog/model"
	"blog/routers"
	"net/http"
	"strconv"
)

type Editor struct {
	cfg   *conf.Config
	dao   *dbutil.Dao
	model *model.Editor
	cache []ICache
}

func NewEditor(cfg *conf.Config, dao *dbutil.Dao, r *routers.Router, middleware ...routers.IHandler) *Editor {
	this := &Editor{
		cfg:   cfg,
		dao:   dao,
		model: &model.Editor{},
		cache: make([]ICache, 0),
	}

	// json rpc
	handlers := append(middleware, this)
	api := this.cfg.Service.Edit
	r.GET(api, handlers...)
	r.PUT(api, handlers...)
	r.POST(api, handlers...)
	r.DELETE(api, handlers...)

	// generate template, return html
	this.model.Init(cfg, r, middleware...)
	return this
}

func (a *Editor) Add(c ICache) {
	a.cache = append(a.cache, c)
}

func (a *Editor) Serve(c *routers.Context) {
	switch c.Req.Method {
	case "GET":
		a.Get(c)
	case "PUT":
		a.Put(c)
	case "POST":
		a.Post(c)
	case "DELETE":
		a.Del(c)
	default:
		c.Json(http.StatusBadRequest, newError(-1, "unsupport request method"))
	}
}

func (a *Editor) Get(c *routers.Context) {
	id, ok := a.Validate(c)

	if !ok {
		return
	}

	data, e := a.dao.GetArticle(int64(id), true)
	if e != nil {
		c.Json(http.StatusBadRequest, newError(-2, e.Error()))
	} else {
		c.Json(http.StatusOK, data)
	}
}

func (a *Editor) Put(c *routers.Context) {
	id, ok := a.Validate(c)

	if !ok {
		return
	}

	h := c.Req.URL.Query().Get("hide") // set in manage.html
	if h != "" {
		if h != "true" && h != "false" {
			c.Json(http.StatusBadRequest, newError(-1, "must set true or false"))
		} else {
			hide := false
			if h == "true" {
				hide = true
			}
			err := a.dao.HideArticle(id, hide)
			if err != nil {
				c.Json(http.StatusBadRequest, newError(-1, err.Error()))
			} else {
				a.Update(c.Req.URL.Query().Get("id"))
				c.Json(http.StatusOK, newError(0, ""))
			}
		}
		return
	}

	var data model.ArticleData
	err := c.Unmarshal(&data)
	if err != nil {
		c.Json(http.StatusBadRequest, newError(-1, "invalid data"))
		return
	}

	err = a.dao.UpdateArticle(id, &data)

	if err != nil {
		c.Json(http.StatusBadRequest, newError(-2, err.Error()))
	} else {
		a.Update(c.Req.URL.Query().Get("id"))
		c.Json(http.StatusOK, newError(0, ""))
	}
}

func (a *Editor) Post(c *routers.Context) {
	var data model.ArticleData
	err := c.Unmarshal(&data)
	if err != nil {
		c.Json(http.StatusBadRequest, newError(-1, "invalid data"))
		return
	}
	err = a.dao.NewArticle(&data)
	if err != nil {
		c.Json(http.StatusBadRequest, newError(-2, err.Error()))
	} else {
		c.Json(http.StatusOK, newError(0, ""))
	}
}

func (a *Editor) Del(c *routers.Context) {
	id, ok := a.Validate(c)

	if !ok {
		return
	}

	err := a.dao.DelArticle(id)
	if err != nil {
		c.Json(http.StatusBadRequest, newError(-2, err.Error()))
	} else {
		a.Update(c.Req.URL.Query().Get("id"))
		c.Json(http.StatusOK, newError(0, ""))
	}
}

func (a *Editor) Validate(c *routers.Context) (int, bool) {
	id := c.Req.URL.Query().Get("id")

	if id == "" {
		c.Json(http.StatusBadRequest, newError(-1, "no id found"))
		return -1, false
	}
	rid, err := strconv.ParseInt(id, 10, 63)
	if err != nil {
		c.Json(http.StatusBadRequest, newError(-1, "id is invalid"))
		return -1, false
	}
	return int(rid), true
}

func (a *Editor) Update(id string) {
	for _, cache := range a.cache {
		cache.Update(id)
	}
}
