package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/eiblog/eiblog/setting"
	"github.com/eiblog/utils/logd"
	"github.com/eiblog/utils/mgo"
	"github.com/gin-gonic/gin"
)

const (
	// 成功
	NOTICE_SUCCESS = "success"
	// 注意
	NOTICE_NOTICE = "notice"
	// 错误
	NOTICE_ERROR = "error"
)

// 全局 API
var APIs = make(map[string]func(c *gin.Context))

func init() {
	// 更新账号信息
	APIs["account"] = apiAccount
	// 更新博客信息
	APIs["blog"] = apiBlog
	// 更新密码
	APIs["password"] = apiPassword
	// 删除文章
	APIs["post-delete"] = apiPostDelete
	// 添加文章
	APIs["post-add"] = apiPostAdd
	// 删除专题
	APIs["serie-delete"] = apiSerieDelete
	// 添加专题
	APIs["serie-add"] = apiSerieAdd
	// 专题排序
	APIs["serie-sort"] = apiSerieSort
	// 删除草稿箱
	APIs["draft-delete"] = apiDraftDelete
	// 删除回收箱
	APIs["trash-delete"] = apiTrashDelete
	// 恢复回收箱
	APIs["trash-recover"] = apiTrashRecover
	// 上传文件
	APIs["file-upload"] = apiFileUpload
	// 删除文件
	APIs["file-delete"] = apiFileDelete
}

// 更新账号信息，Email、PhoneNumber、Address
func apiAccount(c *gin.Context) {
	e := c.PostForm("email")
	pn := c.PostForm("phoneNumber")
	ad := c.PostForm("address")
	logd.Debug(e, pn, ad)
	if (e != "" && !CheckEmail(e)) || (pn != "" && !CheckSMS(pn)) {
		responseNotice(c, NOTICE_NOTICE, "参数错误", "")
		return
	}

	err := UpdateAccountField(mgo.M{"$set": mgo.M{"email": e, "phonen": pn, "address": ad}})
	if err != nil {
		logd.Error(err)
		responseNotice(c, NOTICE_NOTICE, err.Error(), "")
		return
	}
	Ei.Email = e
	Ei.PhoneN = pn
	Ei.Address = ad
	responseNotice(c, NOTICE_SUCCESS, "更新成功", "")
}

// 更新博客信息
func apiBlog(c *gin.Context) {
	bn := c.PostForm("blogName")
	bt := c.PostForm("bTitle")
	ba := c.PostForm("beiAn")
	st := c.PostForm("subTitle")
	ss := c.PostForm("seriessay")
	as := c.PostForm("archivessay")
	if bn == "" || bt == "" {
		responseNotice(c, NOTICE_NOTICE, "参数错误", "")
		return
	}

	err := UpdateAccountField(mgo.M{"$set": mgo.M{"blogger.blogname": bn,
		"blogger.btitle": bt, "blogger.beian": ba, "blogger.subtitle": st,
		"blogger.seriessay": ss, "blogger.archivessay": as}})
	if err != nil {
		logd.Error(err)
		responseNotice(c, NOTICE_NOTICE, err.Error(), "")
		return
	}
	Ei.BlogName = bn
	Ei.BTitle = bt
	Ei.BeiAn = ba
	Ei.SubTitle = st
	Ei.SeriesSay = ss
	Ei.ArchivesSay = as
	Ei.CH <- SERIES_MD
	Ei.CH <- ARCHIVE_MD
	responseNotice(c, NOTICE_SUCCESS, "更新成功", "")
}

// 更新密码
func apiPassword(c *gin.Context) {
	logd.Debug(c.Request.PostForm.Encode())
	od := c.PostForm("old")
	nw := c.PostForm("new")
	cf := c.PostForm("confirm")
	if nw != cf {
		responseNotice(c, NOTICE_NOTICE, "两次密码输入不一致", "")
		return
	}
	if !CheckPwd(nw) {
		responseNotice(c, NOTICE_NOTICE, "密码格式错误", "")
		return
	}
	if !VerifyPasswd(Ei.Password, Ei.Username, od) {
		responseNotice(c, NOTICE_NOTICE, "原始密码不正确", "")
		return
	}
	newPwd := EncryptPasswd(Ei.Username, nw)

	err := UpdateAccountField(mgo.M{"$set": mgo.M{"password": newPwd}})
	if err != nil {
		logd.Error(err)
		responseNotice(c, NOTICE_NOTICE, err.Error(), "")
		return
	}
	Ei.Password = newPwd
	responseNotice(c, NOTICE_SUCCESS, "更新成功", "")
}

