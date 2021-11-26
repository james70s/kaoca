package main

// https://blog.csdn.net/whatday/article/details/109287416

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// var picTypes = map[string]bool{
// 	"jpg":  true,
// 	"jpeg": true,
// 	"png":  true,
// 	"gif":  true,
// 	"bmp":  true,
// }

// var videoTypes = map[string]bool{
// 	"mp4": true,
// 	"mov": true,
// 	"avi": true,
// 	"wmv": true,
// 	"mkv": true,
// 	"rm":  true,
// 	"f4v": true,
// 	"flv": true,
// 	"swf": true,
// }

// 实际中应该用更好的变量名
var (
	h = flag.Bool("h", false, "This `help`")
	c = flag.Bool("c", false, "是拷贝还是移动文件.默认为移动文件.")
	t = flag.Bool("t", false, "如果文件名中包含时间信息，是否根据该时间信息重置文件的修改时间.")
)

func info() {
	log.Println(`

根据文件的修改时间整理到对应的 ‘年/月/日期’ 目录下
james70s@me.com

____________________________________O/_______
                                    O\
	`)
}

func usage() {
	info()

	fmt.Fprintf(os.Stderr, `
Usage: main [-ht] [from] [to] 

Etc: main -t ./src ./test

Options:
`)
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage // 改变默认的 Usage
	flag.Parse()       // 接受命令行参数

	if *h || flag.NArg() != 2 { // 该应用的命令行参数必须要有2个
		flag.Usage()
		return
	}
	info()

	// fmt.Println(flag.Args()) // 返回没有被解析的命令行参数
	// fmt.Println(flag.NArg())          // 返回没有被解析的命令行参数的个数
	// fmt.Println(flag.NFlag())         // 命令行设置的参数个数

	workPath(flag.Args()[0], flag.Args()[1]) // 运行主程序
}

// isVideoFileFix 检测文件后缀是否为视频格式
// func isVideoFileFix(fix string) string {
// 	if _, ok := videoTypes[fix]; ok {
// 		return "videos"
// 	}
// 	return ""
// }

// // isPicFileFix 检测文件后缀是否为图片格式
// func isPicFileFix(fix string) string {
// 	if _, ok := picTypes[fix]; ok {
// 		return "pictures"
// 	}
// 	return ""
// }

// // getAllFile 获取指定目录所有文件
// func getAllFile(pathname string) ([]string, error) {
// 	var s []string
// 	rd, err := ioutil.ReadDir(pathname)
// 	if err != nil {
// 		fmt.Println("read dir fail:", err)
// 		return s, err
// 	}

// 	for _, fi := range rd {
// 		if !fi.IsDir() {
// 			fullName := pathname + "/" + fi.Name()
// 			s = append(s, fullName)
// 		}
// 	}
// 	return s, nil
// }

func workPath(from, to string) {
	// 遍历输入文件夹
	err := filepath.Walk(from, func(srcFile string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", srcFile, err)
			return err
		}
		if info.IsDir() {
			//return filepath.SkipDir // 忽略子目录
			return nil
		}
		if !isPic(srcFile) && !isMov(srcFile) {
			fmt.Println("is not pic or movie, ignore it", srcFile)
			return nil
		}

		// 如果文件名中包含时间信息，是否根据该时间信息重置文件的修改时间
		if *t {
			setModifyTime(srcFile)
		}

		// 目标文件,etc: /Users/James/app/tools/arrange/t/2019/06/2019-06-13/181338.jpg
		destFile := getDestAbsPath(to, srcFile)
		if err = createPath(destFile); err != nil {
			return nil
		}

		// 拷贝文件
		if *c {
			if _, err = copyFile(srcFile, destFile); err != nil {
				fmt.Printf("拷贝文件失败：%s. %v\n", srcFile, err)
				return nil
			}
			fmt.Printf("拷贝文件: %s -> %s\n", srcFile, destFile)
			// os.Remove(path)  // 删除源文件
		} else { // 移动文件
			if err = os.Rename(srcFile, destFile); err != nil {
				fmt.Printf("移动文件失败: %s. %s\n", srcFile, err.Error())
				return nil
			}
			fmt.Printf("移动文件: %s -> %s\n", srcFile, destFile)
		}

		return nil
	})
	if err != nil {
		log.Fatal("filepath.Walk failed; detail: ", err)
	}
}

// ----------------------------------------------------------------

// 根据照片拍摄日期决定存储目录
// "2016-01-02 15:04:05" -> "2016/01/2016-01-02"
func getPlacePath(tm time.Time) string {
	return fmt.Sprintf("%d/%02d/%d-%02d-%02d", tm.Year(), tm.Month(), tm.Year(), tm.Month(), tm.Day())
	// return filepath.Join(strconv.Itoa(tm.Year()), fmt.Sprintf("%02d", tm.Month()))
}

func createPath(destFile string) (err error) {
	destDir := filepath.Dir(destFile)

	if _, err = os.Stat(destDir); os.IsNotExist(err) {
		if err = os.MkdirAll(destDir, 0755); err != nil {
			fmt.Printf("创建目录失败: %s. %s\n", destDir, err.Error())
			return err
		}
		fmt.Println("创建目录: ", strings.TrimLeft(destDir, "./"))
	}
	return nil
}

