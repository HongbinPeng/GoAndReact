package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// 运行方式：go run main.go
//
// 这份示例覆盖 os 标准库中与文件/目录操作相关的核心用法。
// 作业中 os 最常见的场景：
//   1. os.ReadFile("config.json")  读取配置文件
//   2. os.Create / os.WriteFile    写入监控报告
//   3. os.Stat / os.IsNotExist     判断文件是否存在
//   4. os.MkdirAll                 确保输出目录存在

func main() {
	demoDir := "os_demo_temp"
	// 先创建一个临时目录，所有演示文件都放在里面，避免污染项目根目录
	if err := os.MkdirAll(demoDir, 0755); err != nil {
		fmt.Println("创建临时目录失败：", err)
		return
	}
	// 演示结束后清理
	defer os.RemoveAll(demoDir)

	fmt.Println("========== 1. 工作目录 ==========")
	demoGetwd()

	fmt.Println("\n========== 2. 写入文件 ==========")
	demoWriteFile(demoDir)

	fmt.Println("\n========== 3. 读取文件 ==========")
	demoReadFile(demoDir)

	fmt.Println("\n========== 4. 创建 & 写入文件（os.Create） ==========")
	demoCreate(demoDir)

	fmt.Println("\n========== 5. 以追加模式写入 ==========")
	demoAppend(demoDir)

	fmt.Println("\n========== 6. 文件信息（os.Stat） ==========")
	demoStat(demoDir)

	fmt.Println("\n========== 7. 判断文件是否存在 ==========")
	demoFileExists(demoDir)

	fmt.Println("\n========== 8. 重命名 & 移动文件 ==========")
	demoRename(demoDir)

	fmt.Println("\n========== 9. 删除文件 & 目录 ==========")
	demoRemove(demoDir)

	fmt.Println("\n========== 11. 遍历目录树（filepath.WalkDir） ==========")
	demoWalkDir(demoDir)

	fmt.Println("\n========== 12. 路径操作（filepath 包） ==========")
	demoPath()

	fmt.Println("\n========== 13. 修改文件权限（os.Chmod） ==========")
	demoChmod(demoDir)

	fmt.Println("\n========== 14. 软链接（os.Symlink / os.Readlink） ==========")
	demoSymlink(demoDir)

	fmt.Println("\n========== 15. 临时文件 ==========")
	demoTempFile(demoDir)

	fmt.Println("\n========== 16. 文件权限说明 ==========")
	demoPermExplain()

	fmt.Println("\n========== 演示结束，清理临时目录 ==========")
}

// ===========================================================================
// 1. 获取当前工作目录
// ===========================================================================
func demoGetwd() {
	// os.Getwd 返回程序运行时的工作目录
	// 在 VS Code 中 "Run" 时，通常是项目根目录
	dir, err := os.Getwd()
	if err != nil {
		fmt.Println("获取工作目录失败：", err)
		return
	}
	fmt.Println("当前工作目录：", dir)

	// os.Chdir 可以切换工作目录（一般不常用）
	// os.Chdir("/some/path")
}

// ===========================================================================
// 2. 写入文件 — os.WriteFile
// ===========================================================================
func demoWriteFile(dir string) {
	path := filepath.Join(dir, "write_demo.txt")

	// os.WriteFile 是一站式写入：打开 → 写入 → 关闭
	// 如果文件不存在则创建；如果已存在则覆盖全部内容
	// 参数：路径 | 内容（[]byte）| 权限
	content := []byte("这是 os.WriteFile 写入的第一行内容。\n第二行内容。\n")

	if err := os.WriteFile(path, content, 0644); err != nil {
		fmt.Println("写入失败：", err)
		return
	}
	fmt.Printf("已写入文件 %s（覆盖模式）\n", path)

	// 注意：os.WriteFile 每次都覆盖整个文件
	// 如果想追加内容，需要用 os.OpenFile 配合 os.O_APPEND 标志（见第 5 节）
}