// 删除文章，软删除：移入到回收箱
func apiPostDelete(c *gin.Context) {
	var ids []int32
	for _, v := range c.PostFormArray("cid[]") {
		i, err := strconv.Atoi(v)
		if err != nil || int32(i) < setting.Conf.General.StartID {
			responseNotice(c, NOTICE_NOTICE, "参数错误", "")
			return
		}
		ids = append(ids, int32(i))
	}
	err := DelArticles(ids...)
	if err != nil {
		logd.Error(err)
		responseNotice(c, NOTICE_NOTICE, err.Error(), "")
		return
	}

	// elasticsearch
	err = ElasticDelIndex(ids)
	if err != nil {
		logd.Error(err)
	}
	// TODO disqus delete
	responseNotice(c, NOTICE_SUCCESS, "删除成功", "")
}

func apiPostAdd(c *gin.Context) {
	var (
		err error
		do  string
		cid int
	)
	defer func() {
		switch do {
		case "auto": // 自动保存
			if err != nil {
				c.JSON(http.StatusOK, gin.H{"fail": FAIL, "time": time.Now().Format("15:04:05 PM"), "cid": cid})
				return
			}
			c.JSON(http.StatusOK, gin.H{"success": SUCCESS, "time": time.Now().Format("15:04:05 PM"), "cid": cid})
		case "save", "publish": // 草稿，发布
			if err != nil {
				responseNotice(c, NOTICE_NOTICE, err.Error(), "")
				return
			}
			uri := "/admin/manage-draft"
			if do == "publish" {
				uri = "/admin/manage-posts"
			}
			c.Redirect(http.StatusFound, uri)
		}
	}()

	do = c.PostForm("do") // auto or save or publish
	slug := c.PostForm("slug")
	title := c.PostForm("title")
	text := c.PostForm("text")
	date := CheckDate(c.PostForm("date"))
	serie := c.PostForm("serie")
	tag := c.PostForm("tags")
	update := c.PostForm("update")
	if slug == "" || title == "" || text == "" {
		err = errors.New("参数错误")
		return
	}
	var tags []string
	if tag != "" {
		tags = strings.Split(tag, ",")
	}
	serieid := CheckSerieID(serie)
	artc := &Article{
		Title:      title,
		Content:    text,
		Slug:       slug,
		CreateTime: date,
		IsDraft:    do != "publish",
		Author:     Ei.Username,
		SerieID:    serieid,
		Tags:       tags,
	}
	cid, err = strconv.Atoi(c.PostForm("cid"))
	// 新文章
	if err != nil || cid < 1 {
		err = AddArticle(artc)
		if err != nil {
			logd.Error(err)
			return
		}
		cid = int(artc.ID)
		if !artc.IsDraft {
			// 异步执行，快
			go func() {
				// elastic
				ElasticIndex(artc)
				// rss
				DoPings(slug)
				// disqus
				ThreadCreate(artc)
			}()
		}
		return
	}

	// 旧文章
	artc.ID = int32(cid)
	_, a := GetArticle(artc.ID)
	if a != nil {
		artc.IsDraft = false
		artc.Count = a.Count
		artc.UpdateTime = a.UpdateTime
	}
	if CheckBool(update) {
		artc.UpdateTime = time.Now()
	}
	// 数据库更新
	err = UpdateArticle(mgo.M{"id": artc.ID}, artc)
	if err != nil {
		logd.Error(err)
		return
	}
	if !artc.IsDraft {
		ReplaceArticle(a, artc)
		// 异步执行，快
		go func() {
			// elastic
			ElasticIndex(artc)
			// rss
			DoPings(slug)
			// disqus
			if a == nil {
				ThreadCreate(artc)
			}
		}()
	}
}

// 只能逐一删除，专题下不能有文章
func apiSerieDelete(c *gin.Context) {
	for _, v := range c.PostFormArray("mid[]") {
		id, err := strconv.Atoi(v)
		if err != nil || id < 1 {
			responseNotice(c, NOTICE_NOTICE, err.Error(), "")
			return
		}
		err = DelSerie(int32(id))
		if err != nil {
			logd.Error(err)
			responseNotice(c, NOTICE_NOTICE, err.Error(), "")
			return
		}
	}
	responseNotice(c, NOTICE_SUCCESS, "删除成功", "")
}

