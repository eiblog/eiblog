# Eiblog Changelog

## v1.4.2 (2018-02-09)
* 修复博客初始化后，about 页面不能够评论 #6
* 修复编辑专题，按钮显示“添加专题”错误
* 优化“添加文章”从同步改为异步推送：feed，es，disqus。速度显著提升
* （**重要*）头像图片从 avatar.jpg 改为 avatar.png（透明）
* docker-compose.yml mongodb 去掉端口映射，防止用户将端口暴露至外网
* session key 每次重启随机生成等一些细节的修复

## v1.4.1 (2018-01-14)
* 修复创建新文章，disqus 不收录bug
* 修复创建新文章，归档页面不刷新bug
* 修复能够删除关于页面和友情链接页面bug
* 修复重复添加文章错误
* 注释掉 docker-compose.yml 自动备份内容，请自行解开
* 添加当月数大于12，归档页面使用年份归档
* 优化代码逻辑

## v1.4.0 (2018-01-01)
* fix 搜索页面 bug
* CGO_ENABLED=0 关闭 cgo
* 更新Makefile ct log 服务器
* 数据库数据终于可以备份了

## v1.3.4 (2017-11-29)
* fix page:admin/write-post autocomplete tag

## v1.3.3 (2017-11-27)
* fix docker image: exec user process caused "no such file or directory"

## v1.3.2 (2017-11-17)
* 修复文章自动保存引起的发布文章不成功的bug

## v1.3.1 (2017-11-05)
* 修复调整 关于、友情链接 创建时间出现文章乱序
* 修复评论时间计算错误
* 调整acme文件验证路径
* 更改七牛SDK包为github包。
* 调整七牛配置文件名称，app.yml: kodo -> qiniu，name -> bucket，请提高静态文件版本 staticversion

## v1.3.0 (2017-07-13)
* 更改 app.yml 配置项，将大部分配置归在 general 常规配置下。注意，部署时请先更新 app.yml。
* 静态文件采用动态渲染，即用户不再需要管理 view、static 目录。
* 通过 acme.sh 使用双证书啦，可到 Makefile 查看相关信息。
* 使用 autocert 自动生成证书功能，从此再也不用担心证书过期，移步 [证书更新](https://github.com/eiblog/eiblog/blob/master/docs/autocert.md)。
* 开启配置项 enablehttps， 将自动重定向 http 到 https 啦。
* disqus.js 文件由配置指定，请看 app.yml 下的 disqus 相关配置。

## v1.2.0 (2017-06-14)
* 更新评论功能，基础评论 0 回复也可评论了。
* disqus.js 文件由博主自行更新。
* 更正描述 README.md 描述错误 [#4f996](https://github.com/eiblog/eiblog/commit/4f9965b6bdefe087dd0805c1840afcb2752cd155)。
* docker 镜像版本化。

## v1.1.3 (2017-05-12)
* 更新 disqus_78bca4.js 到 disqus_921d24.js，具体请参考 docs/install.md
* 更新 vendor

## v1.1.2 (2017-03-08)
* 解决添加文章描述错误的bug
* 添加vendor目录
* 添加文档docs目录
* 删除多余注释

## v1.1.1 (2017-02-07)
* 添加文章描述功能。
* 修复评论`jQuery`文件引用错误。
* 修复`.travis.yml`描述错误。

## v1.0.0 (2016-01-09)
首次发布版本

* 全站`HTTPS`设计，安全、极速。
* `Elasticsearch`博客搜索系统。
* 开源`Typecho`完整博客后台。
* 全功能`Markdown`编辑器。
* 异步`Google analysts`分析统计。
* `Disqus`评论系统。
* 后台直接对接七牛`CDN`。