// ===========================================================================
// 3. 读取文件 — os.ReadFile
// ===========================================================================
func demoReadFile(dir string) {
	path := filepath.Join(dir, "write_demo.txt")

	// os.ReadFile 是一站式读取：打开 → 读完 → 关闭
	// 返回 []byte，需要 string() 转换为字符串
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("读取失败：", err)
		return
	}
	fmt.Printf("读取到的内容：\n%s", string(data))

	// 适用场景：配置文件、小文件
	// 大文件（几百 MB 以上）应用 bufio.Scanner 逐行读取，避免内存占用过高
}

// ===========================================================================
// 4. 创建 & 写入文件 — os.Create + File.WriteString
// ===========================================================================
func demoCreate(dir string) {
	path := filepath.Join(dir, "create_demo.txt")

	// os.Create 等价于 os.OpenFile(name, O_RDWR|O_CREATE|O_TRUNC, 0666)
	// - 文件不存在 → 创建
	// - 文件已存在 → 截断（清空原有内容）
	// 返回 *os.File，需要手动 Close
	f, err := os.Create(path)
	if err != nil {
		fmt.Println("创建失败：", err)
		return
	}
	defer f.Close() // 必须 Close，否则文件描述符泄漏

	// 方法一：写入字符串
	if _, err := f.WriteString("这是通过 WriteString 写入的内容。\n"); err != nil {
		fmt.Println("写入失败：", err)
		return
	}

	// 方法二：写入 []byte
	if _, err := f.Write([]byte("这是通过 Write 写入的字节。\n")); err != nil {
		fmt.Println("写入失败：", err)
		return
	}

	// 写入后调用 Sync 可以强制刷盘（一般不需要，除非对持久化要求极高）
	// f.Sync()

	fmt.Printf("已通过 os.Create 创建并写入 %s\n", path)
}

// ===========================================================================
// 5. 追加写入 — os.OpenFile + O_APPEND
// ===========================================================================
func demoAppend(dir string) {
	path := filepath.Join(dir, "append_demo.txt")

	// 先写点初始内容
	os.WriteFile(path, []byte("初始内容\n"), 0644)

	// os.OpenFile 是最底层的文件打开方式，可以精细控制打开模式
	// 常用标志（可以 | 组合）：
	//   os.O_RDONLY  — 只读
	//   os.O_WRONLY  — 只写
	//   os.O_RDWR    — 读写
	//   os.O_CREATE  — 不存在则创建
	//   os.O_APPEND  — 追加写入（每次写入自动定位到文件末尾）
	//   os.O_TRUNC   — 截断文件（清空内容）
	//   os.O_EXCL    — 与 O_CREATE 配合使用，文件存在时返回错误
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("打开失败：", err)
		return
	}
	defer f.Close()

	// 追加一行
	if _, err := f.WriteString("这是追加的第一行\n"); err != nil {
		fmt.Println("追加失败：", err)
		return
	}
	if _, err := f.WriteString("这是追加的第二行\n"); err != nil {
		fmt.Println("追加失败：", err)
		return
	}

	// 读取验证
	data, _ := os.ReadFile(path)
	fmt.Printf("追加后文件内容：\n%s", string(data))
}

// ===========================================================================
// 6. 文件信息 — os.Stat
// ===========================================================================
func demoStat(dir string) {
	path := filepath.Join(dir, "write_demo.txt")

	// os.Stat 返回文件/目录的元信息（FileInfo）
	// 如果文件不存在，返回的 error 满足 os.IsNotExist(err) == true
	info, err := os.Stat(path)
	if err != nil {
		fmt.Println("Stat 失败：", err)
		return
	}

	// FileInfo 接口提供的方法：
	fmt.Printf("文件名    ：%s\n", info.Name())   // 文件名（不含路径）
	fmt.Printf("文件大小  ：%d 字节\n", info.Size()) // 大小（字节）
	fmt.Printf("是否目录  ：%v\n", info.IsDir())   // 是否为目录
	fmt.Printf("权限      ：%s\n", info.Mode())  // 权限位（如 -rw-r--r--）
	fmt.Printf("修改时间  ：%s\n", info.ModTime()) // 最后修改时间
	fmt.Printf("底层系统信息：%v\n", info.Sys())     // 平台相关数据（通常 nil）

	// os.Lstat 与 os.Stat 的区别：
	//   os.Stat  遇到软链接会跟踪到目标文件
	//   os.Lstat 遇到软链接返回软链接本身的信息
}