// 添加专题，如果专题有提交 mid 即更新专题
func apiSerieAdd(c *gin.Context) {
	name := c.PostForm("name")
	slug := c.PostForm("slug")
	desc := c.PostForm("description")
	if name == "" || slug == "" || desc == "" {
		responseNotice(c, NOTICE_NOTICE, "参数错误", "")
		return
	}
	mid, err := strconv.Atoi(c.PostForm("mid"))
	if err == nil && mid > 0 {
		serie := QuerySerie(int32(mid))
		if serie == nil {
			responseNotice(c, NOTICE_NOTICE, "专题不存在", "")
			return
		}
		serie.Name = name
		serie.Slug = slug
		serie.Desc = desc
		serie.ID = int32(mid)
		err = UpdateSerie(serie)
		if err != nil {
			logd.Error(err)
			responseNotice(c, NOTICE_NOTICE, err.Error(), "")
			return
		}
	} else {
		err = AddSerie(name, slug, desc)
		if err != nil {
			logd.Error(err)
			responseNotice(c, NOTICE_NOTICE, err.Error(), "")
			return
		}
	}
	responseNotice(c, NOTICE_SUCCESS, "操作成功", "")
}

// NOTE 排序专题，暂未实现
func apiSerieSort(c *gin.Context) {
	v := c.PostFormArray("mid[]")
	logd.Debug(v)
}

// 删除草稿箱，物理删除
func apiDraftDelete(c *gin.Context) {
	for _, v := range c.PostFormArray("mid[]") {
		i, err := strconv.Atoi(v)
		if err != nil || i < 1 {
			responseNotice(c, NOTICE_NOTICE, "参数错误", "")
			return
		}
		err = RemoveArticle(int32(i))
		if err != nil {
			responseNotice(c, NOTICE_NOTICE, err.Error(), "")
			return
		}
	}
	responseNotice(c, NOTICE_SUCCESS, "删除成功", "")
}

// 删除垃圾箱，物理删除
func apiTrashDelete(c *gin.Context) {
	for _, v := range c.PostFormArray("mid[]") {
		i, err := strconv.Atoi(v)
		if err != nil || i < 1 {
			responseNotice(c, NOTICE_NOTICE, "参数错误", "")
			return
		}
		err = RemoveArticle(int32(i))
		if err != nil {
			responseNotice(c, NOTICE_NOTICE, err.Error(), "")
			return
		}
	}
	responseNotice(c, NOTICE_SUCCESS, "删除成功", "")
}

// 从垃圾箱恢复到草稿箱
func apiTrashRecover(c *gin.Context) {
	for _, v := range c.PostFormArray("mid[]") {
		i, err := strconv.Atoi(v)
		if err != nil || i < 1 {
			responseNotice(c, NOTICE_NOTICE, "参数错误", "")
			return

		}
		err = RecoverArticle(int32(i))
		if err != nil {
			responseNotice(c, NOTICE_NOTICE, err.Error(), "")
			return
		}
	}
	responseNotice(c, NOTICE_SUCCESS, "恢复成功", "")
}

// 上传文件到 qiniu 云
func apiFileUpload(c *gin.Context) {
	type Size interface {
		Size() int64
	}
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		logd.Error(err)
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	s, ok := file.(Size)
	if !ok {
		logd.Error("assert failed")
		c.String(http.StatusBadRequest, "false")
		return
	}
	filename := strings.ToLower(header.Filename)
	url, err := FileUpload(filename, s.Size(), file)
	if err != nil {
		logd.Error(err)
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	typ := header.Header.Get("Content-Type")
	c.JSON(http.StatusOK, gin.H{
		"title":   filename,
		"isImage": typ[:5] == "image",
		"url":     url,
		"bytes":   fmt.Sprintf("%dkb", s.Size()/1000),
	})
}

// 删除七牛 CDN 文件
func apiFileDelete(c *gin.Context) {
	defer c.String(http.StatusOK, "删掉了吗？鬼知道。。。")

	name := c.PostForm("title")
	if name == "" {
		logd.Error("参数错误")
		return
	}
	err := FileDelete(name)
	if err != nil {
		logd.Error(err)
	}
}

func responseNotice(c *gin.Context, typ, content, hl string) {
	if hl != "" {
		c.SetCookie("notice_highlight", hl, 86400, "/", "", true, false)
	}
	c.SetCookie("notice_type", typ, 86400, "/", "", true, false)
	c.SetCookie("notice", fmt.Sprintf("[\"%s\"]", content), 86400, "/", "", true, false)
	c.Redirect(http.StatusFound, c.Request.Referer())
}
