/***********************************************
        File Name: user
        Author: Abby Cin
        Mail: abbytsing@gmail.com
        Created Time: 10/27/19 4:44 PM
***********************************************/

package service

import (
	"blog/conf"
	"blog/dbutil"
	"blog/routers"
	"io/ioutil"
	"net/http"
)

type User struct {
	dao *dbutil.Dao
}

func NewUser(cfg *conf.Config, dao *dbutil.Dao, r *routers.Router, middleware ...routers.IHandler) *User {
	u := &User{dao: dao}

	r.PUT(cfg.Service.User, append(middleware, u)...)

	return u
}

func (u *User) Serve(c *routers.Context) {
	data, err := ioutil.ReadAll(c.Req.Body)
	if err != nil {
		c.Json(http.StatusBadRequest, newError(-1, "invalid request"))
		return
	}

	ss := c.StartSession()

	err = u.dao.UpdateUser(ss.Get(ss.Id()), string(data))

	if err != nil {
		c.Json(http.StatusBadRequest, newError(-2, err.Error()))
	} else {
		c.ClearSession()
		c.Json(http.StatusOK, newError(0, ""))
	}
}