// ===========================================================================
// 7. 判断文件是否存在
// ===========================================================================
func demoFileExists(dir string) {
	exists := func(path string) bool {
		_, err := os.Stat(path)
		if err == nil {
			return true // 文件存在
		}
		// os.IsNotExist 判断是否是"文件不存在"类型的错误
		if os.IsNotExist(err) {
			return false // 文件不存在
		}
		// 其他错误（权限问题等），保守返回 false
		fmt.Printf("检查 %s 时遇到未知错误：%v\n", path, err)
		return false
	}

	// 测试存在的文件
	path1 := filepath.Join(dir, "write_demo.txt")
	fmt.Printf("%s 存在: %v\n", path1, exists(path1))

	// 测试不存在的文件
	path2 := filepath.Join(dir, "nonexistent.txt")
	fmt.Printf("%s 存在: %v\n", path2, exists(path2))
}

// ===========================================================================
// 8. 重命名 & 移动文件 — os.Rename
// ===========================================================================
func demoRename(dir string) {
	oldPath := filepath.Join(dir, "old_name.txt")
	newPath := filepath.Join(dir, "new_name.txt")

	// 先创建一个文件
	os.WriteFile(oldPath, []byte("重命名测试\n"), 0644)

	// os.Rename 可以重命名，也可以移动文件
	// 在同文件系统内相当于 mv，跨文件系统可能失败
	if err := os.Rename(oldPath, newPath); err != nil {
		fmt.Println("重命名失败：", err)
		return
	}
	fmt.Printf("已将 %s 重命名为 %s\n", oldPath, newPath)

	// 验证旧文件已不存在，新文件存在
	fmt.Printf("旧文件存在: %v, 新文件存在: %v\n",
		exists(oldPath), exists(newPath))
}

// ===========================================================================
// 9. 删除文件 & 目录
// ===========================================================================
func demoRemove(dir string) {
	// 创建几个用于删除的测试文件
	f1 := filepath.Join(dir, "to_delete_1.txt")
	f2 := filepath.Join(dir, "to_delete_2.txt")
	subDir := filepath.Join(dir, "to_delete_dir")
	os.WriteFile(f1, []byte("删我1\n"), 0644)
	os.WriteFile(f2, []byte("删我2\n"), 0644)
	os.Mkdir(subDir, 0755)

	// os.Remove — 删除单个文件或空目录
	if err := os.Remove(f1); err != nil {
		fmt.Println("删除失败：", err)
	} else {
		fmt.Printf("已删除文件 %s\n", f1)
	}

	// os.RemoveAll — 递归删除（删除目录及其下所有内容）
	// 注意：如果路径不存在，os.RemoveAll 不报错（返回 nil）
	if err := os.RemoveAll(subDir); err != nil {
		fmt.Println("递归删除失败：", err)
	} else {
		fmt.Printf("已递归删除目录 %s\n", subDir)
	}

	// 演示 os.RemoveAll 对不存在路径不报错
	err := os.RemoveAll("不存在的文件.txt")
	fmt.Printf("删除不存在文件: err=%v（nil 表示不报错）\n", err)
}

// ===========================================================================
// 10. 创建目录 & 遍历目录
// ===========================================================================
func demoDir(dir string) {
	subDir := filepath.Join(dir, "sub1", "sub2", "sub3")

	// os.Mkdir — 创建单级目录（父目录必须存在，否则报错）
	if err := os.Mkdir(filepath.Join(dir, "single"), 0755); err != nil {
		fmt.Println("Mkdir 失败：", err)
	}

	// os.MkdirAll — 递归创建多级目录（父目录不存在也会一起创建）
	// 如果目录已存在，不报错（返回 nil）
	if err := os.MkdirAll(subDir, 0755); err != nil {
		fmt.Println("MkdirAll 失败：", err)
		return
	}
	fmt.Printf("已创建目录 %s\n", subDir)

	// os.ReadDir — 读取目录内容，返回 []os.DirEntry（Go 1.16+）
	// DirEntry 是轻量级的目录条目，包含名称、类型等
	entries, err := os.ReadDir(dir)
	if err != nil {
		fmt.Println("读取目录失败：", err)
		return
	}

	// 排序，保证输出顺序稳定
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	fmt.Printf("目录 %s 内容：\n", dir)
	for _, entry := range entries {
		// entry.IsDir()  判断是否为目录
		// entry.Type()   返回文件类型（fs.FileMode）
		// entry.Name()   返回文件名
		typeStr := "文件"
		if entry.IsDir() {
			typeStr = "目录"
		}
		fmt.Printf("  [%s] %s\n", typeStr, entry.Name())
	}
}

