这里只介绍通过 docker 进行安装部署的方式，二进制安装也可参考。

* [存储后端](#存储后端)
* [搜索引擎](#搜索引擎)
* [准备工作](#准备工作)
* [开始部署](#开始部署)

博主提供了下面将要用到的镜像，可到这里查看：[https://hub.docker.com/u/deepzz0](https://hub.docker.com/u/deepzz0)。由于所有配置均在 `app/conf.yml` 下，所以在通过 docker 部署时建议将配置映射出来方便调试。

### 存储后端

首先启动我们的存储后端，用来存储我们的博客数据。eiblog 目前支持多种存储后端：

```
# driver     # source
mongodb      mongodb://localhost:27017
postgres     host=localhost port=5432 user=user dbname=eiblog sslmode=disable password=password
mysql        user:password@tcp(127.0.0.1:3306)/eiblog?charset=utf8mb4&parseTime=True&loc=Local
sqlite       /path/eiblog.db
sqlserver    sqlserver://user:password@localhost:9930?database=eiblog
clickhouse   tcp://localhost:9000?database=eiblog&username=user&password=password&read_timeout=10&write_timeout=20
```

选择自己最熟悉的方式作为存储后端，然后修改 `conf/app.yml` 下的数据库地址：

```
database:
  driver: postgres
  source: host=localhost port=5432 user=postgres dbname=eiblog sslmode=disable password=MTI3LjAuMC4x
```

### 搜索引擎

博客强依赖 ElasticSearch 搜索引擎，如果仅调试可以跳过不部署。但对外提供服务强烈建议部署上 ES，这样可以提高体验感。博主提供了一个已经配置好的 docker 镜像：`deepzz0/elasticsearch`：

```
# 运行
$ docker run --name es \
    -p 9200:9200 \
    deepzz0/elasticsearch:2.4.3
```

修改 `conf/app.yml` 下的 `eshost` 配置：

```
# 如果不部署，请置空
eshost: http://localhost:9200
```

### 准备工作

整个博客部署的复杂点就在这里了，如果你真的想要一款不想再更换的博客，那么继续。

#### 提前准备

请提前准备好以下内容，方便后续工作：

* `一台服务器`，对外提供访问能力。
* `一个域名`，如果服务器在国内域名需要备案（免费域名不建议）。
* `SSL证书`，博客要求全站 HTTPS 访问 + 七牛 CDN。
* `Disqus评论`，作为博客评论系统，如果申请请自行 Google。简单说需要提供 `shortname` 和 `public key`。
* `Google Analystic`，数据统计分析工具。
* `Superfeedr`，加速 RSS 订阅。
* `Twitter账号`，希望你能有一个 twitter 账号。

要求很多吧。其实当初该博客系统只是为个人而设计的，是自己心中想要的那一款。博主些这篇文章不是想要多少人来用该博客，而是希望对那些追求至极的朋友说：你需要这款博客系统。

#### 文件准备

博主是一个有强迫症的人，一些文件的路径我使用了固定的路径，请大家见谅。假如你的 cdn 域名为 `st.example.com`，你需要确定这些文件已经在你的 cdn 中，它们路径分别是：

| 文件               | 地址                                         | 描述                                                         |
| ------------------ | -------------------------------------------- | ------------------------------------------------------------ |
| favicon.ico        | st.example.com/static/img/favicon.ico        | cdn 名为 `static/img/favicon.ico`。你也可以在代理服务器自行配置，只要通过 example.com/favicon.ico 也是能够访问到。 |
| bg04.jpg           | st.example.com/static/img/bg04.jpg           | cdn 名为 `static/img/bg04.jpg`，首页左侧的大背景图，需要更名请到 website/st_blog.css 修改。 |
| avatar.png         | st.example.com/static/img/avatar.png         | cdn 名为 `static/img/avatar.png`，个人博客头像               |
| blank.gif          | st.example.com/static/img/blank.gif          | cdn 名为 `static/img/blank.gif`，空白图片，复制链接下载 https://st.deepzz.com/static/img/blank.gif。 |
| default_avatar.png | st.example.com/static/img/default_avatar.png | cdn 名为 `static/img/default_avatar.png`，disqus 默认头像图片，复制链接下载 https://st.deepzz.com/static/img/default_avatar.png |

>  注意：
>
> 1. cdn 提到的文件下载，请复制链接进行下载，因为博主使用了防盗链功能。
> 2. 每次修改 app.yml 文件（如：更换 cdn 域名或更新头像），如果你不知道是否应该提高 staticversion 一个版本，那么最好提高一个 +1。
> 3. 每次手动修改 website 内的以 `st_` 开头的文件，请将 `app.yml` 中的 staticversion 提高一个版本。

#### 配置说明

走到这里，我相信只走到 `80%` 的路程。放弃还来得及。这里会对 `eiblog/conf` 下的所有文件做说明，希望你做好准备。

```
├── app.yml # 整站配置
└── tpl     # 相关模版
    ├── crossdomainTpl.xml
    ├── feedTpl.xml
    ├── opensearchTpl.xml
    ├── robotsTpl.xml
    └── sitemapTpl.xml
```

具体的配置内容已经在 `app.yml` 中进行说明了。

如果用 nginx 作为代理服务器，博主提供了一份示例配置 `eiblog/eiblog.conf`，该配置涉及到 `ssl` 相关配置建议存放于 `/etc/nginx/ssl` 下。其中关于 `ssl_dhparam`、站点认证均提供了相关配置。

### 开始部署

下面是博主通过 `docker-compose` 一键部署的文件内容，仅供参考：

```
version: '3'
services:
  mongodb:
    image: mongo:3.2
    volumes:
    - ${PWD}/mgodb:/data/db
    restart: always
  elasticsearch:
    image: deepzz0/elasticsearch:2.4.3
    restart: always
  eiblog:
    iamge: deepzz0/eiblog:latest
    volumes:
    - ${PWD}/conf:/app/conf
    extra_hosts:
    - "disqus.com:151.101.192.134"
    links:
    - elasticsearch
    - mongodb
    ports:
    - 9000:9000
    restart: always
  backup:
    image: deepzz0/backup:latest
    volumes:
    - ${PWD}/conf:/app/conf
    links:
    - mongodb
    restart: always
```

当启动成功之后，后续的代理配置请参考 `eiblog/eiblog.conf`。
