/***********************************************
        File Name: access
        Author: Abby Cin
        Mail: abbytsing@gmail.com
        Created Time: 11/1/19 7:11 PM
***********************************************/

package middleware

import (
	"blog/logging"
	"blog/routers"
)

type Access struct {
}

func NewAccess() *Access {
	return &Access{}
}

func (a *Access) Serve(c *routers.Context) {
	logging.Info("ACCESS: %s %s %s", c.Req.RemoteAddr, c.Req.Method, c.Req.URL.String())
}
