/***********************************************
        File Name: interface
        Author: Abby Cin
        Mail: abbytsing@gmail.com
        Created Time: 10/13/19 11:32 AM
***********************************************/

package model

import (
	"bufio"
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"path"
)

type Status struct {
	Code  int    `json:"code"`
	Error string `json:"error"`
}

type Model interface {
	Id() int64
	Parse(data interface{}) ([]byte, error)
	Reparse(data interface{}) ([]byte, error)
	Funcs(funcMap template.FuncMap)
}

type DefaultModel struct {
	id       int64
	tmplPath string
	tmplText []byte
	funcMap  template.FuncMap
}

func NewDefaultModel(tmplPath string) *DefaultModel {
	r := &DefaultModel{
		id:       newId(),
		tmplPath: tmplPath,
	}
	var e error = nil
	r.tmplText, e = ioutil.ReadFile(tmplPath)
	if e != nil {
		panic(fmt.Sprintf("%s while reading file: %s\n", e, r.tmplPath))
	}
	return r
}

func (l *DefaultModel) Id() int64 {
	return l.id
}

func (l *DefaultModel) Parse(data interface{}) ([]byte, error) {
	tmp := template.New(path.Base(l.tmplPath)) // fuck golang, name must be base name of template file
	var err error = nil
	if len(l.funcMap) != 0 {
		tmp.Funcs(l.funcMap)
	}
	tmp, err = tmp.Parse(string(l.tmplText))
	if err != nil {
		return nil, err
	}
	h := bytes.Buffer{}
	b := bufio.NewWriter(&h)
	err = tmp.Execute(b, data)
	if err != nil {
		return nil, err
	}
	b.Flush()
	return h.Bytes(), err
}

func (l *DefaultModel) Reparse(data interface{}) ([]byte, error) {
	var err error = nil
	l.tmplText, err = ioutil.ReadFile(l.tmplPath)
	if err != nil {
		return nil, err
	}
	return l.Parse(data)
}

func (l *DefaultModel) Funcs(funcMap template.FuncMap) {
	l.funcMap = funcMap
}

var gId int64

func init() {
	gId = 0
}

func newId() int64 {
	r := gId
	gId++
	return r
}

func lastId() int64 {
	return gId
}
