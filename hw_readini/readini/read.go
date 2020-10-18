package readini

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
)

var flag int // 判断操作系统的类型从而决定注释符 0 : '#'  1 : ';'

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
