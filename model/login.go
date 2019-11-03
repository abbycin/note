/***********************************************
        File Name: login.go
        Author: Abby Cin
        Mail: abbytsing@gmail.com
        Created Time: 10/13/19 11:29 AM
***********************************************/

package model

import (
	"blog/conf"
	"blog/routers"
	"html/template"
	"log"
	"net/http"
	"path"
)

type LoginModel struct {
	Title string
	Login template.URL
}

type LoginData struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Login struct {
	model Model
	cfg   *conf.Config
	text  []byte
}

func (l *Login) Init(cfg *conf.Config, r *routers.Router) {
	l.cfg = cfg
	l.model = NewDefaultModel(path.Join(cfg.Model.TmplRoot, cfg.Model.Login.Tmpl))
	data := LoginModel{
		Title: cfg.Model.Title,
		Login: template.URL(cfg.Service.Login),
	}
	var err error = nil
	l.text, err = l.model.Parse(data)
	if err != nil {
		log.Panicf("parse %s faild: %s\n", err)
	}
	r.GET(cfg.Model.Login.Api, l)
}

func (l Login) Serve(c *routers.Context) {
	switch c.Req.Method {
	case "GET":
		l.HandleGet(c)
	default:
		c.Json(http.StatusNotFound, routers.J{
			"code":  -1,
			"error": "not found",
		})
	}
}

func (l *Login) HandleGet(c *routers.Context) {
	session := c.StartSession()
	status := session.Get(l.cfg.Session.AuthKey)

	if status != l.cfg.Session.AuthVal {
		c.Resp.Header().Set("Content-Type", "text/html")
		c.Resp.WriteHeader(http.StatusOK)
		c.Resp.Write(l.text)
	} else {
		http.Redirect(c.Resp, c.Req, l.cfg.Model.Manage.MainApi, http.StatusFound)
	}
}
