package main

import (
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/eiblog/eiblog/setting"
	"github.com/eiblog/utils/logd"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"
)

const (
	NOTICE_SUCCESS = "success"
	NOTICE_NOTICE  = "notice"
	NOTICE_ERROR   = "error"
)

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
}

func apiAccount(c *gin.Context) {
	e := c.PostForm("email")
	pn := c.PostForm("phoneNumber")
	ad := c.PostForm("address")
	logd.Debug(e, pn, ad)
	if (e != "" && !CheckEmail(e)) || (pn != "" && !CheckSMS(pn)) {
		responseNotice(c, NOTICE_NOTICE, "参数错误", "")
		return
	}
	Ei.Email = e
	Ei.PhoneN = pn
	Ei.Address = ad
	err := UpdateAccountField(bson.M{"$set": bson.M{"email": e, "phonen": pn, "address": ad}})
	if err != nil {
		responseNotice(c, NOTICE_NOTICE, err.Error(), "")
		return
	}
	responseNotice(c, NOTICE_SUCCESS, "更新成功", "")
}

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
	Ei.BlogName = bn
	Ei.BTitle = bt
	Ei.BeiAn = ba
	Ei.SubTitle = st
	Ei.SeriesSay = ss
	Ei.ArchivesSay = as
	err := UpdateAccountField(bson.M{"$set": bson.M{"blogger.blogname": bn, "blogger.btitle": bt, "blogger.beian": ba, "blogger.subtitle": st, "blogger.seriessay": ss, "blogger.archivessay": as}})
	if err != nil {
		responseNotice(c, NOTICE_NOTICE, err.Error(), "")
		return
	}
	Ei.CH <- SERIES_MD
	Ei.CH <- ARCHIVE_MD
	responseNotice(c, NOTICE_SUCCESS, "更新成功", "")
}

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
	if !VerifyPasswd(Ei.Password, Ei.BlogName, od) {
		responseNotice(c, NOTICE_NOTICE, "原始密码不正确", "")
		return
	}
	Ei.Password = EncryptPasswd(Ei.BlogName, nw)
	responseNotice(c, NOTICE_SUCCESS, "更改成功", "")
}

func apiPostDelete(c *gin.Context) {
	var err error
	defer func() {
		if err != nil {
			logd.Error(err)
			responseNotice(c, NOTICE_NOTICE, err.Error(), "")
			return
		}
		responseNotice(c, NOTICE_SUCCESS, "删除成功", "")
	}()
	err = c.Request.ParseForm()
	if err != nil {
		return
	}
	var ids []int32
	var i int
	for _, v := range c.Request.PostForm["cid[]"] {
		i, err = strconv.Atoi(v)
		if err != nil || i < 1 {
			err = errors.New("参数错误")
			return
		}
		ids = append(ids, int32(i))
	}
	err = DelArticles(ids...)
	if err != nil {
		return
	}
	// elasticsearch 删除索引
	err = ElasticDelIndex(ids)
	if err != nil {
		return
	}
}

func apiPostAdd(c *gin.Context) {
	var err error
	var publish bool
	var cid int
	defer func() {
		if !publish {
			if err == nil {
				c.JSON(http.StatusOK, gin.H{"success": SUCCESS, "time": time.Now().Format("15:04:05 PM"), "cid": cid})
			} else {
				logd.Error(err)
				c.JSON(http.StatusOK, gin.H{"fail": FAIL, "time": time.Now().Format("15:04:05 PM"), "cid": cid})
			}
		}
		if err == nil {
			c.Redirect(http.StatusFound, "/admin/manage-posts")
			return
		}
		logd.Error(err)
	}()
	do := c.PostForm("do") // save or publish
	slug := c.PostForm("slug")
	title := c.PostForm("title")
	text := c.PostForm("text")
	date := c.PostForm("date")
	serie := c.PostForm("serie")
	tag := c.PostForm("tags")
	update := c.PostForm("update")
	if title == "" || text == "" || slug == "" {
		err = errors.New("参数错误")
		return
	}
	var tags []string
	if tag != "" {
		tags = strings.Split(tag, ",")
	}
	t := CheckDate(date)
	serieid := CheckSerieID(serie)
	publish = CheckPublish(do)
	artc := &Article{
		Title:      title,
		Content:    text,
		Slug:       slug,
		CreateTime: t,
		IsDraft:    !publish,
		Author:     Ei.Username,
		SerieID:    serieid,
		Tags:       tags,
	}
	cid, err = strconv.Atoi(c.PostForm("cid"))
	if err != nil || cid < 1 {
		err = AddArticle(artc)
		if err != nil {
			logd.Error(err)
			return
		}
		cid = int(artc.ID)
		if publish {
			ElasticIndex(artc)
			DoPings(slug)
		}
		return
	}
	artc.ID = int32(cid)
	if CheckBool(c.PostForm("update")) {
		artc.UpdateTime = time.Now()
	}
	i, a := GetArticle(artc.ID)
	if a != nil {
		artc.IsDraft = false
		artc.Count = a.Count
		artc.UpdateTime = a.UpdateTime
	}
	if update != "" {
		artc.UpdateTime = time.Now()
	}
	err = UpdateArticle(bson.M{"id": artc.ID}, artc)
	if err != nil {
		logd.Error(err)
		return
	}
	if !artc.IsDraft {
		if a != nil {
			Ei.Articles = append(Ei.Articles[0:i], Ei.Articles[i+1:]...)
			DelFromLinkedList(a)
			ManageTagsArticle(a, false, DELETE)
			ManageSeriesArticle(a, false, DELETE)
			ManageArchivesArticle(a, false, DELETE)
			delete(Ei.MapArticles, a.Slug)
			a = nil
		}
		Ei.MapArticles[artc.Slug] = artc
		Ei.Articles = append(Ei.Articles, artc)
		sort.Sort(Ei.Articles)
		GenerateExcerptAndRender(artc)
		// elasticsearch 索引
		ElasticIndex(artc)
		DoPings(slug)
		if artc.ID >= setting.Conf.StartID {
			ManageTagsArticle(artc, true, ADD)
			ManageSeriesArticle(artc, true, ADD)
			ManageArchivesArticle(artc, true, ADD)
			AddToLinkedList(artc.ID)
		}
	}
}

