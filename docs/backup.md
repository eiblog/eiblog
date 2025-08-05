### 备份数据

EiBlog 镜像仓库地址：https://hub.docker.com/u/deepzz0，备份镜像为：deepzz0/backup。



目前仅支持同步 mongodb 数据到七牛云，参考 `app.yml`：

```
apimode:
  name: cmd-backup
  listen: 0.0.0.0:9000
database: # 数据库配置
  driver: mongodb
  source: mongodb://localhost:27017/eiblog
backupto: qiniu # 备份到, default: qiniu
interval: 7d # 备份周期, default: 7d
validity: 60 # 备份保留时间, default: 60
qiniu: # 七牛OSS
  bucket: eiblog
  domain: st.deepzz.cn
  accesskey: MB6AXl_Sj_mmFsL-Lt59Dml2Vmy2o8XMmiCbbSeC
  secretkey: BIrMy0fsZ0_SHNceNXk3eDuo7WmVYzj2-zrmd5Tf
```



### 运行

1、获取备份镜像：

```
$ docker pull deepzz0/backup
```

2、启动备份镜像：

```
$ docker run --name backup \
    -v ${PWD}/etc/app.yml:/app/etc/app.yml
```

Docker-compose 请参考项目根目录下的 `docker-compose.yml` 文件。
