package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

// 统计一个路径占用字节数，目录就递归累加
func dirSize(path string) (int64, int, error) {
	var total int64
	var files int
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			total += info.Size()
			files++
		}
		return nil
	})
	return total, files, err
}

// 人类可读，到 GB 为止
func human(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for v := n / unit; v >= unit; v /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(n)/float64(div), "KMGTPE"[exp])
}

func main() {
	humanFlag := flag.Bool("h", true, "用 KB/MB 这种人类可读格式")
	flag.Parse()

	paths := flag.Args()
	if len(paths) == 0 {
		paths = []string{"."}
	}
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			fmt.Printf("%s: 读不到 (%v)\n", p, err)
			continue
		}
		var size int64
		var nfiles int
		if info.IsDir() {
			size, nfiles, err = dirSize(p)
			if err != nil {
				fmt.Printf("%s: 统计失败 (%v)\n", p, err)
				continue
			}
			fmt.Printf("%s  %s  (%d 个文件)\n", p, human(size), nfiles)
		} else {
			size = info.Size()
			if *humanFlag {
				fmt.Printf("%s  %s\n", p, human(size))
			} else {
				fmt.Printf("%s  %d\n", p, size)
			}
		}
	}
}
