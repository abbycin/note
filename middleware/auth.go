/***********************************************
        File Name: auth
        Author: Abby Cin
        Mail: abbytsing@gmail.com
        Created Time: 10/31/19 7:52 PM
***********************************************/

package middleware

import (
	"blog/conf"
	"blog/routers"
	"blog/session"
	"net/http"
	"strconv"
)

type Auth struct {
	mgr *session.SessionManager
	cfg *conf.Config
}

func NewAuth(mgr *session.SessionManager, cfg *conf.Config) *Auth {
	return &Auth{
		mgr: mgr,
		cfg: cfg,
	}
}

func (a *Auth) Serve(c *routers.Context) {
	ss := a.mgr.StartSession(c.Resp, c.Req)
	status := ss.Get(a.cfg.Session.AuthKey)

	if status != a.cfg.Session.AuthVal {
		c.Json(http.StatusUnauthorized, routers.J{
			"code":  -1,
			"error": "unauthorized",
		})
		c.Abort()
		return
	}

	if c.Req.Method == "PUT" || c.Req.Method == "POST" {
		n := c.Req.Header.Get("Content-Length")
		r, e := strconv.ParseInt(n, 10, 64)
		if e != nil {
			c.Json(http.StatusBadRequest, routers.J{
				"code":  -1,
				"error": "invaild content length",
			})
			c.Abort()
			return
		}
		if r > (10 << 20) {
			c.Json(http.StatusBadRequest, routers.J{
				"code":  -1,
				"error": "body too large",
			})
			c.Abort()
			return
		}
	}
	c.Next()
}
