/***********************************************
        File Name: gorm_dao
        Author: Abby Cin
        Mail: abbytsing@gmail.com
        Created Time: 2026-02-12
***********************************************/

package dbutil

import (
	"blog/model"
	"crypto/md5"
	"fmt"
	"strings"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// GormDao uses GORM ORM to prevent SQL injection
type GormDao struct {
	db *gorm.DB
}

// NewGormDao creates a new GORM-based DAO instance
func NewGormDao(dbFile string) *GormDao {
	db, err := gorm.Open(sqlite.Open(dbFile), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		panic(err)
	}

	// Auto migrate tables
	db.AutoMigrate(&model.DBPost{}, &model.DBUser{}, &model.DBNavi{})

	return &GormDao{db: db}
}

// Close closes the database connection
func (d *GormDao) Close() error {
	sqlDB, err := d.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// GetArticle retrieves an article by ID
func (d *GormDao) GetArticle(id int64, includeHide bool) (*model.ArticleData, error) {
	var post model.DBPost
	query := d.db.Where("id = ?", id)
	if !includeHide {
		query = query.Where("hide = ?", false)
	}
	result := query.First(&post)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}

	var data model.ArticleData
	data.FromDBPost(&post)
	return &data, nil
}

// UpdateArticle updates an existing article
func (d *GormDao) UpdateArticle(id int, data *model.ArticleData) error {
	post := data.ToDBPost()
	post.LastModified = time.Now()
	return d.db.Model(&model.DBPost{}).Where("id = ?", id).Updates(map[string]interface{}{
		"last_modified": post.LastModified,
		"title":         post.Title,
		"content":       post.Content,
		"tags":          post.Tags,
		"images":        post.Images,
	}).Error
}

// IncrViewCount increments the view count of an article
func (d *GormDao) IncrViewCount(id int) (error, *int) {
	result := d.db.Model(&model.DBPost{}).Where("id = ?", id).UpdateColumn("view_count", gorm.Expr("view_count + ?", 1))
	if result.Error != nil {
		return result.Error, nil
	}

	var post model.DBPost
	if err := d.db.Where("id = ?", id).First(&post).Error; err != nil {
		return err, nil
	}
	return nil, &post.ViewCount
}

// NewArticle creates a new article
func (d *GormDao) NewArticle(data *model.ArticleData) error {
	post := data.ToDBPost()
	post.CreateTime = time.Now()
	post.LastModified = post.CreateTime
	post.Hide = true
	post.ViewCount = 0
	return d.db.Create(post).Error
}

// DelArticle deletes an article by ID
func (d *GormDao) DelArticle(id int) error {
	return d.db.Where("id = ?", id).Delete(&model.DBPost{}).Error
}

// HideArticle toggles the hide status of an article
func (d *GormDao) HideArticle(id int, hide bool) error {
	return d.db.Model(&model.DBPost{}).Where("id = ?", id).Update("hide", hide).Error
}

// GetPosts retrieves all posts (for backward compatibility)
func (d *GormDao) GetPosts() (*model.ManageData, error) {
	return d.GetPostsByPage(0, 0)
}

// GetPostsByPage retrieves posts with pagination
func (d *GormDao) GetPostsByPage(page, pageSize int) (*model.ManageData, error) {
	var posts []model.DBPost
	query := d.db.Where("hide = ?", false).Order("create_time desc")

	if pageSize > 0 {
		offset := page * pageSize
		query = query.Offset(offset).Limit(pageSize)
	}

	result := query.Find(&posts)
	if result.Error != nil {
		return nil, result.Error
	}

	postInfos := make([]model.PostInfo, 0, len(posts))
	for _, post := range posts {
		postInfos = append(postInfos, dbPostToPostInfo(&post))
	}

	return &model.ManageData{
		Posts: postInfos,
	}, nil
}

// GetPostsCount returns the total count of visible posts
func (d *GormDao) GetPostsCount() (int, error) {
	var count int64
	result := d.db.Model(&model.DBPost{}).Where("hide = ?", false).Count(&count)
	return int(count), result.Error
}

// UserLogin validates user credentials
func (d *GormDao) UserLogin(id, pass string) error {
	epass := fmt.Sprintf("%x", md5.Sum([]byte(pass)))

	var count int64
	d.db.Model(&model.DBUser{}).Count(&count)

	// First time login - create user
	if count == 0 {
		return d.db.Create(&model.DBUser{
			Username: id,
			Password: epass,
		}).Error
	}

	// Validate credentials
	var user model.DBUser
	result := d.db.Where("username = ? AND password = ?", id, epass).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return fmt.Errorf("invalid id or pass")
		}
		return result.Error
	}
	return nil
}

