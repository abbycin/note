/***********************************************
        File Name: article
        Author: Abby Cin
        Mail: abbytsing@gmail.com
        Created Time: 10/13/19 11:29 AM
***********************************************/

package model

import (
	"time"
)

type ArticleData struct {
	Status
	Id           int       `json:"id"`
	CreateTime   time.Time `json:"create_time"`
	LastModified time.Time `json:"last_modified"`
	Title        string    `json:"title"`
	Content      string    `json:"content"`
	Tags         string    `json:"tags"`   // present as `,` split string
	Images       string    `json:"images"` // present as json string, where key is link, value is name
	Hide         bool      `json:"hide"`
}
