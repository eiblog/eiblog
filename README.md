# EiBlog [![Build Status](https://travis-ci.org/eiblog/eiblog.svg?branch=v1.3.0)](https://travis-ci.org/eiblog/eiblog) [![License](https://img.shields.io/badge/license-MIT-brightgreen.svg)](LICENSE.md) [![Versuib](https://img.shields.io/github/tag/eiblog/eiblog.svg)](https://github.com/eiblog/eiblog/releases) 

> 系统根据[https://imququ.com](https://imququ.com)一系列文章和方向进行搭建，期间获得了QuQu的很大帮助，在此表示感谢。

用过其它博客系统，不喜欢，不够轻，不够快！自己做过共两款博客系统，完美主义的我（毕竟处女座）也实在是不想再在这件事情上过多纠结了。`Eiblog` 应该是一个比较稳定的博客系统，且是博主以后使用的博客系统，稳定性和维护你是不用担心的，唯独该系统部署过程太过复杂，并且不推荐没有计算机知识的朋友搭建，欢迎咨询。该博客的个中优点（明显快，安全），等你体验。

<!--more-->

### 介绍

整个博客系统涉及到模块如下：

* 自动更新证书：
  * 接入 [acme/autocert](https://github.com/golang/crypto/tree/master/acme/autocert)，在 TLS 层开启全自动更新证书，从此证书的更新再也不用惦记了，不过 Go 的 HTTPS 兼容性不够好（不想兼容），在如部分 IE 和 UC 之类的浏览器不能访问，请悉知。
  * 如果你采用如 Nginx 代理，推荐使用 [acme.sh](https://github.com/Neilpang/acme.sh) 实现证书的自动部署。博主实现 aliyun dns 的自动验证方式，详见 [Makefile/gencert](https://github.com/eiblog/eiblog/blob/master/Makefile)。
* `MongoDB`，博客采用 mongodb 作为存储数据库。
* `Elasticsearch`，采用 `elasticsearch` 作为博客的站内搜索，尽管占用内存稍高。
* `Disqus`，作为博客评论系统，国内大部分被墙，故实现两种评论方式。
* `Nginx`，作为反向代理服务器，并做相关 `http header` 和证书的设置。
* `Google Analytics`，作为博客系统的数据分析统计工具。
* `七牛 CDN`，作为博客系统的静态文件存储，博文的图片附件什么上传至这里。

### 图片展示

可以容易的看到 [httpsecurityreport](https://httpsecurityreport.com/?report=deepzz.com) 评分`96`，[ssllabs](https://www.ssllabs.com/ssltest/analyze.html?d=deepzz.com&latest) 评分`A+`，[myssl](https://myssl.com/deepzz.com) 评分`A+`，堪称完美。这些安全的相关配置会在后面的部署过程中接触到。

相关图片展示：
![show-home](http://7xokm2.com1.z0.glb.clouddn.com/static/img/show-home1.png)

![show-home2](http://7xokm2.com1.z0.glb.clouddn.com/static/img/show-home2.png)

![show-admin](http://7xokm2.com1.z0.glb.clouddn.com/static/img/show-admin.png)

![eiblog-mem](http://7xokm2.com1.z0.glb.clouddn.com/img/eiblog-mem.png)

> `注`：图片1，图片2是博客界面，图片3是后台界面，图片4是性能展示。

### 极速体验
1. 到 [这里](https://github.com/eiblog/eiblog/releases) 下载对应平台 `.tar.gz` 文件。

2. 搭建 `MongoDB`（必须）和 `Elasticsearch`（可选）服务，正式部署需要。

3. 修改 `/etc/hosts` 文件，添加 `MongoDB` 数据库 IP 地址，如：`127.0.0.1       mongodb`。

4. 执行 `./eiblog`，运行博客系统。看到：
```
...
...
[GIN-debug] Listening and serving HTTP on :9000
```
代表运行成功了。

默认监听 `HTTP 9000` 端口，后台 `/admin/login`，默认账号密码均为 `deepzz`。更多详细请查阅 [安装部署](https://github.com/eiblog/eiblog/blob/master/docs/install.md) 文档。

### 特色功能

作为博主之心血之作，`Eiblog` 实现了什么功能，有什么特点，做了什么优化呢？

1. 系统目前只有 `首页`、`专题`、`归档`、`友链`、`关于`、`搜索` 界面。相信已经可以满足大部分用户的需求。
2. `.js`、`.css` 等静态文件本地存储，小图片 base64 内置到 css 中，不会产生网络所带来的延迟，加速网页访问。版本控制方式，动态更新静态文件。
3. 采用谷歌统计，并实现异步（将访问信息发给后端，后端提交给谷歌）统计，加速访问速度。
4. 采用直接缓存 markdown 转过的 html 文档的方式，加速后端处理。响应速度均在 3ms 以内，真正极速。
5. 通过 Nginx 的配置，开启压缩缩小传输量，服务器传输证书链、开启 `Session Resumption`、`Session Ticket`、`OCSP Stapling `等加速证书握手，再次提高速度。
  * `CDN`，使用七牛融合CDN，并 `https` 化，实现全站 `https`。七牛可申请免费证书了。
  * `CT`，证书透明度检测，提供一个开放的审计和监控系统。可以让任何域名所有者或者 CA 确定证书是否被错误签发或者被恶意使用，从而提高 HTTPS 网站的安全性。
  * `OSCP`，在线证书状态协议。用来检验证书合法性的在线查询服务.
  * `HSTS`，强制客户端（如浏览器）使用 HTTPS 与服务器创建连接。可以很好的解决 HTTPS 降级攻击。
  * `HPKP`，HTTP 公钥固定扩展，防范由「伪造或不正当手段获得网站证书」造成的中间人攻击。该功能让我们选择信任哪些`CA`。请不要轻易尝试 Nginx 线上运行，因为该配置目前只指定了 Letsencrypt X3 和 TrustAsia G5 证书 pin-sha256。
  * `SSL Protocols`，罗列支持的 `TLS` 协议，SSLv3 被证实是不安全的。
  * `SSL dhparam`，迪菲赫尔曼密钥交换。
  * `Cipher suite`，罗列服务器支持加密套件。
6. 文章评论数量（不重要）后端跑定时脚本，定时更新，所以有时评论数是不对的。这样减少了 api 调用，又再次达到加速访问的目的。
7. 针对 `disqus` 被墙原因，实现 [Jerry Qu](https://imququ.com) 的另类评论方式，保证评论的流畅。
8. 开源 `Typecho` 完整后台系统，全功能 `markdown` 编辑器，让你体验什么是简洁清爽。
9. 博客后台直接对接 `七牛 SDK`，实现后台上传文件和删除文件的简单功能。
10. 采用 `elasticsearch` 作为站内搜索，添加 `google opensearch` 功能，搜索更加自然。
11. 自动备份数据库数据到七牛云。

### 文档

* [证书更新](https://github.com/eiblog/eiblog/blob/master/docs/autocert.md)
* [安装部署](https://github.com/eiblog/eiblog/blob/master/docs/install.md)
* [写作需知](https://github.com/eiblog/eiblog/blob/master/docs/writing.md)
* [好玩的功能](https://github.com/eiblog/eiblog/blob/master/docs/amusing.md)
* [关于备份](https://github.com/eiblog/backup)

### 成功搭建者博客

* [https://blog.netcj.com](https://blog.netcj.com) - Razeen's Blog

如果你的博客使用`Eiblog`搭建，你可以在 [这里](https://github.com/eiblog/eiblog/issues/1) 提交网址。