func apiSerieDelete(c *gin.Context) {
	err := c.Request.ParseForm()
	if err != nil {
		responseNotice(c, NOTICE_NOTICE, err.Error(), "")
		return
	}
	// 只能逐一删除
	for _, v := range c.Request.PostForm["mid[]"] {
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

func apiSerieAdd(c *gin.Context) {
	name := c.PostForm("name")
	slug := c.PostForm("slug")
	desc := c.PostForm("description")
	if name == "" || slug == "" || desc == "" {
		responseNotice(c, NOTICE_NOTICE, "参数错误", "")
		return
	}
	mid, err := strconv.Atoi(c.Query("mid"))
	if err == nil && mid > 0 {
		serie := QuerySerie(int32(mid))
		if serie == nil {
			responseNotice(c, NOTICE_NOTICE, "not found serie", "")
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

// 暂未启用
func apiSerieSort(c *gin.Context) {
	err := c.Request.ParseForm()
	if err != nil {
		responseNotice(c, NOTICE_NOTICE, err.Error(), "")
		return
	}
	v := c.Request.PostForm["mid[]"]
	logd.Debug(v)
}

func apiDraftDelete(c *gin.Context) {
	err := c.Request.ParseForm()
	if err != nil {
		responseNotice(c, NOTICE_NOTICE, err.Error(), "")
		return
	}
	for _, v := range c.Request.PostForm["mid[]"] {
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

func apiTrashDelete(c *gin.Context) {
	logd.Debug(c.PostForm("key"))
	logd.Debug(c.Request.PostForm)
	err := c.Request.ParseForm()
	if err != nil {
		responseNotice(c, NOTICE_NOTICE, err.Error(), "")
		return
	}
	for _, v := range c.Request.PostForm["mid[]"] {
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

func apiTrashRecover(c *gin.Context) {
	logd.Debug(c.PostForm("key"))
	logd.Debug(c.Request.PostForm)
	err := c.Request.ParseForm()
	if err != nil {
		responseNotice(c, NOTICE_NOTICE, err.Error(), "")
		return
	}
	for _, v := range c.Request.PostForm["mid[]"] {
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

func apiFileUpload(c *gin.Context) {

	// file, header , err := c.Request.FormFile("upload")
	// filename := header.Filename
	// fmt.Println(header.Filename)
	// out, err := os.Create("./tmp/"+filename+".png")
	// if err != nil {
	//     log.Fatal(err)
	// }
	// defer out.Close()
	// _, err = io.Copy(out, file)
	// if err != nil {
	//     log.Fatal(err)
	// }
}

func responseNotice(c *gin.Context, typ, content, hl string) {
	if hl != "" {
		c.SetCookie("notice_highlight", hl, 86400, "/", "", true, false)
	}
	c.SetCookie("notice_type", typ, 86400, "/", "", true, false)
	c.SetCookie("notice", fmt.Sprintf("[\"%s\"]", content), 86400, "/", "", true, false)
	c.Redirect(http.StatusFound, c.Request.Referer())
}
