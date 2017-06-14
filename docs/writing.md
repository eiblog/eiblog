### 郑重提醒
**标题**、**slug**、**内容**。在你点击保存的时候一定确保三者不能为空，否则页面刷新内容就没了。所以，养成一个良好的写作习惯很重要。

当然，博客的自动保存功能也非常的好。在你不确定是否发布前，你可以将之保存到草稿，以便下次继续编辑。

### 文章标题
文章标题，这个可能要看个人习惯。我习惯从三级标题开始（###），依次往下四级标题，五级标题...。要注意的是一定不能跳级：
```
### 标题一

#### 标题1.1
#### 标题1.2
##### 标题1.2.1
##### 标题1.2.2

### 标题二

##### 标题2.1

##### 标题2.2
###### 标题2.2.1
###### 标题2.2.2
```

结果是：
![article-title](http://7xokm2.com1.z0.glb.clouddn.com/article-title.png)

### 文章描述
文章描述，主要是给`html->head->meta`中的 name 为 description 用的。现采用了一个临时的办法：在文章的第一行通过前缀识别（只看第一行）。

该前缀可到`conf/app.yml`设置，默认为`Desc:`，如：

![article-description](http://7xokm2.com1.z0.glb.clouddn.com/img/article-description.png)

### 图片懒加载
博客系统提供图片懒加载功能（浏览到某个位置，图片才会加载），以此来提高页面加载速度。我们可根据需要是否使用。

当然由此带来的坏处就是rss不能够正确加载图片。后续看是否解决这个问题或朋友提PR。

首先看下图片的`markdown`标准写法：
```
![alt](img_addres)
```
如：
```
![sublime-dialog](https://st.deepzz.com/blog/img/dialog-box-without-all-contols.png)
```
![sublime-dialog](https://st.deepzz.com/blog/img/dialog-box-without-all-contols.png)

懒加载，需要为该图片指定大小（长高）：
```
![alt](img_addres =widthxheight)
```

x 为小写字母（x,y,z）中的 x。使页面未加载时也占了相应的位置大小，这样设计是为了让读者在浏览页面时不会感到抖动。

如：
```
![sublime-dialog](https://st.deepzz.com/blog/img/dialog-box-without-all-contols.png =640x301)
```

### 摘要截取
摘要截取主要是提供给首页显示，如：
![home-page](http://7xokm2.com1.z0.glb.clouddn.com/img/deepzz_home_page.jpg)

红框中圈出来的就是截取出来的内容。在 `conf/app.yml` 的配置项有两个：
```
# 自动截取预览, 字符数
length: 400
# 截取预览标识
identifier: <!--more-->

```
当程序不能检查到 identifier 的标识符时，会采用长度的方式进行截取。
