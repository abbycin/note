/***********************************************
        File Name: service
        Author: Abby Cin
        Mail: abbytsing@gmail.com
        Created Time: 10/26/19 1:21 PM
***********************************************/

package service

type Error struct {
	Code  int    `json:"code"`
	Error string `json:"error"`
}

func newError(code int, msg string) Error {
	return Error{
		Code:  code,
		Error: msg,
	}
}

type ICache interface {
	Update(arg interface{})
}
