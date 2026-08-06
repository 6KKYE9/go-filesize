package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// fileEntry 记录一个文件的路径和大小，给 -top 用
type fileEntry struct {
	path string
	size int64
}

// dirSize 统计一个路径占用字节数，目录就递归累加
// maxDepth 控制递归深度，<=0 表示不限；同时收集所有文件给 -top 用
func dirSize(path string, maxDepth int, files *[]fileEntry) (int64, int, error) {
	var total int64
	var nfiles int
	err := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			// 超过深度就别往里走：rel 的段数 +1 就是当前目录距起点的层级
			if maxDepth > 0 {
				rel, rerr := filepath.Rel(path, p)
				if rerr == nil && rel != "." {
					level := strings.Count(rel, string(os.PathSeparator)) + 1
					if level >= maxDepth {
						return filepath.SkipDir
					}
				}
			}
			return nil
		}
		total += info.Size()
		nfiles++
		if files != nil {
			*files = append(*files, fileEntry{p, info.Size()})
		}
		return nil
	})
	return total, nfiles, err
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
	top := flag.Int("top", 0, "列出最大的 N 个文件（0 表示不列）")
	depth := flag.Int("depth", 0, "递归层级上限，0 表示不限")
	byExt := flag.Bool("ext", false, "按扩展名汇总大小")
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
		if info.IsDir() {
			var files []fileEntry
			size, nfiles, err := dirSize(p, *depth, &files)
			if err != nil {
				fmt.Printf("%s: 统计失败 (%v)\n", p, err)
				continue
			}
			if *byExt {
				printByExt(files)
			}
			fmt.Printf("%s  %s  (%d 个文件)\n", p, human(size), nfiles)
			if *top > 0 {
				printTop(files, *top)
			}
		} else {
			size := info.Size()
			if *humanFlag {
				fmt.Printf("%s  %s\n", p, human(size))
			} else {
				fmt.Printf("%s  %d\n", p, size)
			}
		}
	}
}

func printTop(files []fileEntry, n int) {
	sort.Slice(files, func(i, j int) bool { return files[i].size > files[j].size })
	if n > len(files) {
		n = len(files)
	}
	fmt.Println("最大的几个文件：")
	for i := 0; i < n; i++ {
		fmt.Printf("  %s  %s\n", human(files[i].size), files[i].path)
	}
}

// 按扩展名把大小汇总，从大到小排
func printByExt(files []fileEntry) {
	totals := map[string]int64{}
	for _, f := range files {
		ext := strings.ToLower(filepath.Ext(f.path))
		if ext == "" {
			ext = "(无扩展名)"
		}
		totals[ext] += f.size
	}
	type kv struct {
		ext string
		sz  int64
	}
	pairs := make([]kv, 0, len(totals))
	for e, s := range totals {
		pairs = append(pairs, kv{e, s})
	}
	sort.Slice(pairs, func(i, j int) bool { return pairs[i].sz > pairs[j].sz })
	fmt.Println("按类型汇总：")
	for _, p := range pairs {
		fmt.Printf("  %-12s %s\n", p.ext, human(p.sz))
	}
}
