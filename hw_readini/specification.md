[toc]

# 程序包开发，读取 ini 配置文件

## 任务目标

1. 熟悉程序包的编写习惯（idioms）和风格（convetions）
2. 熟悉 io 库操作
3. 使用测试驱动的方法
4. 简单 Go 程使用
5. 事件通知

## 设计说明

### 读取配置文件

我们首先需要读取 `.ini` 文件，读取的函数我都放在了 `read.go` 文件中

- 全局变量：`flag` 数据类型为 `int`，用于判断操作系统的类型，从而选择注释符：0 代表 '#' ，1 代表 ';'

```go
var flag int // 判断操作系统的类型从而决定注释符 0 : '#'  1 : ';'
```

- 结构体：config 用来存储配置文件的相关信息，包含了 `filePath`，代表配置文件的路径，info 用了二维 map 来存储 section，key，value 的信息

```go
// 存储配置文件的相关信息
type config struct {
	filePath string                       // 配置文件路径
	info     map[string]map[string]string // 二维map存储 section，key，value 的关系
}
```

- init 函数：由于老师要求在 init 函数中判断操作系统的类型，所以我在里面实现了该要求（由于我的电脑是 macbook，所以操作系统为 darwin，所以我将操作系统为 darwin 的，flag 也设置为0）

```go
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
```

- readini 函数：该函数用于读取配置文件中的内容，该函数是结构体定义的方法，传入了 *config 参数，传入了指针，可以方便修改。返回了 error 变量，用来判断文件是否存在，或者在读取时是否有误。信息的读取是通过二维 map 和相关的 string 操作，具体的过程代码中的注释部分也有说明(值得一提的是，如果配置信息没有section 的相关信息，而是直接出现 key = value 的形式，此时将section设置成固定的值："emptySection")

```go
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

### 监听配置文件

实现该功能需要一个 watch 函数，还需要实现 Listener 的接口，我把实现的代码放到了 `watch.go` 文件中

- Listener 接口：接口中定义了 listen 函数用于监听是否产生变化

```go
// Listener 接口
type Listener interface {
	listen(infile string)
}
```

- ListenFunc：实现接口的函数类型

```go
type ListenFunc func(infile string) (*config, error)
```

- listen 函数：实现了接口中定义的函数

```go
// 实现了接口里的 listen 函数
func (f ListenFunc) listen(infile string) (*config, error) {
	return f(infile)
}
```

- Watch 函数：监听配置文件的变化

```go
// Watch 监听配置文件的变化
func Watch(filename string, listener ListenFunc) (*config, error) {
	listener = isChanged
	return listener.listen(filename)
}
```

- isChanged 函数：具体实现了监听配置文件变化的函数，通过比对前后两个config 中的 info 信息是否一致进行判断配置信息是否改变

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

## 单元测试

- `read.go`：该文件主要是用来读取配置文件信息的，然后将读取到的配置信息保存到 `config` 结构体中。测试该文件首先写一个简单的 `test.ini`配置文件，配置文件的信息如下所示：

  ```json
  [paths]
  # Path to where grafana can store temp files, sessions, and the sqlite3 db (if that is used)
  data = /home/git/grafana
  ```

  写好测试文件后，需要写测试函数，就是用配置文件中的正确信息，和通过 `read.go` 文件得到的配置信息进行对比，如果二者一致，则会 PASS，若不一致，会生成错误信息，`read_test.go` 文件如下所示：

  ```go
  package readini
  
  import "testing"
  
  func Testread(t *testing.T) {
  	c := new(config)
  	c.filePath = "test.ini"
  	c.readini()
  	section := "paths"
  	key := "data"
  	value := "/home/git/grafana"
  	k, ok1 := c.info[section]
  	v, ok2 := c.info[section][key]
  	if v != value || !ok1 || !ok2 {
  		t.Errorf("expected: section : " + section + "; key : " + key + "; value : " + value)
  		t.Errorf("but got: section : " + section + "; key : " + k[key] + "; value : " + v)
  	}
  }
  ```

  写好代码后就可以在命令行中输入命令进行测试：`go test -v`

  得到的结果如下所示：

  <img src="https://img-blog.csdnimg.cn/20201018214520295.png#pic_center" alt="在这里插入图片描述" style="zoom:50%;" />

  可以看到成功测试功能

- `watch.go` ：该文件的作用是监听配置文件信息是否有改动，有改动则需要告知。由于该函数是需要一直监听一直执行的，所以我会在功能测试中进行具体的测试，并且展示具体的测试结果，该文件的测试文件`watch_test.go` 我暂时设置成空文件

## 功能测试

该部分主要进行整体的测试，所以需要一个包含 `main` 函数的文件，并且导入上面的包就可以进行测试

- 首先测试文件的读取功能是否正确，在 `main` 中定义一个接口函数的变量，然后直接调用 Watch 函数：

  ```go
  var a ListenFunc
  	_, _ = Watch("test.ini", a)
  ```

  从上面的调用函数中可以看出，调用的是 `test.ini` 配置文件，该文件的内容是老师在课间中要求的测试文件的内容，所以我直接进行复制测试：

  ```json
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

  调用后，在读取文件的函数中，直接打印该文件信息即可，打印的方式就是按照每块的 section，将 key = value 的内容打印下来，代码的实现如下所示：

  ```go
  for s, m := range c.info {
  		fmt.Println("[Section]:", s)
  		for k, v := range m {
  			fmt.Println("Key: " + k + "------" + " Value: " + v)
  		}
  	}
  ```

  调用后，得到的输出结果如下所示：

  <img src="https://img-blog.csdnimg.cn/20201018220356446.png?x-oss-process=image/watermark,type_ZmFuZ3poZW5naGVpdGk,shadow_10,text_aHR0cHM6Ly9ibG9nLmNzZG4ubmV0L3FxXzQzMjY3Nzcz,size_16,color_FFFFFF,t_70#pic_center" alt="在这里插入图片描述" style="zoom:50%;" />

  通过输出的结果与配置文件的结果进行比对发现，该功能实现正确