// ===========================================================================
// 11. 遍历目录树 — filepath.WalkDir
// ===========================================================================
func demoWalkDir(dir string) {
	// 创建一些嵌套文件用于遍历
	os.WriteFile(filepath.Join(dir, "root.txt"), []byte("root"), 0644)
	nestedDir := filepath.Join(dir, "nested")
	os.MkdirAll(nestedDir, 0755)
	os.WriteFile(filepath.Join(nestedDir, "a.txt"), []byte("a"), 0644)
	os.WriteFile(filepath.Join(nestedDir, "b.txt"), []byte("b"), 0644)

	// filepath.WalkDir 递归遍历目录树，每个文件/目录都会调用回调函数
	// 回调函数返回非 nil error 时，WalkDir 停止遍历并返回该错误
	fmt.Printf("递归遍历 %s：\n", dir)
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err // 如果读取某个条目出错，中止遍历
		}

		// 计算缩进层级（通过相对路径中 "/" 的数量判断）
		rel, _ := filepath.Rel(dir, path)
		depth := strings.Count(rel, string(filepath.Separator))
		indent := strings.Repeat("  ", depth+1)

		icon := "📄"
		if d.IsDir() {
			icon = "📁"
		}
		fmt.Printf("%s%s %s\n", indent, icon, d.Name())
		return nil // 继续遍历
	})
	if err != nil {
		fmt.Println("遍历失败：", err)
	}

	// filepath.Walk 是旧版本 API，回调接收 os.FileInfo（会调用 Stat）
	// filepath.WalkDir 是新版本 API，回调接收 fs.DirEntry（更轻量，推荐）
}

// ===========================================================================
// 12. 路径操作 — filepath 包
// ===========================================================================
func demoPath() {
	// filepath.Join — 智能拼接路径（自动处理分隔符）
	// 跨平台兼容：Windows 用 \，Linux/macOS 用 /
	p := filepath.Join("week05", "practice", "os_01", "main.go")
	fmt.Printf("Join: %s\n", p)

	// filepath.Base — 获取路径的最后一个元素（文件名）
	fmt.Printf("Base(\"%s\"): %s\n", p, filepath.Base(p))

	// filepath.Dir — 获取路径的目录部分
	fmt.Printf("Dir(\"%s\"): %s\n", p, filepath.Dir(p))

	// filepath.Ext — 获取文件扩展名
	fmt.Printf("Ext(\"%s\"): %s\n", p, filepath.Ext(p))

	// filepath.IsAbs — 判断是否为绝对路径
	fmt.Printf("IsAbs(\"%s\"): %v\n", p, filepath.IsAbs(p))
	fmt.Printf("IsAbs(\"C:/test\"): %v\n", filepath.IsAbs("C:/test"))

	// filepath.Abs — 将相对路径转为绝对路径
	abs, _ := filepath.Abs(p)
	fmt.Printf("Abs(\"%s\"): %s\n", p, abs)

	// filepath.Rel — 计算相对路径
	base, _ := filepath.Rel("week05/practice", "week05/practice/os_01/main.go")
	fmt.Printf("Rel: %s\n", base)

	// filepath.Match — 通配符匹配（支持 * ? [ ]）
	matched, _ := filepath.Match("*.go", "main.go")
	fmt.Printf("Match(\"*.go\", \"main.go\"): %v\n", matched)
}

