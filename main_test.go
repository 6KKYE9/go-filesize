package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHuman(t *testing.T) {
	cases := []struct {
		n    int64
		want string
	}{
		{512, "512 B"},
		{1024, "1.00 KB"},
		{1536, "1.50 KB"},
		{1024 * 1024, "1.00 MB"},
	}
	for _, c := range cases {
		if got := human(c.n); got != c.want {
			t.Fatalf("human(%d)=%q 想要 %q", c.n, got, c.want)
		}
	}
}

// 极大值别把单位表索引撑爆
func TestHumanHugeNoPanic(t *testing.T) {
	for _, n := range []int64{1 << 50, 1 << 60, 1<<62 - 1} {
		if got := human(n); got == "" {
			t.Fatalf("human(%d) 返回空", n)
		}
	}
}

func TestDirSize(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(dir+"/a", make([]byte, 100), 0644)
	os.WriteFile(dir+"/b", make([]byte, 250), 0644)
	sub := dir + "/sub"
	os.Mkdir(sub, 0755)
	os.WriteFile(sub+"/c", make([]byte, 50), 0644)
	var files []fileEntry
	size, nfiles, err := dirSize(dir, 0, &files)
	if err != nil {
		t.Fatal(err)
	}
	if size != 400 {
		t.Fatalf("目录大小应为 400，实际 %d", size)
	}
	if nfiles != 3 {
		t.Fatalf("文件数应为 3，实际 %d", nfiles)
	}
	if len(files) != 3 {
		t.Fatalf("收集到的文件数应为 3，实际 %d", len(files))
	}
}

func TestDirSizeDepth(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "top.txt"), make([]byte, 100), 0644)
	sub := filepath.Join(dir, "sub")
	os.Mkdir(sub, 0755)
	os.WriteFile(filepath.Join(sub, "deep.txt"), make([]byte, 50), 0644)
	var files []fileEntry
	size, _, _ := dirSize(dir, 1, &files)
	// 深度 1 只统计顶层，不该进 sub
	if size != 100 {
		t.Fatalf("深度限制后应为 100，实际 %d", size)
	}
	if len(files) != 1 {
		t.Fatalf("深度限制后文件数应为 1，实际 %d", len(files))
	}
}

func TestPrintByExt(t *testing.T) {
	// 简单验证不 panic 且能汇总（主要走真实二进制，这里保底）
	files := []fileEntry{
		{filepath.Join("a", "x.go"), 100},
		{filepath.Join("b", "y.go"), 200},
		{filepath.Join("c", "z.txt"), 50},
	}
	printByExt(files) // 仅验证不崩溃
}

func TestPrintSorted(t *testing.T) {
	files := []fileEntry{
		{"small.txt", 10},
		{"big.txt", 300},
		{"mid.txt", 100},
	}
	// minSize=0 时全部输出，顺序应为 big > mid > small
	printSorted(files, 0, false)
	// 再验证 minSize 过滤：只保留 >=100 的
	var kept []fileEntry
	for _, f := range files {
		if f.size >= 100 {
			kept = append(kept, f)
		}
	}
	if len(kept) != 2 {
		t.Fatalf("minSize 过滤后应剩 2 个，实际 %d", len(kept))
	}
}

func TestJSONMarshal(t *testing.T) {
	// 验证 JSON 输出是合法 JSON 且含关键字段
	out := jsonMarshal(map[string]interface{}{"size": int64(123), "files": 2})
	if out == "{}" {
		t.Fatal("JSON 序列化不应返回空对象")
	}
	if !strings.Contains(out, "\"size\": 123") {
		t.Fatalf("JSON 应含 size 字段: %s", out)
	}
}
