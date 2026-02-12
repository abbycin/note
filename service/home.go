/***********************************************
        File Name: home
        Author: Abby Cin
        Mail: abbytsing@gmail.com
        Created Time: 10/28/19 9:45 PM
***********************************************/

package service

import (
	"blog/conf"
	"blog/dbutil"
	"blog/logging"
	"blog/model"
	"blog/routers"
	"fmt"
	"html/template"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"
)

const PageSize = 10 // 每页显示10条

type Post struct {
	Date  time.Time
	Link  template.URL
	Title string
	Tags  []string
}

type Pagination struct {
	CurrentPage int
	TotalPages  int
	HasPrev     bool
	HasNext     bool
	PrevPage    int
	NextPage    int
	Pages       []int
}

type home struct {
	dao  *dbutil.Dao
	h    *Home
	text []byte
	etag string
	tag  string // 当前筛选的标签
	page int    // 当前页码
}

func (h *home) Write(c *routers.Context) {
	if len(h.text) == 0 {
		_, err := h.Update()
		if err != nil {
			c.Text(http.StatusNotFound, "not found")
			return
		}
	}

	header := c.Resp.Header()
	header.Add("ETag", h.etag)
	if h.etag == c.Req.Header.Get("If-None-Match") {
		c.Resp.WriteHeader(http.StatusNotModified)
		c.Resp.Write(nil)
	} else {
		c.Resp.WriteHeader(http.StatusOK)
		header.Add("Content-Length", strconv.FormatInt(int64(len(h.text)), 10))
		c.Resp.Write(h.text)
	}
}

func (h *home) Makelink(p *model.PostInfo) template.URL {
	r := path.Join(h.h.article, fmt.Sprintf("%v", p.Id))
	return template.URL(r)
}

func (h *home) Update() (interface{}, error) {
	var data *model.ManageData
	var err error
	var totalCount int

	// 如果指定了标签，按标签筛选（标签页不分页）
	if h.tag != "" {
		data, err = h.dao.GetPostsByTag(h.tag)
		if err != nil {
			return nil, err
		}
		totalCount = len(data.Posts)
	} else {
		// 正常分页查询
		data, err = h.dao.GetPostsByPage(h.page, PageSize)
		if err != nil {
			return nil, err
		}
		totalCount, err = h.dao.GetPostsCount()
		if err != nil {
			totalCount = len(data.Posts)
		}
	}

	if err != nil {
		return nil, err
	}
	navis, err := h.dao.GetNavis()
	if err != nil {
		return nil, err
	}

	// 获取所有标签
	allTags, err := h.dao.GetAllTags()
	if err != nil {
		allTags = []string{}
	}

	res := make([]Post, 0)
	for _, p := range data.Posts {
		if p.Hidden {
			continue
		}
		t := Post{
			Date:  time.Time(p.CreateTime),
			Link:  h.Makelink(&p),
			Title: p.Title,
			Tags:  p.Tags,
		}
		res = append(res, t)
	}

	// 计算分页信息
	pagination := h.calculatePagination(totalCount)

	h.text, err = h.h.build(res, navis, allTags, h.tag, pagination)
	if err != nil {
		return nil, err
	}
	h.etag = fmt.Sprintf("%v", time.Now().Unix())
	return nil, nil
}

func (h *home) calculatePagination(totalCount int) Pagination {
	totalPages := (totalCount + PageSize - 1) / PageSize
	if totalPages < 1 {
		totalPages = 1
	}

	currentPage := h.page + 1 // 页码从1开始显示
	if currentPage > totalPages {
		currentPage = totalPages
	}

	// 生成分页列表（显示当前页前后各2页）
	pages := make([]int, 0)
	startPage := currentPage - 2
	if startPage < 1 {
		startPage = 1
	}
	endPage := currentPage + 2
	if endPage > totalPages {
		endPage = totalPages
	}
	for i := startPage; i <= endPage; i++ {
		pages = append(pages, i)
	}

	return Pagination{
		CurrentPage: currentPage,
		TotalPages:  totalPages,
		HasPrev:     currentPage > 1,
		HasNext:     currentPage < totalPages,
		PrevPage:    currentPage - 1,
		NextPage:    currentPage + 1,
		Pages:       pages,
	}
}

type Home struct {
	model   model.Model
	cfg     *conf.Config
	article string
	cache   *home
}

func NewHome(cfg *conf.Config, dao *dbutil.Dao, r *routers.Router) *Home {
	h := &Home{
		model:   model.NewDefaultModel(path.Join(cfg.Model.TmplRoot, cfg.Model.Home.Tmpl)),
		cfg:     cfg,
		article: cfg.Model.Article.Api[:strings.Index(cfg.Model.Article.Api, ":")],
		cache:   &home{dao: dao},
	}

	h.cache.h = h

	r.GET(cfg.Model.Home.Api, h)
	// 添加分页路由 /page/:page
	r.GET("/page/:page", h)
	// 添加标签筛选路由 /tag/:tag
	r.GET("/tag/:tag", h)
	return h
}

func (h *Home) build(posts []Post, navis *model.NaviData, allTags []string, currentTag string, pagination Pagination) ([]byte, error) {
	h.model.Funcs(template.FuncMap{"formatTime": h.formatTime})
	return h.model.Parse(map[string]interface{}{
		"Title":      h.cfg.Model.Title,
		"Posts":      posts,
		"Navis":      navis,
		"Tags":       allTags,
		"CurrentTag": currentTag,
		"Pagination": pagination,
		"Year":       time.Now().Year(),
	})
}

func (h *Home) Serve(c *routers.Context) {
	// 获取页码
	pageStr := c.GetParam("page")
	page := 0
	if pageStr != "" {
		p, err := strconv.Atoi(pageStr)
		if err == nil && p > 1 {
			page = p - 1 // 内部页码从0开始
		}
	}

	// 获取标签
	tag := c.GetParam("tag")

	if tag != "" {
		// 为标签路由创建临时 cache
		tagCache := &home{
			dao:  h.cache.dao,
			h:    h,
			tag:  tag,
			page: 0,
		}
		tagCache.Update()
		tagCache.Write(c)
	} else if page > 0 {
		// 为分页路由创建临时 cache
		pageCache := &home{
			dao:  h.cache.dao,
			h:    h,
			tag:  "",
			page: page,
		}
		pageCache.Update()
		pageCache.Write(c)
	} else {
		h.cache.Write(c)
	}
}

func (h *Home) formatTime(arg interface{}) string {
	return arg.(time.Time).Format("2006-01-02 15:04:05")
}

func (h *Home) Update(arg interface{}) {
	_, err := h.cache.Update()
	if err != nil {
		logging.Error("upate home cache: %v", err)
	}
}
