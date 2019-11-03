/***********************************************
        File Name: navi
        Author: Abby Cin
        Mail: abbytsing@gmail.com
        Created Time: 10/27/19 3:59 PM
***********************************************/

package service

import (
	"blog/conf"
	"blog/dbutil"
	"blog/logging"
	"blog/model"
	"blog/routers"
	"net/http"
	"strconv"
)

type Navi struct {
	dao   *dbutil.Dao
	cache []ICache
}

func NewNavi(cfg *conf.Config, dao *dbutil.Dao, r *routers.Router, middleware ...routers.IHandler) *Navi {
	n := &Navi{dao: dao, cache: make([]ICache, 0)}

	r.GET(cfg.Service.Navi, n)
	cb := append(middleware, n)
	r.POST(cfg.Service.Navi, cb...)
	r.PUT(cfg.Service.Navi, cb...)
	r.DELETE(cfg.Service.Navi, cb...)

	return n
}

func (n *Navi) Add(c ICache) {
	n.cache = append(n.cache, c)
}

func (n *Navi) Serve(c *routers.Context) {
	switch c.Req.Method {
	case "GET":
		n.HandleGet(c)
	case "PUT":
		n.HandlePut(c)
	case "POST":
		n.HandlePost(c)
	case "DELETE":
		n.HandleDel(c)
	}
}

func (n *Navi) HandleGet(c *routers.Context) {
	data, err := n.dao.GetNavis()
	if err != nil {
		c.Json(http.StatusBadRequest, newError(-2, err.Error()))
	} else {
		c.Json(http.StatusOK, data)
	}
}

func (n *Navi) HandlePut(c *routers.Context) {
	var data model.Navi
	err := c.Unmarshal(&data)
	if err != nil {
		c.Json(http.StatusBadRequest, newError(-1, "invalid request"))
		return
	}

	err = n.dao.UpdateNavi(&data)
	if err != nil {
		c.Json(http.StatusBadRequest, newError(-2, err.Error()))
	} else {
		n.Update()
		c.Json(http.StatusOK, newError(0, ""))
	}
}

func (n *Navi) HandlePost(c *routers.Context) {
	var data model.Navi
	err := c.Unmarshal(&data)
	if err != nil {
		logging.Error("unmarshal: %s", err)
		c.Json(http.StatusBadRequest, newError(-1, "invalid request"))
		return
	}
	err = n.dao.NewNavi(&data)
	if err != nil {
		c.Json(http.StatusBadRequest, newError(-2, err.Error()))
	} else {
		n.Update()
		c.Json(http.StatusOK, newError(0, ""))
	}
}

func (n *Navi) HandleDel(c *routers.Context) {
	id := c.Req.URL.Query().Get("id")

	if id == "" {
		c.Json(http.StatusBadRequest, newError(-1, "id is not given"))
		return
	}
	iid, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.Json(http.StatusBadRequest, newError(-1, "invalid id"))
		return
	}
	err = n.dao.DelNavi(iid)

	if err != nil {
		c.Json(http.StatusBadRequest, newError(-2, err.Error()))
	} else {
		n.Update()
		c.Json(http.StatusBadRequest, newError(0, ""))
	}
}

func (n *Navi) Update() {
	for _, cache := range n.cache {
		cache.Update(nil)
	}
}
