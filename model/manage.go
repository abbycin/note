/***********************************************
        File Name: manage
        Author: Abby Cin
        Mail: abbytsing@gmail.com
        Created Time: 10/13/19 11:29 AM
***********************************************/

package model

import (
	"blog/conf"
	"blog/routers"
	"fmt"
	"log"
	"net/http"
	"path"
	"time"
)

type JSONTime time.Time

func (t JSONTime) MarshalJSON() ([]byte, error) {
	stamp := fmt.Sprintf(`"%s"`, time.Time(t).Format("2006 01-02 15:04:05"))
	return []byte(stamp), nil
}

// no draft status, post can be viewed by authorized user even Hidden is true
type Post struct {
	Id           int      `json:"id"`
	Title        string   `json:"title"` // a link to post, maybe /post?id=1
	Tags         []string `json:"tags"`
	Hidden       bool     `json:"hide"`
	CreateTime   JSONTime `json:"create_time"`
	LastModified JSONTime `json:"last_modified"`
}

type ManageData struct {
	Status
	Posts []Post `json:"posts"`
}

type Navi struct {
	Id       int    `json:"id"`
	Sequence int    `json:"sequence"`
	Name     string `json:"name"`
	Target   string `json:"target"`
}

type NaviData struct {
	Status
	Navis []Navi `json:"navis"`
}

type ManageModel struct {
	Title string
}

type SettingModel struct {
	Title string
}

type Manage struct {
	mainModel    Model
	settingModel Model
	mainText     []byte
	settingText  []byte
}

func (m *Manage) Init(cfg *conf.Config, r *routers.Router, middleware ...routers.IHandler) *Manage {
	m.mainModel = NewDefaultModel(path.Join(cfg.Model.TmplRoot, cfg.Model.Manage.Main))
	m.settingModel = NewDefaultModel(path.Join(cfg.Model.TmplRoot, cfg.Model.Manage.Setting))

	data1 := ManageModel{
		Title: cfg.Model.Title,
	}

	var err error = nil
	m.mainText, err = m.mainModel.Parse(data1)
	if err != nil {
		log.Panicf("parse %s failed: %s\n", cfg.Model.Manage.Main, err)
	}

	data2 := SettingModel{
		Title: cfg.Model.Title,
	}

	m.settingText, err = m.settingModel.Parse(data2)
	if err != nil {
		log.Panicf("parse %s failed: %s\n", cfg.Model.Manage.Setting, err)
	}

	r.GET(cfg.Model.Manage.MainApi, append(middleware, m)...)
	r.GET(cfg.Model.Manage.SubApi, append(middleware, m)...)

	return m
}

func (m *Manage) Serve(c *routers.Context) {
	node := c.GetParam("sub")
	switch node {
	case "setting":
		c.Html(http.StatusOK, m.settingText)
	default:
		c.Html(http.StatusOK, m.mainText)
	}
}
