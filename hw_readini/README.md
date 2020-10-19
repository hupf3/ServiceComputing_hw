# 读取 ini 配置文件的程序包 readini

## 概述

本仓库包括了程序包说明文档 `README.md` ，实现程序包的过程的文档 `specification.md` ，程序包文件夹 `readini` 里面包含了实现读取和监听功能的函数；`main.go` 文件可以更方便进行读取文件，我已经将程序包中的所有函数放到了 `main.go` 中，只需要在文件中的 main 函数进行调用即可。`test.ini` 是用来测试的 ini 配置文件

## 获取包

使用命令 `go get github.com/hupf3/ServiceComputing_hw/hw_readini/readini`

## 使用说明

程序包中包括两个重要的函数文件 `read.go` 和 `watch.go` 以及二者的测试文件，两个文件的主要功能就是读取配置文件，和监听配置文件的改变，两个文件的主要函数如下所示：

- `read.go`

  ```go
  // 存储配置文件的相关信息
  type config struct {
  	filePath string                       // 配置文件路径
  	info     map[string]map[string]string // 二维map存储 section，key，value 的关系
  }
  
  // 先于 main 函数运行，用来判断操作系统
  func init() {
  	if runtime.GOOS == "unix" || runtime.GOOS == "darwin" {
  		flag = 0
  	} else if runtime.GOOS == "windows" {
  		flag = 1
  	}
  	fmt.Println("The current operating system is " + runtime.GOOS)
  	if flag == 1 {
  		fmt.Println("Use ; as the annotation character")
  	} else if flag == 0 {
  		fmt.Println("Use # as the annotation character")
  	}
  }
  
  func (c *config) readini() error {
  	// 打开文件
  	file, err := os.Open(c.filePath)
  
  	// 文件打开失败
  	if err != nil {
  		err = errors.New("Open file failed!")
  		fmt.Println("Error:", err)
  		return err
  	}
  
  	// 在函数调用后关闭文件
  	defer file.Close()
  
  	// 读取配置文件的内容
  	reader := bufio.NewReader(file)
  
  	var section, key, value string              // 配置文件的信息
  	var note string                             // 注释符号
  	c.info = make(map[string]map[string]string) // 初始化
  	if flag == 0 {
  		note = "#"
  	} else if flag == 1 {
  		note = ";"
  	}
  
  	// 读取文件信息
  	for {
  		// 用来读取文件的每一行
  		str, err := reader.ReadString('\n')
  		// 去除每行左右两边的空白
  		str = strings.TrimSpace(str)
  
  		if err != nil {
  			if err != io.EOF {
  				err = errors.New("Can not read the file!")
  				fmt.Println("Error:", err)
  				return err
  			}
  
  			// 文件读取完毕
  			if len(str) == 0 {
  				break
  			}
  		}
  		// 使用 switch 原因使代码读的更加清晰
  		switch {
  		case len(str) == 0: // 忽略空行
  		case string(str[0]) == note: // 忽略注释行
  		// 读取 section
  		case str[0] == '[' && str[len(str)-1] == ']':
  			section = str[1 : len(str)-1]
  			// fmt.Println(section)
  			c.info[section] = make(map[string]string)
  		default:
  			// 获取 key = value
  			i := strings.IndexAny(str, "=")
  			key = strings.TrimSpace(str[0:i])
  			value = strings.TrimSpace(str[i+1 : len(str)])
  
  			// 如果没有 section 则设置为空
  			if section == "" {
  				section = "emptySection"
  				c.info[section] = make(map[string]string)
  			}
  			// fmt.Println(key + " = " + value)
  			c.info[section][key] = value
  		}
  	}
  	// fmt.Println("hello  " + c.info["server"]["enforce_domain"])
  	return nil
  }
  ```

- `watch.go`

  ```go
  // 监听文件的变化
  func isChanged(infile string) (*config, error) {
  	// 原始配置文件
  	c := new(config)
  	c.filePath = infile
  	err := c.readini()
  	if err != nil {
  		return nil, err
  	}
  	// fmt.Println("hello  " + c.info["server"]["protocol"])
  	for {
  		// 新的配置文件
  		newc := new(config)
  		newc.filePath = infile
  		err := newc.readini()
  		if err != nil {
  			return nil, err
  		}
  
  		// 判断是否改变
  		var change bool = false
  		for s1, m1 := range c.info {
  			for k1, v1 := range m1 {
  				_, ok := newc.info[s1]
  				if !ok {
  					fmt.Println("File is changed!")
  					change = true
  				}
  				if change == true {
  					break
  				}
  
  				v2, ok := newc.info[s1][k1]
  				if !ok {
  					fmt.Println("File is changed!")
  					change = true
  				}
  				if v2 != v1 {
  					fmt.Println("File is changed!")
  					change = true
  				}
  				if change == true {
  					break
  				}
  			}
  			if change == true {
  				break
  			}
  		}
  
  		c = newc // 每次监听后都需要重新对原始配置文件进行修改
  	}
  	// return nil, nil
  }
  ```

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

