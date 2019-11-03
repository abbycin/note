/***********************************************
        File Name: edit
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

type editorModel struct {
	Title template.URL
}

type Editor struct {
	model Model
	text  []byte
}

func (e *Editor) Init(cfg *conf.Config, r *routers.Router, middleware ...routers.IHandler) {
	e.model = NewDefaultModel(path.Join(cfg.Model.TmplRoot, cfg.Model.Edit.Tmpl))
	r.GET(cfg.Model.Edit.Api, append(middleware, e)...)
	data := editorModel{
		Title: template.URL(cfg.Model.Title),
	}

	var err error = nil
	e.text, err = e.model.Parse(data)
	if err != nil {
		log.Panicf("parse %s failed: %s\n", cfg.Model.Edit.Tmpl, err)
	}
}

func (e *Editor) Serve(c *routers.Context) {
	c.Html(http.StatusOK, e.text)
}
