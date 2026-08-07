看文件或目录占多大。目录就递归累加里面所有文件，默认用 KB/MB 这种人类可读格式。

用法：
  go-filesize 文件名
  go-filesize 某个目录
  go-filesize -top 10 某个目录          # 列出最大的 10 个文件
  go-filesize -depth 2 某个目录          # 只统计前两层
  go-filesize -ext 某个目录              # 按扩展名汇总大小
  go-filesize -sort 某个目录             # 列出每个文件，按大小降序
  go-filesize -min 1048576 某个目录      # 只看 1MB 以上的文件
  go-filesize -json 某个目录             # 用 JSON 输出

参数：
  -h        用 KB/MB 这种人类可读格式（默认开）
  -top N    列出最大的 N 个文件
  -depth N  递归层级上限，0 表示不限
  -ext      按扩展名汇总大小
  -sort     目录下列出每个文件，按大小从大到小排
  -min 字节 只看大于等于这个字节数的文件
  -json     用 JSON 输出统计结果

测试：
  go test
