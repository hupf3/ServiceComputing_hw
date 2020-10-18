# 读取 ini 配置文件的程序包 readini

## 概述

本仓库包括了程序包说明文档 `README.md` ，实现程序包的过程的文档 `specification.md` ，程序包文件夹 `readini` 里面包含了实现读取和监听功能的函数；`main.go` 文件可以更方便进行读取文件，我已经将程序包中的所有函数放到了 `main.go` 中，只需要在文件中的 main 函数进行调用即可。`test.ini` 是用来测试的 ini 配置文件

## 使用说明

程序包的使用，我们先创建两个文件，`main.go` 用来调用程序包，和 `test.ini` 被读取的配置文件。配置文件的信息我在程序包中也有放置，具体内容如下：

```
# possible values : production, development
app_mode = development

[paths]
# Path to where grafana can store temp files, sessions, and the sqlite3 db (if that is used)
data = /home/git/grafana

[server]
# Protocol (http or https)
protocol = http

# The http port  to use
http_port = 9999

# Redirect to correct domain if host header does not match domain
# Prevents DNS rebinding attacks
enforce_domain = true
```

得到了配置文件之后，我们可以在 main 函数中调用 Watch 函数，即可进行读取和监听文件包的内容：

```go
func main() {
	var a ListenFunc
	_, _ = Watch("test.ini", a)
}
```

然后可以运行此文件，首先会打印出，您当前使用的操作系统环境，和对应的注释符应当是什么：

<img src="https://img-blog.csdnimg.cn/20201018234832595.png#pic_center" alt="在这里插入图片描述" style="zoom:50%;" />

之后会先显示配置文件的初始信息是什么，显示的内容如下所示：

<img src="https://img-blog.csdnimg.cn/20201018234917137.png?x-oss-process=image/watermark,type_ZmFuZ3poZW5naGVpdGk,shadow_10,text_aHR0cHM6Ly9ibG9nLmNzZG4ubmV0L3FxXzQzMjY3Nzcz,size_16,color_FFFFFF,t_70#pic_center" alt="在这里插入图片描述" style="zoom:50%;" />

与上方配置文件的信息进行对应发现，读取配置文件成功，然后就可以进行修改配置文件中的内容，然后进行保存，在终端我们就可以看见此时会打印出一条提示信息：

<img src="https://img-blog.csdnimg.cn/20201018221312281.png#pic_center" alt="在这里插入图片描述" style="zoom:50%;" />

即可证明成功实现监听功能，再次修改文件也会打印出该提示信息。

至此，示例展示完毕，也是比较简单的进行了实现。也可以通过查看 API 文档进行具体的学习，API 文档生成的过程在 `specification.md` 中有具体的说明！

