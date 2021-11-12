### 备份数据

EiBlog 镜像仓库地址：https://hub.docker.com/u/deepzz0，备份镜像为：deepzz0/backup。



目前仅支持同步 mongodb 数据到七牛云，参考 `app.yml`：

```
backupapp:
  mode:
    name: cmd-backup
    enablehttp: true
    httpport: 9001
  backupto: qiniu # 备份到七牛云
  interval: 7d # 多久备份一次
  validity: 60 # 保存时长days
  qiniu: # 七牛OSS
    bucket: backup
    domain: st.deepzz.com
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
    -v ${PWD}/conf:/app/conf
```

Docker-compose 请参考项目根目录下的 `docker-compose.yml` 文件。
