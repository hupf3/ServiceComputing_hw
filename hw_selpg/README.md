# Golang 开发 selpg

## 实验要求

使用 golang 开发 [开发 Linux 命令行实用程序](https://www.ibm.com/developerworks/cn/linux/shell/clutil/index.html) 中的 **selpg**

提示：

1. 请按文档 **使用 selpg** 章节要求测试你的程序
2. 请使用 pflag 替代 goflag 以满足 Unix 命令行规范， 参考：[Golang之使用Flag和Pflag](https://o-my-chenjian.com/2017/09/20/Using-Flag-And-Pflag-With-Golang/)
3. golang 文件读写、读环境变量，请自己查 os 包
4. “-dXXX” 实现，请自己查 `os/exec` 库，例如案例 [Command](https://godoc.org/os/exec#example-Command)，管理子进程的标准输入和输出通常使用 `io.Pipe`，具体案例见 [Pipe](https://godoc.org/io#Pipe)
5. 请自带测试程序，确保函数等功能正确

## CLI 介绍

**CLI（Command Line Interface）**实用程序是Linux下应用开发的基础。正确的编写命令行程序让应用与操作系统融为一体，通过shell或script使得应用获得最大的灵活性与开发效率。例如：

- Linux提供了cat、ls、copy等命令与操作系统交互；
- go语言提供一组实用程序完成从编码、编译、库管理、产品发布全过程支持；
- 容器服务如docker、k8s提供了大量实用程序支撑云服务的开发、部署、监控、访问等管理任务；
- git、npm等也是大家比较熟悉的工具。

尽管操作系统与应用系统服务可视化、图形化，但在开发领域，CLI在编程、调试、运维、管理中提供了图形化程序不可替代的灵活性与效率。

## selpg介绍

`selpg` 实用程序，全名为 SELect PaGes。selpg 允许用户指定从输入文本抽取的页的范围，这些输入文本可以来自文件或另一个进程。selpg 是以在 Linux 中创建命令的事实上的约定为模型创建的，这些约定包括：

- 独立工作
- 在命令管道中作为组件工作（通过读取标准输入或文件名参数，以及写至标准输出和标准错误）
- 接受修改其行为的命令行选项

## 实验过程

#### 前言

本次代码实现，是参考老师给的开发说明参照中的[C语言实现版](https://www.ibm.com/developerworks/cn/linux/shell/clutil/selpg.c)

### 实验环境

**操作系统**：Mac OS

**编译环境**：VScode

### 实验准备

本实验的要求中需要使用 `pflag` 代替 `flag`，所以先需要获取 `github` 上面关于 `pflag` 的包，具体的命令为：`go get github.com/spf13/pflag`，获取之后就可以在之前的 `GOPATH` 路径中找到获取的包

<img src="https://img-blog.csdnimg.cn/2020100622073691.png?x-oss-process=image/watermark,type_ZmFuZ3poZW5naGVpdGk,shadow_10,text_aHR0cHM6Ly9ibG9nLmNzZG4ubmV0L3FxXzQzMjY3Nzcz,size_16,color_FFFFFF,t_70#pic_center" alt="在这里插入图片描述" style="zoom:33%;" />

### 代码实现

#### import 包

由于需要进行获取命令行参数，以及读取文件内容，和正常输入输出内容，所以需要的包如下：

```go
import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/spf13/pflag"
)
```

#### 全局变量

设置了一个结构体 `selpgArgs` 里面的成员变量包括了参数的相关信息，如起始页码、终止页码、页长度、判断命令为 -f 还是 -l 、输入文件的名字、打印的位置：

```go
// selpg命令的相关参数
type selpgArgs struct {
	startPage int // 起始页码
	endPage   int // 终止页码
	pageLen   int // 页长度

	pageType bool // 判断是 -l 还是 -f

	inFileName string // 输入的文件名
	printDest  string // 打印的位置
}
```

#### input() 函数

函数实现了获取输入参数以及文件名，具体使用了 pflag 的一些函数的使用，具体使用方法参照[Golang之使用Flag和Pflag](https://o-my-chenjian.com/2017/09/20/Using-Flag-And-Pflag-With-Golang/)，其他的说明代码的注释也说的十分清楚：

```go
// 输入参数以及文件名
func input(sa *selpgArgs) {
	// 使用pflag，可以添加shorthand参数
	pflag.IntVarP(&(sa.startPage), "startPage", "s", -1, "Set startPage")
	pflag.IntVarP(&(sa.endPage), "endPage", "e", -1, "Set endPage")
	// 为了方便查看输出情况默认10行为1页
	pflag.IntVarP(&(sa.pageLen), "pageLength", "l", 10, "Set pageLength")
	pflag.StringVarP(&(sa.printDest), "printDest", "d", "", "Set printDest")
	pflag.BoolVarP(&(sa.pageType), "pageType", "f", false, "Set pageType")

	pflag.Parse() // flag解析，将命令行参数加入到绑定的变量

	fileName := pflag.Args() // 获取非flag命令行参数
	cnt := pflag.NArg()      // 获取参数的个数

	if cnt > 0 {
		sa.inFileName = string(fileName[0])
	} else {
		sa.inFileName = ""
	}
}
```

#### check() 函数

此函数实现了判断参数的输入是否正确，如果错误则会输出错误的信息，错误原因有：输入的参数补全、输入的起始和终止页码是负数、起始页码大于终止页码、页长度小于0：

```go
// 判断参数的正确与否
func check(sa *selpgArgs) {
	if (sa.startPage == -1) || (sa.endPage == -1) {
		fmt.Fprintf(os.Stderr, "\nError : not enough arguments\n")
		os.Exit(1)
	} else if (sa.startPage <= 0) || (sa.endPage <= 0) {
		fmt.Fprintf(os.Stderr, "\nError: invalid start page or invalid end page\n")
		os.Exit(1)
	} else if sa.startPage > sa.endPage {
		fmt.Fprintf(os.Stderr, "\nError: invalid arguments (start page <= end page)\n")
		os.Exit(1)
	} else if (sa.pageType == true) && (sa.pageLen != 10) {
		fmt.Fprintf(os.Stderr, "\nError: invalid arguments (-l and -f can not exit together)\n")
		os.Exit(1)
	} else if sa.pageLen <= 0 {
		fmt.Fprintf(os.Stderr, "\nError: invalid page length\n")
		os.Exit(1)
	} else {
		pageType := "page length."
		if sa.pageType == true {
			pageType = "The end sign /f."
		}
		fmt.Printf("\nArguments information: \n")
		fmt.Printf("- startPage: %d\n- endPage: %d\n- inputFile: %s\n- pageLength: %d\n- pageType: %s\n- printDestation: %s\n", sa.startPage, sa.endPage, sa.inFileName, sa.pageLen, pageType, sa.printDest)
	}
}
```

#### output() 函数

输出从文件中读取的内容，并且需要判断读取的方式是什么，也需要判断起始和终止页码设置的是否合理，文件的读入和输出是否有问题：

```go
// 输出
func output(fout interface{}, file *os.File, sa *selpgArgs) {

	cntL := 0
	var cntP int
	if sa.pageType {
		cntP = 0
	} else {
		cntP = 1
	}
	bio := bufio.NewReader(file)
	for true {
		var ss string
		var err error
		// 判断命令是 -f 还是 -l
		if sa.pageType {
			ss, err = bio.ReadString('\f') // 换页符
			cntP++
		} else {
			ss, err = bio.ReadString('\n')
			cntL++
			if cntL > sa.pageLen {
				cntP++
				cntL = 1
			}
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "\nError file read in:")
			panic(err)
		}

		if (cntP >= sa.startPage) && (cntP <= sa.endPage) {
			var errO error
			if stdOutput, ok := fout.(*os.File); ok {
				_, errO = fmt.Fprintf(stdOutput, "%s", ss)
			} else if pipeOutput, ok := fout.(io.WriteCloser); ok {
				_, errO = pipeOutput.Write([]byte(ss))
			} else {
				fmt.Fprintf(os.Stderr, "\nError:fout type error. ")
				os.Exit(1)
			}
			if errO != nil {
				fmt.Fprintf(os.Stderr, "\nError happend when output the pages:")
				panic(errO)
			}

		}
	}

	// 判断起始和终止页是否合理
	if cntP < sa.startPage {
		fmt.Fprintf(os.Stderr, "\nError: startPage (%d) greater than total pages (%d), no output written\n", sa.startPage, cntP)
		os.Exit(1)
	} else if cntP < sa.endPage {
		fmt.Fprintf(os.Stderr, "\nError: endPage (%d) greater than total pages (%d), less output than expected\n", sa.endPage, cntP)
		os.Exit(1)
	}
}
```

#### readF() 函数

读取文件，并且获取文件的状态，判断文件的打开是否有误，然后读取内容进行输出：

```go
// 读取文件
func readF(sa *selpgArgs) {
	var file *os.File
	if sa.inFileName == "" {
		file = os.Stdin // 读取输入
	} else {
		// 返回文件的状态
		_, ferr := os.Stat(sa.inFileName)
		if os.IsNotExist(ferr) {
			fmt.Fprintf(os.Stderr, "\nError: input file \"%s\" does not exist\n", sa.inFileName)
			os.Exit(1)
		}

		// 判断文件打开是否有误
		var err error
		file, err = os.Open(sa.inFileName) // 打开文件
		if err != nil {
			fmt.Fprintf(os.Stderr, "\nError File input:")
			panic(err)
		}
	}

	// 输出
	if len(sa.printDest) == 0 {
		output(os.Stdout, file, sa)
	} else {
		cmd := exec.Command("lp", "-d"+sa.printDest)
		fout, err := cmd.StdinPipe()
		if err != nil {
			fmt.Fprintf(os.Stderr, "\nError StdinPipe:")
			panic(err)
		}

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		errs := cmd.Start()
		if errs != nil {
			fmt.Fprintf(os.Stderr, "\nError CMD Start:")
			panic(errs)
		}
		output(fout, file, sa)
	}
}
```

#### main() 函数

主要调用了上述的函数，实现整个命令的完成：

```go
// 主函数
func main() {
	var sa selpgArgs
	input(&sa) // 输入参数和输入文件
	check(&sa) // 检查输入的参数是否正确
	readF(&sa) // 读取文件信息
}
```

### 文件结构

最终的代码文件如下所示：

<img src="https://img-blog.csdnimg.cn/2020100622241314.png#pic_center" alt="在这里插入图片描述" style="zoom: 50%;" />

- `selpg.go`：该文件实现了 selpg 命令
- `test.txt`：该文件是用来测试代码编写是否正确

### 实验测试

#### 测试文件说明

测试文件的每一行由一个数字构成，第一行为1，第二行为2，以此类推，但是第9行并不是空，而是由一个换行符构成，但是在txt文件中并不会有什么显示的存在，该换行符是我在vim中进行复制过来的，所以后面的行数所写的数字都要相应减1

#### 测试过程

**说明**：我的实验环境传入参数的方法为，在命令行中输入`go run --.go args`，如上的格式均可

- 将第一页写至标准输出，并没有重定向或管道，命令为`go run selpg.go -s1 -e1 test.txt`

<img src="https://img-blog.csdnimg.cn/20201006223154769.png?x-oss-process=image/watermark,type_ZmFuZ3poZW5naGVpdGk,shadow_10,text_aHR0cHM6Ly9ibG9nLmNzZG4ubmV0L3FxXzQzMjY3Nzcz,size_16,color_FFFFFF,t_70#pic_center" alt="在这里插入图片描述" style="zoom:33%;" />

​		由于我默认的页大小为10行，所以没有任何参数的时候，会直接输出第一页10行的所有内容，并且会显示相关的一些参数输入的信息

- selpg 读取标准输入，重定向来自于 `test.txt` 文件而不是显示命名的文件名参数。输入的第一页被写入屏幕，命令为 `go run selpg.go -s1 -e1 < test.txt`

<img src="https://img-blog.csdnimg.cn/20201006223513570.png?x-oss-process=image/watermark,type_ZmFuZ3poZW5naGVpdGk,shadow_10,text_aHR0cHM6Ly9ibG9nLmNzZG4ubmV0L3FxXzQzMjY3Nzcz,size_16,color_FFFFFF,t_70#pic_center" alt="在这里插入图片描述" style="zoom: 33%;" />

​		可以看到输出的结果与上方相同所以实现成功

- 前半部分作为后半部分的标准输入，前半部分为第二页的内容，后半部分要显示第一页的内容，所以就是显示总共的第二页的内容，命令为 `go run selpg.go -s2 -e2 test.txt | go run selpg.go -s1 -e1 `

<img src="https://img-blog.csdnimg.cn/20201006224450583.png?x-oss-process=image/watermark,type_ZmFuZ3poZW5naGVpdGk,shadow_10,text_aHR0cHM6Ly9ibG9nLmNzZG4ubmV0L3FxXzQzMjY3Nzcz,size_16,color_FFFFFF,t_70#pic_center" alt="在这里插入图片描述" style="zoom:33%;" />

- 将第一页到第二页的内容写至标准输出；标准输出被 shell／内核重定向至“output_file”，命令为：`go run selpg.go -s1 -e2 test.txt > output.txt`

  <img src="https://img-blog.csdnimg.cn/20201006223914628.png#pic_center" alt="在这里插入图片描述" style="zoom:50%;" />

<img src="https://img-blog.csdnimg.cn/20201006223958186.png?x-oss-process=image/watermark,type_ZmFuZ3poZW5naGVpdGk,shadow_10,text_aHR0cHM6Ly9ibG9nLmNzZG4ubmV0L3FxXzQzMjY3Nzcz,size_16,color_FFFFFF,t_70#pic_center" alt="在这里插入图片描述" style="zoom:33%;" />

​		由上面两张图的结果显示可以看出，程序的输出结果并没有显示在终端中，而是将输出结果到了 `output.txt` 文件中，也证实了实现成功

- selpg 将第 1 页到第 2 页写至标准输出（屏幕）；所有的错误消息被 shell／内核重定向至“error_file”。命令为：`go run selpg.go -s1 -e2 test.txt 2>error_file`

<img src="https://img-blog.csdnimg.cn/20201006224739274.png?x-oss-process=image/watermark,type_ZmFuZ3poZW5naGVpdGk,shadow_10,text_aHR0cHM6Ly9ibG9nLmNzZG4ubmV0L3FxXzQzMjY3Nzcz,size_16,color_FFFFFF,t_70#pic_center" alt="在这里插入图片描述" style="zoom:33%;" />

- selpg 将第 1 页到第 2 页写至标准输出，标准输出被重定向至“output_file”；selpg 写至标准错误的所有内容都被重定向至“error_file”,命令为：`go run selpg.go -s1 -e2 test.txt >ouput.txt 2>error_file`

<img src="https://img-blog.csdnimg.cn/20201006224941748.png#pic_center" alt="在这里插入图片描述" style="zoom:50%;" />

<img src="https://img-blog.csdnimg.cn/20201006225003304.png?x-oss-process=image/watermark,type_ZmFuZ3poZW5naGVpdGk,shadow_10,text_aHR0cHM6Ly9ibG9nLmNzZG4ubmV0L3FxXzQzMjY3Nzcz,size_16,color_FFFFFF,t_70#pic_center" alt="在这里插入图片描述" style="zoom:33%;" />

​		由上图所示，并没有输出在终端，输出到了output文件中，即可证明输出成功

- selpg 将第 10 页到第 20 页写至标准输出，标准输出被重定向至“output_file”；selpg 写至标准错误的所有内容都被重定向至 /dev/null（空设备），这意味着错误消息被丢弃了。设备文件 /dev/null 废弃所有写至它的输出，当从该设备文件读取时，会立即返回 EOF。命令为：`go run selpg.go -s1 -e2 test.txt >output.txt 2>/dev/null`

<img src="https://img-blog.csdnimg.cn/20201006225242756.png#pic_center" alt="在这里插入图片描述" style="zoom:33%;" />

​		由于空设备，所以废弃了输出

- 使用-l命令将页长设置为 7 行，这样 selpg 就可以把输入当作被定界为该长度的页那样处理。第 1 页到第 2 页被写至 selpg 的标准输出（屏幕）。命令为：`go run selpg.go -s1 -e2 -l7 test.txt `

<img src="https://img-blog.csdnimg.cn/20201006225621669.png?x-oss-process=image/watermark,type_ZmFuZ3poZW5naGVpdGk,shadow_10,text_aHR0cHM6Ly9ibG9nLmNzZG4ubmV0L3FxXzQzMjY3Nzcz,size_16,color_FFFFFF,t_70#pic_center" alt="在这里插入图片描述" style="zoom:33%;" />

​		由参数的说明和最后的结果输出的显示都可以看出最终的实现成功

- -f命令实现输出的换页，由于我设置的默认页大小为10，我在第九行设置了换行符，所以只输出第一页的内容，来检验是否配置成功，命令为：`go run selpg.go -s1 -e1 -f test.txt`

<img src="https://img-blog.csdnimg.cn/20201006230023650.png?x-oss-process=image/watermark,type_ZmFuZ3poZW5naGVpdGk,shadow_10,text_aHR0cHM6Ly9ibG9nLmNzZG4ubmV0L3FxXzQzMjY3Nzcz,size_16,color_FFFFFF,t_70#pic_center" alt="在这里插入图片描述" style="zoom:33%;" />

​		有最后的输出以及参数的内容可以看到实现成功

- 第 1 页到第 2 页由管道输送至命令“lp -dlp1”，该命令将使输出在打印机 lp1 上打印，命令为`go run selpg.go -s1 -e2 -dlp1 test.txt`

<img src="https://img-blog.csdnimg.cn/20201006230048269.png?x-oss-process=image/watermark,type_ZmFuZ3poZW5naGVpdGk,shadow_10,text_aHR0cHM6Ly9ibG9nLmNzZG4ubmV0L3FxXzQzMjY3Nzcz,size_16,color_FFFFFF,t_70#pic_center" alt="在这里插入图片描述" style="zoom:33%;" />

​		由于没有打印机，所以并不会输出到哪，结果也显示了没有此文件和目录

#### 错误命令测试

以上均为正确的命令测试，以下用来测试错误的命令，看看错误是否能正确提示

-  `go run selpg.go -s4 -e2 -dlp1 test.txt`

<img src="https://img-blog.csdnimg.cn/20201006230417332.png#pic_center" alt="在这里插入图片描述" style="zoom:33%;" />

-  `go run selpg.go -s-12 -e2 -dlp1 test.txt`

<img src="https://img-blog.csdnimg.cn/20201006230507949.png#pic_center" alt="在这里插入图片描述" style="zoom:33%;" />

-  `go run selpg.go -e2 -dlp1 test.txt`

<img src="https://img-blog.csdnimg.cn/2020100623055565.png#pic_center" alt="在这里插入图片描述" style="zoom:33%;" />

-  `go run selpg.go -s1 -e2 -l-12 test.txt`

<img src="https://img-blog.csdnimg.cn/20201006230710785.png#pic_center" alt="在这里插入图片描述" style="zoom:33%;" />

-  `go run selpg.go -s1 -e2 -l1 -f test.txt`

![在这里插入图片描述](https://img-blog.csdnimg.cn/20201006231710146.png#pic_center)

通过上述的测试可以看到错误提示信息也成功实现。

至此实验完全结束。