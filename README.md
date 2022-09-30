# EiBlog [![Build Status](https://travis-ci.org/eiblog/eiblog.svg?branch=v1.3.0)](https://travis-ci.org/eiblog/eiblog) [![License](https://img.shields.io/badge/license-MIT-brightgreen.svg)](LICENSE.md) [![Versuib](https://img.shields.io/github/tag/eiblog/eiblog.svg)](https://github.com/eiblog/eiblog/releases) 

> 博客项目结构参考模版：https://github.com/deepzz0/appdemo

用过其它博客系统，不喜欢，不够轻，不够快！这是我开发的第二款博客系统，也实在不想再在这件事情上过多纠结了。`EiBlog` 是一个比较稳定的博客系统，现已迭代至 `2.0` 版本，稳定性和维护你是不用担心的。

但它有着部署简单（上线复杂！）的特点，不推荐没有计算机知识的朋友搭建，欢迎咨询。该博客的个中优点（简洁、轻快，安全），等你体验。

Docker镜像地址：

* 博客服务：[deepzz0/eiblog](https://hub.docker.com/r/deepzz0/eiblog)
* 博客搜索：[deepzz0/elasticsearch](https://hub.docker.com/r/deepzz0/elasticsearch)
* 数据备份：[deepzz0/backup](https://hub.docker.com/r/deepzz0/backup)

### 快速体验

**二进制**

1、下载压缩包，到 [这里](https://github.com/eiblog/eiblog/releases) 下载 eiblog（非backup） 相应系统压缩包，然后解压缩。

2、启动服务：

```
./backend
```

**Docker**

```
$ docker run --name eiblog \
    -p 9000:9000 \
    deepzz0/eiblog:latest
```

**Docker compose**

参考项目根目录下的 [docker-compose.yml](https://github.com/eiblog/eiblog/blob/v2/docker-compose.yml)，修改相关配置：

```
$ docker-compose up -d
或
$ docker compose up -d
```

然后访问 `localhost:9000` 就可以了，后台地址 `localhost:9000/admin/login`，默认账户密码 `deepzz/deepzz`。

> 默认情况下未开启博客搜索 `elasticsearch`，需要的话需要启动 es 服务并修改配置 `app.yml`。

**数据库支持**

| 类型（driver） | 地址（source）示例                                           |
| -------------- | ------------------------------------------------------------ |
| mongodb        | mongodb://localhost:27017                                    |
| mysql          | user:password@tcp(localhost:3306)/eiblog?charset=utf8mb4&parseTime=True&loc=Local |
| postgres       | host=localhost port=5432 user=user password=password dbname=eiblog sslmode=disable |
| sqlite         | /path/eiblog.db                                              |
| sqlserver      | sqlserver://user:password@localhost:9930?database=eiblog     |
| clickhouse     | tcp://localhost:9000?database=eiblog&username=user&password=password&read_timeout=10&write_timeout=20 |

### 功能特性

本着博客本质用来分享知识的特点，`EiBlog` 不会有较强的定制功能（包括主题，CDN支持等），仅保持常用简单页面与功能：

```
首页、专题、归档、友链、关于、搜索
```

功能说明：

* 博客归档，利用时间线帮助我们将归纳博文，内容少于一年按月归档，大于则按年归档。
* 博客专题，有时候博文是同一系列，专题能够帮助我们很好归纳博文，对阅读者是非常友好的。
* 标签系统，每篇博文都可以打上不同标签，使得在归档和专题不满足的情况下自定义归档，这块辅助搜索简直完美。
* 搜索系统，依托ElasticSearch实现的站内搜索，速度与效率并存，再加上google opensearch，搜索只流畅。
* 管理后台，内嵌全功能 `Typecho` 后台系统，全功能 `Markdown` 编辑器让你感觉什么是简洁清爽。
* 谷歌统计，由于google api的速度问题，从而实现了后端API异步统计，使得博客页面加载飞速。
* Disqus评论，国内评论系统不友好，因此选择disqus，又由于众所周知原因国内不能用，实现另类disqus评论方式。
* 多存储后端，支持mongodb、mysql、postgres、sqlite等存储后端。
* 七牛CDN，支持在 `Markdown` 编辑器直接上传附件，让你只考虑编辑内容，解放思想。
* 自动备份，支持多存储后端的备份功能，备份数据保存到七牛CDN上。

当然，为了让整个系统加载速度更快，还做了更多优化措施：

* 文章评论数量（不重要）通过后端跑定时任务获取，所以有时评论数量是不对的，这样减少了 API 调用。
* 整站内容全部内存缓存，`mardown` 文档全部转换为 html 进行缓存，减少了转换过程。
* `.js`、`.css` 等静态文件浏览器本地存储，小图片 base64 内置到 css 中，二次访问不会产生网络带来的延迟，加速访问。通过版本控制更新。
* 最佳实践 nginx 配置，可以查看 `eiblog.conf`，开启压缩缩小传输量，服务器传输证书链、开启 `Session Resumption`、`Session Ticket`、`OCSP Stapling `等加速证书握手，再次提高速度。

### 博客页面

可以容易的看到 [ssllabs](https://www.ssllabs.com/ssltest/analyze.html?d=deepzz.com&latest) 评分`A+`，[myssl](https://myssl.com/deepzz.com) 评分`A+`，堪称完美。这些安全的相关配置会在后面的部署过程中接触到。

![show-home](./docs/img/show-home.png)
![show-home2](./docs/img/show-home2.png)
![show-admin](./docs/img/show-admin.png)

### 更多文档

* [安装部署](https://eiblog.github.io/eiblog/install)
* [写作须知](https://eiblog.github.io/eiblog/writing)
* [好玩功能](https://eiblog.github.io/eiblog/amusing)
* [如何备份](https://eiblog.github.io/eiblog/backup)

### 贡献成员

![graphs/contributors](https://opencollective.com/eiblog/contributors.svg?width=890&button=false)

### 授权许可

本项目采用 MIT 开源授权许可证，完整的授权说明已放置在 [LICENSE](https://github.com/eiblog/eiblog/blob/master/LICENSE) 文件中。

