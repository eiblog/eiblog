# Eiblog Changelog

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