- 监听功能的测试，该功能测试主要通过在程序运行的途中进行修改文件，在终端会不会产生提示配置文件已经修改的信息，我们可以进行测试：

  首先配置文件还是如图所示，看看得到的输出结果：

  <img src="https://img-blog.csdnimg.cn/20201018220632309.png?x-oss-process=image/watermark,type_ZmFuZ3poZW5naGVpdGk,shadow_10,text_aHR0cHM6Ly9ibG9nLmNzZG4ubmV0L3FxXzQzMjY3Nzcz,size_16,color_FFFFFF,t_70#pic_center" alt="在这里插入图片描述" style="zoom:50%;" />

  此时我们修改 section 为 paths 的 key 由原来的 `data` 改为 `datas`

  <img src="https://img-blog.csdnimg.cn/20201018220829478.png#pic_center" alt="在这里插入图片描述" style="zoom:50%;" />

  进行保存后，程序会自动进行更新，并且打印出文件已经修改的信息：

  <img src="https://img-blog.csdnimg.cn/20201018221312281.png#pic_center" alt="在这里插入图片描述" style="zoom:50%;" />

  由上图可知程序监听到了更改文件的信息并且打印出了提示，我们继续修改 paths datas 对应的 value 值为 null，看最后的结果：

  <img src="https://img-blog.csdnimg.cn/20201018221437373.png#pic_center" alt="在这里插入图片描述" style="zoom:50%;" />

  <img src="https://img-blog.csdnimg.cn/20201018221451535.png#pic_center" alt="在这里插入图片描述" style="zoom:50%;" />

  由上图可知第二个提示信息即为刚刚进行的修改，所以该功能进行成功测试
  
- 对于不同的错误会有不同的提示信息，比如打开文件错误，或者文件不存在，或者读取文件有误，这体现了错误的自定义，具体的代码如下：

  ```go
  	// 文件打开失败
  	if err != nil {
  		err = errors.New("Open file failed!")
  		fmt.Println("Error:", err)
  		return err
  	}
  
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
  ```

  上面就是两个不同的错误的提示信息。

  我们来验证一下，文件不存在的错误信息，也就是打开文件出现错误，实现的方法也比较简单，直接将文件的名字改成一个不存在的名字即可，然后运行程序，看运行的结果

  <img src="https://img-blog.csdnimg.cn/20201019000258948.png#pic_center" alt="在这里插入图片描述" style="zoom:50%;" />

  可以看出自定义的错误成功实现

## API文档

自动生成 api 文档，首先需要安装 godoc 命令，由于国内的网络环境不能直接访问外网进行下载安装，所以需要选择更换方式，用以下命令即可成功实现：

`git clone https://github.com/golang/tools $GOPATH/src/golang.org/x/tools`

成功获取后执行以下命令：

`go build golang.org/x/tools`

这时会得到一个 godoc 文件，只需将该文件放到 $GOPATH/bin 中即可成功运行 godoc 命令，执行的结果如下所示：

<img src="https://img-blog.csdnimg.cn/20201018232906735.png#pic_center" alt="在这里插入图片描述" style="zoom:50%;" />

然后用浏览器打开网页：[http://127.0.0.1:8080](http://127.0.0.1:8080)，即可访问到 godoc 网页版，结果如下图所示：

<img src="https://img-blog.csdnimg.cn/20201018233118214.png?x-oss-process=image/watermark,type_ZmFuZ3poZW5naGVpdGk,shadow_10,text_aHR0cHM6Ly9ibG9nLmNzZG4ubmV0L3FxXzQzMjY3Nzcz,size_16,color_FFFFFF,t_70#pic_center" alt="在这里插入图片描述" style="zoom:50%;" />

然后我们找到我们新建的 `readini` 包，然后打开，就可以查看 API 文档：

<img src="https://img-blog.csdnimg.cn/20201019084034765.png?x-oss-process=image/watermark,type_ZmFuZ3poZW5naGVpdGk,shadow_10,text_aHR0cHM6Ly9ibG9nLmNzZG4ubmV0L3FxXzQzMjY3Nzcz,size_16,color_FFFFFF,t_70#pic_center" alt="在这里插入图片描述" style="zoom:50%;" />

## 总结

通过本次的学习，学会了包的创建，并且了解了测试文件对于包的意义，有了测试文件能够更好的检查包中的函数是否定义正确。学会了利用 godoc 命令生成 API 文档，对 go 语言的掌握也更加的深刻。

