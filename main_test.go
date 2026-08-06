package main

import (
	"os"
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

func TestDirSize(t *testing.T) {
	// 建个临时目录放两个文件，校验累加正确
	dir := t.TempDir()
	os.WriteFile(dir+"/a", make([]byte, 100), 0644)
	os.WriteFile(dir+"/b", make([]byte, 250), 0644)
	sub := dir + "/sub"
	os.Mkdir(sub, 0755)
	os.WriteFile(sub+"/c", make([]byte, 50), 0644)
	size, files, err := dirSize(dir)
	if err != nil {
		t.Fatal(err)
	}
	if size != 400 {
		t.Fatalf("目录大小应为 400，实际 %d", size)
	}
	if files != 3 {
		t.Fatalf("文件数应为 3，实际 %d", files)
	}
}
