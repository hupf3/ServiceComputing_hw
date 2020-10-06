package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/spf13/pflag"
)

// selpg命令的相关参数
type selpgArgs struct {
	startPage int // 起始页码
	endPage   int // 终止页码
	pageLen   int // 页长度

	pageType bool // 判断是 -l 还是 -f

	inFileName string // 输入的文件名
	printDest  string // 打印的位置
}

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

// 主函数
func main() {
	var sa selpgArgs
	input(&sa) // 输入参数和输入文件
	check(&sa) // 检查输入的参数是否正确
	readF(&sa) // 读取文件信息
}
