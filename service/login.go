/***********************************************
        File Name: login
        Author: Abby Cin
        Mail: abbytsing@gmail.com
        Created Time: 10/21/19 9:02 PM
***********************************************/

package service

import (
	"blog/conf"
	"blog/dbutil"
	"blog/logging"
	"blog/model"
	"blog/routers"
	"net/http"
	"sync"
	"time"
)

type blacklist struct {
	white  map[string]int
	black  map[string]time.Time
	limit  int
	expiry int64
	ch     chan struct{}
	timer  *time.Ticker // fuck golang, expiry will shift
	mtx    sync.Mutex
}

func (b *blacklist) valid(host string) bool {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	c, ok := b.white[host]
	if ok {
		return c < b.limit
	}
	return true
}

func (b *blacklist) count(host string) {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	c, ok := b.white[host]
	if ok {
		c += 1
		if c >= b.limit {
			logging.Warn("blacklist: %s", host)
			b.black[host] = time.Now().Add(time.Duration(b.expiry))
		}
	} else {
		c = 1
	}
	b.white[host] = c
}

func (b *blacklist) check() {
outer:
	for {
		select {
		case <-b.ch:
			break outer
		case <-b.timer.C:
			b.mtx.Lock()
			now := time.Now()
			for k, v := range b.black {
				if now.Sub(v) >= 0 {
					delete(b.black, k)
					delete(b.white, k)
				}
			}
			b.mtx.Unlock() // fuck golang, defer is not work in `case:`
		}
	}
	b.timer.Stop()
	close(b.ch)
}

func (b blacklist) stop() {
	b.ch <- struct{}{}
}

func newBlacklist(expiry int64, limit int) *blacklist {
	b := &blacklist{
		white:  make(map[string]int),
		black:  make(map[string]time.Time),
		limit:  limit,
		expiry: expiry * int64(time.Second),
		ch:     make(chan struct{}),
		timer:  time.NewTicker(time.Duration(expiry * int64(time.Second))),
		mtx:    sync.Mutex{},
	}

	go b.check()
	return b
}

type Login struct {
	cfg   *conf.Config
	dao   *dbutil.Dao
	model *model.Login
	bl    *blacklist
}

func NewLogin(cfg *conf.Config, dao *dbutil.Dao, r *routers.Router) *Login {
	l := &Login{
		cfg:   cfg,
		dao:   dao,
		model: &model.Login{},
		bl:    newBlacklist(cfg.Blacklist.Expiry, cfg.Blacklist.Limit),
	}

	r.POST(cfg.Service.Login, l)
	r.DELETE(cfg.Service.Login, l)
	l.model.Init(cfg, r)

	return l
}

func (l *Login) Serve(c *routers.Context) {
	switch c.Req.Method {
	case "POST":
		l.HandlePost(c)
	case "DELETE":
		l.HandleDel(c)
	default:
		c.Json(http.StatusNotFound, newError(-1, "not found"))
	}
}

func (l *Login) HandlePost(c *routers.Context) {
	remoteIp := c.RemoteAddr()
	if !l.bl.valid(remoteIp) {
		logging.Warn("too many attempts: %s", remoteIp)
		c.Json(http.StatusBadRequest, newError(-1, "fuck you!"))
		return
	}
	ss := c.StartSession()
	status := ss.Get(l.cfg.Session.AuthKey)

	if status != l.cfg.Session.AuthVal {
		var data model.LoginData
		err := c.Unmarshal(&data)
		if err != nil {
			c.Json(http.StatusBadRequest, newError(-1, "invalid request"))
			return
		}
		err = l.dao.UserLogin(data.Username, data.Password)
		if err == nil {
			ss.Set(l.cfg.Session.AuthKey, l.cfg.Session.AuthVal)
			ss.Set(ss.Id(), data.Username)
			c.Json(http.StatusOK, newError(0, ""))
		} else {
			l.bl.count(remoteIp)
			c.Json(http.StatusForbidden, newError(-1, err.Error()))
		}
	} else {
		l.bl.count(remoteIp)
		c.Json(http.StatusBadRequest, newError(-1, "invalid request"))
	}
}

func (l *Login) HandleDel(c *routers.Context) {
	c.SessionDestroy()
	c.Json(http.StatusOK, newError(0, ""))
}
