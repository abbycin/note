/***********************************************
        File Name: session
        Author: Abby Cin
        Mail: abbytsing@gmail.com
        Created Time: 10/13/19 2:36 PM
***********************************************/

package session

import (
	"blog/conf"
	"blog/logging"
	"container/list"
	"crypto/md5"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

var gId int64

func init() {
	gId = 0
}

func newId() int64 {
	r := gId
	gId++
	return r
}

func buildId() string {
	tmp := md5.Sum([]byte(strconv.FormatInt(newId(), 10)))
	return fmt.Sprintf("%x", tmp)
}

type ISession interface {
	Set(key, val string) error
	Get(key string) string
	Del(key string) error
	Id() string
}

type ISessionBackend interface {
	Init(id string) (ISession, error)
	Get(id string) (ISession, error)
	Del(id string) error
	Update(id string)
	Clean(maxLifetime time.Duration)
}

type SessionManager struct {
	mtx     sync.Mutex
	name    string
	maxAge  time.Duration
	backend ISessionBackend
}

// fuck golang, no generics
type DefaultBackend struct {
	mtx      sync.Mutex
	sessions map[string]*list.Element
	queue    *list.List
}

type DefaultSession struct {
	id       string
	cookie   map[string]string
	lastTime time.Time
	backend  ISessionBackend
}

func NewDefaultBackend() *DefaultBackend {
	return &DefaultBackend{
		mtx:      sync.Mutex{},
		sessions: make(map[string]*list.Element),
		queue:    list.New(),
	}
}

func (b *DefaultBackend) Init(id string) (ISession, error) {
	b.mtx.Lock()
	defer b.mtx.Unlock()

	r := &DefaultSession{
		id:       id,
		cookie:   make(map[string]string),
		lastTime: time.Now(),
		backend:  b,
	}
	e := b.queue.PushBack(r)
	b.sessions[id] = e

	return r, nil
}

func (b *DefaultBackend) Update(id string) {
	b.mtx.Lock()
	defer b.mtx.Unlock()

	if e, ok := b.sessions[id]; ok {
		e.Value.(*DefaultSession).lastTime = time.Now()
		b.queue.MoveToFront(e)
	}
}

func (b *DefaultBackend) Get(id string) (ISession, error) {
	b.mtx.Lock()
	defer b.mtx.Unlock()

	e, ok := b.sessions[id]
	if !ok {
		return nil, errors.New("no such session")
	}
	return e.Value.(*DefaultSession), nil
}

func (b *DefaultBackend) Del(id string) error {
	b.mtx.Lock()
	defer b.mtx.Unlock()

	if e, ok := b.sessions[id]; ok {
		delete(b.sessions, id)
		b.queue.Remove(e)
	}
	return nil
}

func (b *DefaultBackend) Clean(exp time.Duration) {
	b.mtx.Lock()
	defer b.mtx.Unlock()

	cur := time.Now()
	for {
		e := b.queue.Back()
		if e == nil {
			return
		}
		if s := e.Value.(*DefaultSession); cur.Sub(s.lastTime) > exp {
			b.queue.Remove(e)
			delete(b.sessions, s.Id())
		} else {
			break
		}
	}
}

func (s *DefaultSession) Id() string {
	return s.id
}

func (s *DefaultSession) Set(k, v string) error {
	s.cookie[k] = v
	s.backend.Update(s.Id())
	return nil
}

func (s *DefaultSession) Get(k string) string {
	if e, ok := s.cookie[k]; ok {
		s.backend.Update(s.Id())
		return e
	}
	return ""
}

func (s *DefaultSession) Del(k string) error {
	if _, ok := s.cookie[k]; ok {
		delete(s.cookie, k)
		s.backend.Update(s.Id())
	}
	return nil
}

func NewSessionManager(cfg *conf.Config) *SessionManager {
	return &SessionManager{
		mtx:     sync.Mutex{},
		name:    cfg.Session.Key,
		maxAge:  time.Duration(cfg.Session.Expiry * int(time.Second)),
		backend: NewDefaultBackend(),
	}
}

func (m *SessionManager) newSession(w http.ResponseWriter, r *http.Request) (ISession, error) {
	sid := buildId()
	session, e := m.backend.Init(sid)
	if e != nil {
		logging.Info("SessionManager::newSession: %s", e.Error())
	}
	cookie := http.Cookie{Name: m.name, Value: url.QueryEscape(sid), Path: "/", HttpOnly: true, MaxAge: int(m.maxAge)}
	http.SetCookie(w, &cookie)
	return session, e
}

func (m *SessionManager) StartSession(w http.ResponseWriter, r *http.Request) (session ISession) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	cookie, e := r.Cookie(m.name)

	if e != nil || cookie.Value == "" {
		session, _ = m.newSession(w, r)
	} else {
		sid, e := url.QueryUnescape(cookie.Value)
		if e != nil {
			logging.Error("unescape cookie: %s", e.Error())
		}
		session, e = m.backend.Get(sid)
		if e != nil {
			logging.Info("session id is not exists(maybe a previous session), create a new one")
			session, _ = m.newSession(w, r)
		}
	}
	return
}

func (m *SessionManager) SessionDestroy(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(m.name)
	if err != nil || cookie.Value == "" {
		return
	} else {
		m.mtx.Lock()
		defer m.mtx.Unlock()
		m.backend.Del(cookie.Value)
		expiration := time.Now()
		cookie := http.Cookie{Name: m.name, Path: "/", HttpOnly: true, Expires: expiration, MaxAge: -1}
		http.SetCookie(w, &cookie)
	}
}

func (m *SessionManager) Clear() {
	m.backend.Clean(0)
}

func (m *SessionManager) StartGC() {
	go m.GC()
}

func (m *SessionManager) GC() {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	m.backend.Clean(m.maxAge)
	time.AfterFunc(m.maxAge, func() { m.GC() })
}
