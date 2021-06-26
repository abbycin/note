/***********************************************
        File Name: dbutil
        Author: Abby Cin
        Mail: abbytsing@gmail.com
        Created Time: 9/15/19 1:46 PM
***********************************************/

package dbutil

import (
	"blog/logging"
	"blog/model"
	"crypto/md5"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"strings"
	"time"
)

type Dao struct {
	db *sql.DB
}

func NewDao(dbFile string) *Dao {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		panic(err)
	}
	r := &Dao{
		db: db,
	}

	_, err = db.Exec(`create table if not exists posts(
    	id integer primary key,
    	create_time timestamp,
    	last_modified timestamp,
    	title varchar(50),
    	content text,
    	tags text,
    	images text,
    	hide boolean,
    	view_count integer default 0
    	)`)

	if err != nil {
		panic(err)
	}

	_, err = db.Exec(`create table if not exists users(
    	id integer primary key,
    	username varchar(20) unique,
    	password varchar(33)
		)`)

	if err != nil {
		panic(err)
	}

	_, err = db.Exec(`create table if not exists navis(
    	id integer primary key,
    	sequence int,
    	name varchar(20) unique,
    	target text unique
	)`)

	if err != nil {
		panic(err)
	}

	return r
}

func (d *Dao) Close() {
	d.db.Close()
}

func (d *Dao) GetArticle(id int64, includeHide bool) (*model.ArticleData, error) {
	q := "select * from posts where id = ?"
	if !includeHide {
		q = "select * from posts where id = ? and hide = 0"
	}
	res, err := d.db.Query(q, id)
	if err != nil {
		return nil, err
	}
	var data model.ArticleData

	count := 0
	for res.Next() {
		count += 1
		err = res.Scan(&data.Id, &data.CreateTime, &data.LastModified,
			&data.Title, &data.Content, &data.Tags, &data.Images, &data.Hide, &data.ViewCount)

		if err != nil {
			return nil, err
		}
	}

	if count == 0 {
		return nil, nil
	}
	return &data, nil
}

func (d *Dao) UpdateArticle(id int, data *model.ArticleData) error {
	stmt, err := d.db.Prepare(`update posts set last_modified = ?,
		title = ?, content = ?, tags = ?, images = ? where id = ?`)
	if err != nil {
		return err
	}

	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Stmt(stmt).Exec(time.Now(), data.Title, data.Content, data.Tags, data.Images, id)
	if err != nil {
		tx.Rollback()
	} else {
		tx.Commit()
	}
	return err
}

func (d *Dao) IncrViewCount(id int) (error, *int) {
	stmt, err := d.db.Prepare(`update posts set view_count = view_count + 1 where id = ?`)
	if err != nil {
		logging.Error("%s", err)
		return err, nil
	}

	tx, err := d.db.Begin()
	if err != nil {
		logging.Error("%s", err)
		return err, nil
	}

	_, err = tx.Stmt(stmt).Exec(id)
	if err != nil {
		logging.Error("%s", err)
		tx.Rollback()
		return err, nil
	} else {
		tx.Commit()
	}

	r, err := d.db.Query("select view_count from posts where id = ?", id)
	if err != nil {
		logging.Error("%s", err)
		return err, nil
	}
	res := 0
	for r.Next() {
		err = r.Scan(&res)
		if err != nil {
			logging.Error("%s", err)
			return err, nil
		}
	}
	return nil, &res
}

func (d *Dao) NewArticle(data *model.ArticleData) error {
	stmt, err := d.db.Prepare(`insert into posts(create_time, last_modified,
                  title, content, tags, images, hide, view_count) values(?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}

	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	t := time.Now()
	_, err = tx.Stmt(stmt).Exec(t, t, data.Title, data.Content, data.Tags, data.Images, true, 0)
	if err != nil {
		tx.Rollback()
	} else {
		tx.Commit()
	}
	return err
}

func (d *Dao) DelArticle(id int) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	_, err = tx.Exec(`delete from posts where id = ?`, id)
	if err != nil {
		tx.Rollback()
	} else {
		tx.Commit()
	}
	return err
}

func (d *Dao) HideArticle(id int, hide bool) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	_, err = tx.Exec(`update posts set hide = ? where id = ?`, hide, id)
	if err != nil {
		tx.Rollback()
	} else {
		tx.Commit()
	}
	return err
}

func (d *Dao) GetPosts() (*model.ManageData, error) {
	rows, err := d.db.Query(`select id, title, create_time, last_modified, tags, hide from posts order by create_time desc`)
	if err != nil {
		return nil, err
	}

	posts := make([]model.Post, 0)
	for rows.Next() {
		post := model.Post{}
		tags := ""
		err = rows.Scan(&post.Id, &post.Title, &post.CreateTime, &post.LastModified, &tags, &post.Hidden)
		if err != nil {
			return nil, err
		}
		post.Tags = strings.Split(tags, ",")
		posts = append(posts, post)
	}
	res := &model.ManageData{
		Posts: posts,
	}
	return res, nil
}

func (d *Dao) UserLogin(id, pass string) error {
	epass := fmt.Sprintf("%x", md5.Sum([]byte(pass)))
	rows, err := d.db.Query("select count(id) from users")
	if err != nil {
		return err
	}
	count := -1
	for rows.Next() {
		rows.Scan(&count)
	}

	// first time login
	if count == 0 {
		tx, err := d.db.Begin()
		if err != nil {
			return err
		}
		_, err = tx.Exec("insert into users(username, password) values(?, ?)", id, epass)
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
		return err
	}

	// validate
	rows, err = d.db.Query("select id from users where username = ? and password = ?", id, epass)
	if err != nil {
		logging.Error("err: %v", err)
		return err
	}

	count = 0
	for rows.Next() {
		count += 1
	}
	if count == 0 {
		return errors.New("invalid id or pass")
	}
	return nil
}

func (d *Dao) UpdateUser(id, pass string) error {
	stmt, err := d.db.Prepare("update users set password = ? where username = ?")
	if err != nil {
		return err
	}
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	epass := fmt.Sprintf("%x", md5.Sum([]byte(pass)))
	_, err = tx.Stmt(stmt).Exec(epass, id)
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

func (d *Dao) GetNavis() (*model.NaviData, error) {
	rows, err := d.db.Query("select * from navis order by sequence")

	if err != nil {
		return nil, err
	}

	navis := make([]model.Navi, 0)
	for rows.Next() {
		var data model.Navi
		err = rows.Scan(&data.Id, &data.Sequence, &data.Name, &data.Target)
		if err != nil {
			return nil, err
		}
		navis = append(navis, data)
	}
	return &model.NaviData{
		Navis: navis,
	}, nil
}

func (d *Dao) UpdateNavi(data *model.Navi) error {
	stmt, err := d.db.Prepare("update navis set sequence = ?, name = ?, target = ? where id = ?")
	if err != nil {
		return err
	}
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	_, err = tx.Stmt(stmt).Exec(data.Sequence, data.Name, data.Target, data.Id)
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

func (d *Dao) NewNavi(data *model.Navi) error {
	stmt, err := d.db.Prepare("insert into navis(sequence, name, target) values(?, ?, ?)")
	if err != nil {
		return err
	}
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	_, err = tx.Stmt(stmt).Exec(data.Sequence, data.Name, data.Target)
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

func (d *Dao) DelNavi(id int64) error {
	stmt, err := d.db.Prepare("delete from navis where id = ?")
	if err != nil {
		return err
	}
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	_, err = tx.Stmt(stmt).Exec(id)
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}
