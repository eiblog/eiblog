### 安装
1、`Eiblog` 提供多个平台的压缩包下载，可到 [Eiblog release](https://github.com/eiblog/eiblog/releases) 选择相应版本和平台下载。也可通过：
``` sh
$ curl -L https://github.com/eiblog/eiblog/releases/download/v1.0.0/eiblog-v1.0.0.`uname -s | tr '[A-Z]' '[a-z]'`-amd64.tar.gz > eiblog-v1.0.0.`uname -s | tr '[A-Z]' '[a-z]'`-amd64.tar.gz
```

2、如果有幸你也是 `Gopher`，相信你会亲自动手，你可以通过：
``` sh
$ git clone https://github.com/eiblog/eiblog.git
```
进行源码编译二进制文件运行。

3、如果你对 `docker` 技术也有研究的话，你也可以通过 `docker` 来安装：
``` sh
$ docker pull registry.cn-hangzhou.aliyuncs.com/deepzz/eiblog:v1.2.0
```
`注意`，镜像内部没有提供 conf 文件夹内的配置内容，因为该内容定制化的需求过高。所以需要将 `conf` 目录映射出来，后面会具体说到。

### 本地测试
采用二进制包进行测试，在下载好可执行程序之后，我们可以开始本地测试的工作了。本地测试需要搭建两个服务 `mongodb` （必须）和 `elasticsearch2.4.1`（可选，搜索服务不可用）。

`Eiblog ` 默认会连接 `hostname` 为 `mongodb` 和 `elasticsearch` 的地址，因此你需要将信息填入 `/etc/hosts` 下。假如你搭建的 `mongodb` 地址为 `127.0.0.1:27017`，`elasticsearch` 地址为 `192.168.99.100:9200`，如：
``` sh
$ sudo vi /etc/hosts

# 在末尾加上两行
172.42.0.1      mongodb
192.168.99.100  elasticsearch
```

下面先看两个服务的搭建。

#### MongoDB 搭建

`MongoDB` 搭建，Mac 可通过 `brew install mongo` 进行安装，其它平台请查询资料。
#### Elasticsearch 搭建
`Elasticsearch `搭建，它的搭建要些许复杂。建议通过 `docker` 搭建。需要注意的是 es 自带的分析器对中文分词是不友好的，这里采用了 `elasticsearch-analysis-ik` 分词器。如果你想了解更多 [Github](https://github.com/medcl/elasticsearch-analysis-ik) 或则如何实现 [博客站内搜索](https://imququ.com/post/elasticsearch.html)。

1. pull 镜像  `docker pull elasticsearch:2.4.1`。

2. 添加环境变量 `ES_JAVA_OPTS: "-Xms512m -Xmx512m"`，除非你想让你的服务器爆掉。

3. 映射相关目录：

   ```
   $PWD/conf/es/config:/usr/share/elasticsearch/config
   $PWD/conf/es/plugins:/usr/share/elasticsearch/plugins
   ```

博主已经准备好了必要的 es 配置文件，请将这四个目录映射至 `eiblog` 下的 `conf` 目录。如果你想查看更多，请查看 `docker-compose.yml` 文件。

总结一下，`docker` 运行 es 的命令为：
``` sh
$ docker run -d --name eisearch \
    -p 9200:9200 \
    -e ES_JAVA_OPTS="-Xms512m -Xmx512m" \
    -v $PWD/conf/es/config:/usr/share/elasticsearch/config \
    -v $PWD/conf/es/plugins:/usr/share/elasticsearch/plugins \
    elasticsearch:2.4.1
```

之后执行 `./eiblog`，咱们的 `eiblog` 就可以运行起来了。

通过 `127.0.0.1:9000` 可以进入博客首页，`127.0.0.1:9000/admin/login` 进入后台登陆，账号密码为 `eiblog/conf/app.yml` 下的 `username` 和 `password`。初始账号密码 `deepz`、`deepzz`。

> `注意`，因为配置 `conf/app.yml` 均是博主自用配置。有些操作可能（如评论）会评论到我的博客，还请尽量避免，谢谢。

### 准备部署
如果你在感受了该博客的魅力了之后，仍然坚持想要搭建它。那么，恭喜你，获得的一款不想再更换的博客系统。下面，我们跟随步骤对部署流程进一步说明。

这里只提供 `Docker` 的相关部署说明。你如果需要其它方式部署，请参考该方式。

#### 前提准备
这里需要准备一些必要的东西，如果你已准备好。请跳过。

* `一台服务器`。
* `一个域名`，国内服务器需备案。
* `有效的证书`。通过开启 autocert 可自动申请更新证书。也可去七牛、qcloud 申请一年有效证书。
* `七牛CDN`。博客只设计接入了 七牛cdn，相信该 CDN 服务商不会让你失望。
* `Disqus`。作为博客评论系统，你得有翻墙的能力注册到该账号，具体配置我想又可以写一片博客了。简单说需要 `shorname` 和 `public key`。
* `Google Analystic`。数据统计分析工具。
* `Superfeedr`。加速 RSS 订阅。
* `Twitter`。希望你能够有一个 twitter 账号。

是不是这么多要求，很费解。其实当初该博客系统只是为个人而设计的，是自己心中想要的那一款。博主些这篇文章不是想要多少人来用该博客，而是希望对那些追求至极的朋友说：你需要这款博客系统。
#### 文件准备
博主是一个有强迫症的人，一些文件的路径我使用了固定的路径，请大家见谅。假如你的 cdn 域名为 `st.example.com`，你需要确定这些文件已经在你的 cdn 中，它们路径分别是：

| 文件                 | 地址                                       | 描述                                       |
| ------------------ | ---------------------------------------- | ---------------------------------------- |
| favicon.ico        | st.example.com/static/img/favicon.ico    | cdn 中的文件名为 `static/img/favicon.ico`。你也可以复制 favicon.ico 到 static 文件夹下，通过 example.com/favicon.ico 也是能够访问到。docker 用户可能需要重新打包镜像。 |
| bg04.jpg           | st.example.com/static/img/bg04.jpg       | 首页左侧的大背景图，需要更名请到 views/st_blog.css 修改。   |
| avatar.png         | st.example.com/static/img/avatar.png     | 头像                                       |
| blank.gif          | st.example.com/static/img/blank.gif      | 空白图片，[下载](https://st.deepzz.com/static/img/blank.gif) |
| default_avatar.png | st.example.com/static/img/default_avatar.png | disqus 默认图片，[下载](https://st.deepzz.com/static/img/default_avatar.png) |
| disqus.js          | st.example.com/static/js/disqus_xxx.js   | disqus 文件，你可以通过 https://short_name.disqus.com/embed.js 下载你的专属文件，并上传到七牛。更新配置文件 app.yml。 |

> 注意，cdn 提到的文件下载，请复制链接进行下载，因为博主使用了防盗链功能，还有：  
  1、每次修改 app.yml 文件（如：更换 cdn 域名或更新头像），如果你不知道是否应该提高 staticversion 一个版本，那么最好提高一个 +1。  
  2、每次手动修改 views 内的以 `st_` 开头的文件，请将 `app.yml` 中的 staticversion 提高一个版本。

#### 配置说明
走到这里，我相信只走到 `60%` 的路程。放弃还来得及。

这里会对 `eiblog/conf` 下的所有文件做说明，希望你做好准备。
```
├── app.yml                         # 博客配置文件
├── blackip.yml                     # 博客 ip 黑名单
├── es                              # elasticsearch 配置
│   ├── config                      # 配置文件
│   │   ├── analysis                # 同义词
│   │   ├── elasticsearch.yml       # 具体配置
│   │   ├── logging.yml             # 日志配置
│   │   └── scripts                 # 脚本文件夹
│   └── plugins                     # 插件文件夹
│       └── ik1.10.1                # ik 分词器
├── nginx                           # nginx 配置
│   ├── domain                      # 域名配置，nginx 会读区改文件夹下的 .conf 文件
│   │   └── eiblog.conf
│   ├── ip.blacklist                # nginx ip黑名单
│   └── nginx.conf                  # nginx 配置，请替换 nginx 原有配置
├── scts                            # ct 透明
│   ├── ecc
│   │   ├── aviator.sct
│   │   └── digicert.sct
│   └── rsa
│       ├── aviator.sct
│       └── digicert.sct
├── ssl                             # 证书相关文件，可参考 eiblog.conf 生成
│   ├── dhparams.pem
│   ├── domain.rsa.key
│   ├── domain.rsa.pem
│   ├── full_chained.pem
│   └── session_ticket.key
└── tpl                             # 模版文件
    ├── crossdomainTpl.xml
    ├── feedTpl.xml
    ├── opensearchTpl.xml
    ├── robotsTpl.xml
    └── sitemapTpl.xml
```
| 名称          | 描述                                       |
| ----------- | ---------------------------------------- |
| app.yml     | 整个程序的配置文件，里面已经列出了所有配置项的说明，这里不再阐述。        |
| blackip.yml | 如果没有使用 `Nginx`，博客内置 `ip` 过滤系统。           |
| es          | elasticsearch，非常强大的分布式搜索引擎，`github` 用的就是它。里面的配置基本不用修改，但 `es/analysis/synonym.txt` 是同义词，你可以照着已有的随意增加。scripts 是 es 的脚本文件夹 |
| nginx       | 系统采用 `nginx` 作为代理(相信博客系统也不会独占一台服务器～)。请使用 `nginx.conf` 替换原 `nginx` 的配置。博客系统的配置文件是 `domain/eiblog.conf`，或则重命名(只要是满足`*.conf`)。`eiblog.conf`文件里面学问是最多的。或许你想一一弄懂，或许…。注意本配置需要更新 nginx 到最新版，openssl 更新到1.0.2j，具体请到 Jerry Qu 的 [本博客 Nginx 配置之完整篇](https://imququ.com/post/my-nginx-conf.html) 查看，了解详情。 |
| scts        | 存放 ct 文件。                                |
| ssl         | 这里存放了所有证书相关的内容。                          |
| tpl         | 模版相关，不用修改。                               |

### 开始部署

#### docker
请确定你已经完成了上面所说的所有步骤，在本地已经测试成功。服务器上 `MognoDB` 和`Elasticsearch` 已经安装并已经运行成功。

首先，请将本地测试好的 `conf` 文件夹上传至服务器，建议存储到服务器 `/data/eiblog` 下。
``` sh
$ tree /data/eiblog -L 1

├── conf
```

然后，将镜像 PULL 到服务器本地。
``` sh
# PULL下Eiblog镜像
$ docker pull registry.cn-hangzhou.aliyuncs.com/deepzz/eiblog
```

最后，执行 `docker run` 命令，希望你能成功。
``` sh
$ docker run -d --name eiblog --restart=always \
    --add-host disqus.com:23.235.33.134 \
    --add-host mongodb:172.42.0.1 \
    --add-host elasticsearch:192.168.99.100 \
    -p 9000:9000 \
    -e GODEBUG=netdns=cgo \
    -v /data/eiblog/logdata:/eiblog/logdata \
    -v /data/eiblog/conf:/eiblog/conf \
    registry.cn-hangzhou.aliyuncs.com/deepzz/eiblog
```
这里默认 `MongDB` 和 `Elasticsearch` 均为 `docker` 部署，且名称为`eidb`，`eisearch`。

#### nginx + docker
通过 `Nginx+docker` 部署，是博主推荐的方式。这里采用 `Docker Compose` 管理我们整个博客系统。

请确认你已经成功安装好 `Nginx`、`docker`、`docker-compose`。Nginx 请一定参照 Jerry Qu 的[Nginx 配置完整篇](https://imququ.com/post/my-nginx-conf.html)。

首先，请将本地测试好的 `conf`，`docker-compose.yml` 文件夹和文件上传至服务器。`conf` 建议存储到服务器 `/data/eiblog` 下，`docker-compose.yml` 存放在你使用方便的地方。

``` sh
$ tree /data/eiblog -L 1

├── conf

$ ls ~/

docker-compose.yml
```

然后，执行：
``` sh
$ cd ~
$ docker-compose up -d
```

等待些许时间，成功运行。
