package main

import (
	"encoding/json"
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
	// 加个 exp 上限兜底，不然单位表一旦不够长就是越界 panic
	const units = "KMGTPE"
	div, exp := int64(unit), 0
	for v := n / unit; v >= unit && exp < len(units)-1; v /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(n)/float64(div), units[exp])
}

// jsonMarshal 统一用缩进输出，避免在各处重复处理错误
func jsonMarshal(v interface{}) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(b)
}

func main() {
	humanFlag := flag.Bool("h", true, "用 KB/MB 这种人类可读格式")
	top := flag.Int("top", 0, "列出最大的 N 个文件（0 表示不列）")
	depth := flag.Int("depth", 0, "递归层级上限，0 表示不限")
	byExt := flag.Bool("ext", false, "按扩展名汇总大小")
	sortBySize := flag.Bool("sort", false, "目录下列出每个文件，按大小从大到小排")
	minSize := flag.Int64("min", 0, "只看大于等于这个字节数的文件（0 表示不限）")
	jsonOut := flag.Bool("json", false, "用 JSON 输出统计结果")
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
			if *jsonOut {
				printJSON(p, size, nfiles, files, int64(*top), *minSize)
				continue
			}
			fmt.Printf("%s  %s  (%d 个文件)\n", p, human(size), nfiles)
			if *sortBySize {
				printSorted(files, *minSize, *humanFlag)
			}
			if *top > 0 {
				printTop(files, *top)
			}
		} else {
			size := info.Size()
			if *minSize > 0 && size < *minSize {
				continue
			}
			if *jsonOut {
				b := jsonMarshal(map[string]interface{}{"path": p, "size": size, "size_human": human(size)})
				fmt.Println(b)
				continue
			}
			if *humanFlag {
				fmt.Printf("%s  %s\n", p, human(size))
			} else {
				fmt.Printf("%s  %d\n", p, size)
			}
		}
	}
}

func printSorted(files []fileEntry, minSize int64, humanFlag bool) {
	sort.Slice(files, func(i, j int) bool { return files[i].size > files[j].size })
	fmt.Println("各文件（按大小降序）：")
	for _, f := range files {
		if minSize > 0 && f.size < minSize {
			continue
		}
		if humanFlag {
			fmt.Printf("  %s  %s\n", human(f.size), f.path)
		} else {
			fmt.Printf("  %d  %s\n", f.size, f.path)
		}
	}
}

func printJSON(root string, size int64, nfiles int, files []fileEntry, top, minSize int64) {
	type item struct {
		Path string `json:"path"`
		Size int64  `json:"size"`
	}
	var topItems []item
	if top > 0 {
		sort.Slice(files, func(i, j int) bool { return files[i].size > files[j].size })
		for _, f := range files {
			if minSize > 0 && f.size < minSize {
				continue
			}
			if len(topItems) >= int(top) {
				break
			}
			topItems = append(topItems, item{f.path, f.size})
		}
	}
	out := map[string]interface{}{
		"path":       root,
		"size":       size,
		"size_human": human(size),
		"files":      nfiles,
	}
	if top > 0 {
		out["top"] = topItems
	}
	b := jsonMarshal(out)
	fmt.Println(b)
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