// ===========================================================================
// 13. 修改文件权限 — os.Chmod
// ===========================================================================
func demoChmod(dir string) {
	path := filepath.Join(dir, "chmod_demo.txt")
	os.WriteFile(path, []byte("权限测试\n"), 0644)

	// 查看原始权限
	info, _ := os.Stat(path)
	fmt.Printf("修改前权限：%s\n", info.Mode())

	// os.Chmod 修改文件权限
	if err := os.Chmod(path, 0755); err != nil {
		fmt.Println("Chmod 失败：", err)
		return
	}

	info, _ = os.Stat(path)
	fmt.Printf("修改后权限：%s\n", info.Mode())

	// 注意：在 Windows 上，Chmod 的效果有限
	// 只读属性会被映射，但 rwx 权限模型不完全适用
}

// ===========================================================================
// 14. 软链接 — os.Symlink / os.Readlink
// ===========================================================================
func demoSymlink(dir string) {
	target := filepath.Join(dir, "chmod_demo.txt")
	link := filepath.Join(dir, "link_to_chmod")

	// os.Symlink 创建软链接（符号链接）
	// 参数：目标路径 | 链接路径
	if err := os.Symlink(target, link); err != nil {
		// Windows 上创建软链接可能需要管理员权限
		fmt.Printf("创建软链接失败（Windows 可能需要管理员权限）：%v\n", err)
		return
	}
	fmt.Printf("已创建软链接 %s -> %s\n", link, target)

	// os.Readlink 读取软链接指向的目标
	targetPath, err := os.Readlink(link)
	if err != nil {
		fmt.Println("Readlink 失败：", err)
		return
	}
	fmt.Printf("软链接指向：%s\n", targetPath)

	// os.Stat 会跟踪软链接到目标文件
	info, _ := os.Stat(link)
	fmt.Printf("通过 Stat 查看软链接：名称=%s 大小=%d\n", info.Name(), info.Size())

	// os.Lstat 返回软链接本身（不跟踪）
	linfo, _ := os.Lstat(link)
	fmt.Printf("通过 Lstat 查看软链接：名称=%s Mode=%s\n", linfo.Name(), linfo.Mode())
}

// ===========================================================================
// 15. 临时文件 — os.CreateTemp
// ===========================================================================
func demoTempFile(dir string) {
	// os.CreateTemp 创建临时文件
	// 参数：目录（空字符串表示系统临时目录）| 文件名前缀
	// 返回：唯一的文件名（前缀 + 随机字符串）
	f, err := os.CreateTemp(dir, "temp_*.txt")
	if err != nil {
		fmt.Println("创建临时文件失败：", err)
		return
	}
	defer f.Close()
	defer os.Remove(f.Name()) // 用完删除

	fmt.Printf("创建临时文件：%s\n", f.Name())

	f.WriteString("这是临时数据\n")
	f.Sync() // 刷盘确保写入

	// os.MkdirTemp 创建临时目录
	tmpDir, err := os.MkdirTemp(dir, "tmpdir_*")
	if err != nil {
		fmt.Println("创建临时目录失败：", err)
		return
	}
	defer os.RemoveAll(tmpDir)
	fmt.Printf("创建临时目录：%s\n", tmpDir)
}

// ===========================================================================
// 16. 文件权限详解
// ===========================================================================
func demoPermExplain() {
	fmt.Println(`
权限用三位八进制数表示，每位对应一个角色：

  第1位 = 文件所有者（user）
  第2位 = 同组用户（group）
  第3位 = 其他用户（others）

每位数字由以下值相加得到：
  4 = r（read，读取）
  2 = w（write，写入/修改）
  1 = x（execute，执行）

常见权限值：
  0644  →  rw-r--r--  普通文件默认值（属主可读写，其他人只读）
  0755  →  rwxr-xr-x  可执行文件/目录（属主完全控制，其他人可读可执行）
  0600  →  rw-------  私密文件（仅属主可读写，如密钥）
  0700  →  rwx------  仅属主可执行的脚本
  0777  →  rwxrwxrwx  所有人完全访问（强烈不推荐）

实际生效权限 = 指定权限 & ~umask
  Linux 默认 umask 通常是 022，所以：
  0666 & ~022 = 0644
  0777 & ~022 = 0755

Windows 不使用 Unix 权限模型，Go 会做有限的映射。
在 Windows 上 0644 和 0755 的实际效果差异不大。
`)
}

// ===========================================================================
// 辅助函数
// ===========================================================================
func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
