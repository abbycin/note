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

// DBPost is the GORM model for posts table
type DBPost struct {
	Id           int       `json:"id" gorm:"primaryKey;autoIncrement"`
	CreateTime   time.Time `json:"create_time" gorm:"index"`
	LastModified time.Time `json:"last_modified"`
	Title        string    `json:"title" gorm:"size:200"`
	Content      string    `json:"content" gorm:"type:text"`
	Tags         string    `json:"tags" gorm:"type:text"`
	Images       string    `json:"images" gorm:"type:text"`
	Hide         bool      `json:"hide" gorm:"index"`
	ViewCount    int       `json:"view_count" gorm:"default:0"`
}

// TableName specifies the table name for GORM
func (DBPost) TableName() string {
	return "posts"
}

// DBUser is the GORM model for users table
type DBUser struct {
	Id       int    `json:"id" gorm:"primaryKey;autoIncrement"`
	Username string `json:"username" gorm:"size:20;uniqueIndex"`
	Password string `json:"password" gorm:"size:33"`
}

// TableName specifies the table name for GORM
func (DBUser) TableName() string {
	return "users"
}

// DBNavi is the GORM model for navis table
type DBNavi struct {
	Id       int    `json:"id" gorm:"primaryKey;autoIncrement"`
	Sequence int    `json:"sequence"`
	Name     string `json:"name" gorm:"size:20;uniqueIndex"`
	Target   string `json:"target" gorm:"uniqueIndex"`
}

// TableName specifies the table name for GORM
func (DBNavi) TableName() string {
	return "navis"
}

// ArticleData is used for backward compatibility with existing code
type ArticleData struct {
	Status
	Id           int       `json:"id"`
	CreateTime   time.Time `json:"create_time"`
	LastModified time.Time `json:"last_modified"`
	Title        string    `json:"title"`
	Content      string    `json:"content"`
	Tags         string    `json:"tags"`
	Images       string    `json:"images"`
	Hide         bool      `json:"hide"`
	ViewCount    int       `json:"view_count"`
}

// ToDBPost converts ArticleData to GORM DBPost model
func (a *ArticleData) ToDBPost() *DBPost {
	return &DBPost{
		Id:           a.Id,
		CreateTime:   a.CreateTime,
		LastModified: a.LastModified,
		Title:        a.Title,
		Content:      a.Content,
		Tags:         a.Tags,
		Images:       a.Images,
		Hide:         a.Hide,
		ViewCount:    a.ViewCount,
	}
}

// FromDBPost converts GORM DBPost model to ArticleData
func (a *ArticleData) FromDBPost(p *DBPost) {
	a.Id = p.Id
	a.CreateTime = p.CreateTime
	a.LastModified = p.LastModified
	a.Title = p.Title
	a.Content = p.Content
	a.Tags = p.Tags
	a.Images = p.Images
	a.Hide = p.Hide
	a.ViewCount = p.ViewCount
}