// // meta
// func modifyTime(fname string) time.Time {
// 	var tm time.Time
// 	var x *exif.Exif
// 	f, err := os.Open(fname)
// 	if err != nil {
// 		return time.Now()
// 	}

// 	if isPic(fname) {
// 		x, err = exif.Decode(f)
// 		if err != nil {
// 			goto UseFileTime
// 		}
// 		tm, _ = x.DateTime()
// 		return tm
// 	} else if isMov(fname) {
// 		// TODO
// 		return time.Now()
// 	}

// UseFileTime:
// 	fi, err := f.Stat()
// 	if err != nil {
// 		return time.Now()
// 	}

// 	tm = fi.ModTime()
// 	return tm
// }

// 根据文件的修改时间，获取文件将要存放的目录
// dest: ./t 目标路径
// src: test/2019-06-13 181338.jpg 原文件
func getDestAbsPath(dest string, src string) string {
	mt := getModifyTime(src) // 文件修改日期
	path := getPlacePath(mt) // 生成存放路径

	destPath := filepath.Join(dest, path, filepath.Base(src))
	absPath, _ := filepath.Abs(destPath)
	if absPath == "" {
		return destPath
	}
	return absPath
}

// func IsExist(path string) bool {
// 	_, err := os.Stat(path)
// 	return err == nil || os.IsExist(err)
// 	// 或者
// 	//return err == nil || !os.IsNotExist(err)
// 	// 或者
// 	//return !os.IsNotExist(err)
// }

func copyFile(src, des string) (written int64, err error) {
	srcFile, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer srcFile.Close()

	//获取源文件的权限
	fi, _ := srcFile.Stat()
	perm := fi.Mode()

	//desFile, err := os.Create(des)  //无法复制源文件的所有权限
	desFile, err := os.OpenFile(des, os.O_RDWR|os.O_CREATE|os.O_TRUNC, perm) //复制源文件的所有权限
	if err != nil {
		return 0, err
	}
	defer desFile.Close()

	return io.Copy(desFile, srcFile)
}

func isPic(fname string) bool {
	switch strings.ToLower(path.Ext(fname)) {
	case ".jpeg", ".jpg", ".png", ".bmp", ".gif", ".tiff", ".tif", ".pcx", ".svg", ".psd", ".raw", ".raf", ".heic":
		return true
	default:
		return false
	}
}

func isMov(fname string) bool {
	switch strings.ToLower(path.Ext(fname)) {
	case ".mp4", ".mov":
		return true
	default:
		return false
	}
}

// Time 字符串 -> 时间
func toTime(s string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02 15:04:05", s, time.Local)
}

// 时间戳 --> 日期字符串
// func Timestring(sec int64) string {
// 	return time.Unix(sec, 0).Format("2006-01-02 15:04:05")
// }

// func FileTime(file string) {
// 	if fi, err := os.Stat(file); err == nil {
// 		// tm := fi.ModTime()
// 		stat := fi.Sys().(*syscall.Stat_t)
// 		fmt.Printf("At:%s, Ct:%s, Mt:%s, Mt2:%s\n", Timestring(stat.Atimespec.Sec), Timestring(stat.Ctimespec.Sec), Timestring(stat.Mtimespec.Sec), fi.ModTime())
// 	} else {
// 		fmt.Println(err)
// 	}
// }

// 获取文件的修改时间
func getModifyTime(file string) time.Time {
	if fi, err := os.Stat(file); err == nil {
		return fi.ModTime()
	}
	return time.Now()
}

// 如果文件名中包含时间信息，那么根据该时间信息重置文件的修改时间，设置正确的modifyTime
func setModifyTime(srcFile string) {
	// 从文件名中取出日期格式字符串
	r := regexp.MustCompile(`(20[0-2][0-9])[-|_|\s]?([0-9]{2})[-|_|\s]?([0-9]{2})[-|_|\s]?([0-9]{2})[-|_|\s]?([0-9]{2})[-|_|\s]?([0-9]{2})`)

	fileName := filepath.Base(srcFile)       // 文件名
	matchs := r.FindStringSubmatch(fileName) // 测试文件名是否包含日期信息
	if matchs != nil {
		omt := getModifyTime(srcFile) // 老的信息

		date := fmt.Sprintf("%s-%s-%s %s:%s:%s", matchs[1], matchs[2], matchs[3], matchs[4], matchs[5], matchs[6])
		mt, _ := toTime(date)
		// fmt.Printf("%s %s %s\n", fileName, date, mt)
		// 设置正确的modifyTime
		if err := os.Chtimes(srcFile, time.Now(), mt); err == nil {
			fmt.Printf("重设修改时间：%s %s -> %s\n", fileName, omt, mt)
		} else {
			log.Fatal(err)
		}
	}
}

// func rename(from string, to string) {
// 	err := filepath.Walk(from, func(srcFile string, info os.FileInfo, err error) error {
// 		// FileTime(srcFile)

// 		if *t {
// 			setModifyTime(srcFile)
// 		}
// 		return nil
// 	})

// 	if err != nil {
// 		log.Fatal("filepath.Walk failed; detail: ", err)
// 	}
// }