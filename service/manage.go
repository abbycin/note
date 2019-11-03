/***********************************************
        File Name: manage
        Author: Abby Cin
        Mail: abbytsing@gmail.com
        Created Time: 10/23/19 8:08 PM
***********************************************/

package service

import (
	"blog/conf"
	"blog/dbutil"
	"blog/model"
	"blog/routers"
	"net/http"
)

type Manage struct {
	model *model.Manage
	dao   *dbutil.Dao
}

func NewManage(cfg *conf.Config, dao *dbutil.Dao, r *routers.Router, middleware ...routers.IHandler) *Manage {
	m := &Manage{
		model: &model.Manage{},
		dao:   dao,
	}

	r.GET(cfg.Service.Manage, append(middleware, m)...)
	m.model.Init(cfg, r, middleware...)

	return m
}

func (m *Manage) Serve(c *routers.Context) {
	posts, err := m.dao.GetPosts()
	if err != nil {
		c.Json(http.StatusInternalServerError, newError(-2, err.Error()))
		return
	} else {
		c.Json(http.StatusOK, posts)
	}
}