// UpdateUser updates user password
func (d *GormDao) UpdateUser(id, pass string) error {
	epass := fmt.Sprintf("%x", md5.Sum([]byte(pass)))
	return d.db.Model(&model.DBUser{}).Where("username = ?", id).Update("password", epass).Error
}

// GetNavis retrieves all navigation items
func (d *GormDao) GetNavis() (*model.NaviData, error) {
	var navis []model.DBNavi
	result := d.db.Order("sequence").Find(&navis)
	if result.Error != nil {
		return nil, result.Error
	}

	naviInfos := make([]model.NaviInfo, 0, len(navis))
	for _, navi := range navis {
		naviInfos = append(naviInfos, dbNaviToNaviInfo(&navi))
	}

	return &model.NaviData{
		Navis: naviInfos,
	}, nil
}

// GetAllTags retrieves all unique tags
func (d *GormDao) GetAllTags() ([]string, error) {
	var posts []model.DBPost
	result := d.db.Where("hide = ?", false).Select("tags").Find(&posts)
	if result.Error != nil {
		return nil, result.Error
	}

	tagMap := make(map[string]bool)
	for _, post := range posts {
		if post.Tags != "" {
			for _, tag := range strings.Split(post.Tags, ",") {
				tag = strings.TrimSpace(tag)
				if tag != "" {
					tagMap[tag] = true
				}
			}
		}
	}

	tags := make([]string, 0, len(tagMap))
	for tag := range tagMap {
		tags = append(tags, tag)
	}

	// Sort tags
	for i := 0; i < len(tags); i++ {
		for j := i + 1; j < len(tags); j++ {
			if tags[i] > tags[j] {
				tags[i], tags[j] = tags[j], tags[i]
			}
		}
	}
	return tags, nil
}

// GetPostsByTag retrieves posts by tag
func (d *GormDao) GetPostsByTag(tag string) (*model.ManageData, error) {
	// Use parameterized query to prevent SQL injection
	var posts []model.DBPost
	result := d.db.Where("hide = ? AND tags LIKE ?", false, "%"+tag+"%").
		Order("create_time desc").
		Find(&posts)
	if result.Error != nil {
		return nil, result.Error
	}

	// Filter exact tag matches
	filteredPosts := make([]model.PostInfo, 0)
	for _, post := range posts {
		if post.Tags != "" {
			for _, t := range strings.Split(post.Tags, ",") {
				if strings.TrimSpace(t) == tag {
					filteredPosts = append(filteredPosts, dbPostToPostInfo(&post))
					break
				}
			}
		}
	}

	return &model.ManageData{
		Posts: filteredPosts,
	}, nil
}

// UpdateNavi updates a navigation item
func (d *GormDao) UpdateNavi(data *model.DBNavi) error {
	return d.db.Where("id = ?", data.Id).Updates(data).Error
}

// NewNavi creates a new navigation item
func (d *GormDao) NewNavi(data *model.DBNavi) error {
	return d.db.Create(data).Error
}

// DelNavi deletes a navigation item
func (d *GormDao) DelNavi(id int64) error {
	return d.db.Where("id = ?", id).Delete(&model.DBNavi{}).Error
}

// Helper functions
func dbPostToPostInfo(post *model.DBPost) model.PostInfo {
	rawTags := strings.Split(post.Tags, ",")
	tags := make([]string, 0, len(rawTags))
	for _, tag := range rawTags {
		tag = strings.TrimSpace(tag)
		if tag != "" {
			tags = append(tags, tag)
		}
	}

	return model.PostInfo{
		Id:           post.Id,
		Title:        post.Title,
		Tags:         tags,
		Hidden:       post.Hide,
		CreateTime:   model.JSONTime(post.CreateTime),
		LastModified: model.JSONTime(post.LastModified),
	}
}

func dbNaviToNaviInfo(navi *model.DBNavi) model.NaviInfo {
	return model.NaviInfo{
		Id:       navi.Id,
		Sequence: navi.Sequence,
		Name:     navi.Name,
		Target:   navi.Target,
	}
}
