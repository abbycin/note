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
	"regexp"
	"strings"
	"time"
)

type Dao struct {
	db *sql.DB
}

// sanitizeTag 清理标签输入，防止注入攻击
func sanitizeTag(tag string) string {
	// 移除特殊字符，只允许字母、数字、中文、空格和常见符号
	re := regexp.MustCompile(`[^\w\s\-\u4e00-\u9fa5]`)
	tag = re.ReplaceAllString(tag, "")
	// 限制长度
	if len(tag) > 50 {
		tag = tag[:50]
	}
	return strings.TrimSpace(tag)
}

// sanitizeInput 清理一般输入
func sanitizeInput(input string, maxLen int) string {
	if len(input) > maxLen {
		input = input[:maxLen]
	}
	// 移除 null 字节和危险字符
	input = strings.ReplaceAll(input, "\x00", "")
	input = strings.ReplaceAll(input, "\x1a", "")
	return input
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
    	title varchar(200),
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
	return d.GetPostsByPage(0, 0)
}

func (d *Dao) GetPostsByPage(page, pageSize int) (*model.ManageData, error) {
	var rows *sql.Rows
	var err error

	if pageSize > 0 {
		offset := page * pageSize
		rows, err = d.db.Query(`select id, title, create_time, last_modified, tags, hide from posts where hide = 0 order by create_time desc limit ? offset ?`, pageSize, offset)
	} else {
		rows, err = d.db.Query(`select id, title, create_time, last_modified, tags, hide from posts where hide = 0 order by create_time desc`)
	}

	if err != nil {
		return nil, err
	}

	posts := make([]model.PostInfo, 0)
	for rows.Next() {
		post := model.PostInfo{}
		tags := ""
		err = rows.Scan(&post.Id, &post.Title, &post.CreateTime, &post.LastModified, &tags, &post.Hidden)
		if err != nil {
			return nil, err
		}
		rawTags := strings.Split(tags, ",")
		post.Tags = make([]string, 0, len(rawTags))
		for _, tag := range rawTags {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				post.Tags = append(post.Tags, tag)
			}
		}
		posts = append(posts, post)
	}
	res := &model.ManageData{
		Posts: posts,
	}
	return res, nil
}

func (d *Dao) GetPostsCount() (int, error) {
	row := d.db.QueryRow("select count(*) from posts where hide = 0")
	var count int
	err := row.Scan(&count)
	return count, err
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

	navis := make([]model.NaviInfo, 0)
	for rows.Next() {
		var data model.NaviInfo
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

// GetAllTags 获取所有标签
func (d *Dao) GetAllTags() ([]string, error) {
	rows, err := d.db.Query("select tags from posts where hide = 0")
	if err != nil {
		return nil, err
	}

	tagMap := make(map[string]bool)
	for rows.Next() {
		var tags string
		err = rows.Scan(&tags)
		if err != nil {
			continue
		}
		if tags != "" {
			for _, tag := range strings.Split(tags, ",") {
				tag = strings.TrimSpace(tag)
				if tag != "" {
					tagMap[tag] = true
				}
			}
		}
	}

	// 转换为切片并排序
	result := make([]string, 0, len(tagMap))
	for tag := range tagMap {
		result = append(result, tag)
	}
	// 简单排序
	for i := 0; i < len(result); i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i] > result[j] {
				result[i], result[j] = result[j], result[i]
			}
		}
	}
	return result, nil
}

// GetPostsByTag 根据标签获取文章
func (d *Dao) GetPostsByTag(tag string) (*model.ManageData, error) {
	rows, err := d.db.Query(`select id, title, create_time, last_modified, tags, hide from posts 
		where hide = 0 and tags like ? order by create_time desc`, "%"+tag+"%")
	if err != nil {
		return nil, err
	}

	posts := make([]model.PostInfo, 0)
	for rows.Next() {
		post := model.PostInfo{}
		tags := ""
		err = rows.Scan(&post.Id, &post.Title, &post.CreateTime, &post.LastModified, &tags, &post.Hidden)
		if err != nil {
			return nil, err
		}
		// 检查标签是否匹配（避免部分匹配）
		rawTags := strings.Split(tags, ",")
		postTags := make([]string, 0, len(rawTags))
		for _, t := range rawTags {
			t = strings.TrimSpace(t)
			if t != "" {
				postTags = append(postTags, t)
			}
		}
		for _, t := range postTags {
			if t == tag {
				post.Tags = postTags
				posts = append(posts, post)
				break
			}
		}
	}
	res := &model.ManageData{
		Posts: posts,
	}
	return res, nil
}

func (d *Dao) UpdateNavi(data *model.NaviInfo) error {
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

func (d *Dao) NewNavi(data *model.NaviInfo) error {
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
