# Gin Fullstack 二开说明

## 项目基本信息

- 学校：华中师范大学
- 姓名：彭鸿斌
- 学号：2024124379

## 项目说明

本项目基于 `gin-vue-admin` 进行二次开发，我在二开的过程中发现并修复了后端服务重启后登录失败的Bug，同时完成了任务二，即用户最后登录行为追踪。

- 开发问题记录： [Problem.md](./Problem.md)

## 开发任务索引

### 任务 1：SQLite 环境迁移与初始化

- 将项目数据库运行环境切换为 SQLite。
- 完成初始化页面选择 SQLite 的流程验证。
- 修复 SQLite 在 `dbPath` 为空时，初始化阶段与重启阶段 DSN 路径不一致的问题，具体的Bug发现和修复请参考[Problem.md](./Problem.md)。
- 确保服务重启后仍能连接到同一个数据库文件并正常登录。

### 任务 2：用户最后登录行为追踪

- 为用户表增加“最后登录 IP”和“最后登录时间”字段。
- 在用户成功登录后写入最后登录 IP 和最后登录时间。
- 在“用户管理”页面新增“登录IP”“登录时间”两列。
- 将登录时间统一格式化为 `YYYY-MM-DD HH:mm`。

## 核心技术实现

### 1. SQLite 初始化问题修复

在开发过程中，我发现 SQLite 初始化阶段和服务重启阶段使用了不同的 DSN 拼接方式，导致 `dbPath` 留空时，第一次初始化和后续重启访问的不是同一个数据库文件，最终出现“重启后登录失败”问题。

本次修复统一了 SQLite 路径拼接逻辑：

- `server/model/system/request/sys_init.go` 中的 `SqliteEmptyDsn()` 改为使用 `filepath.Join(...)`
- `server/config/gorm_sqlite.go` 中的 `Dsn()` 保持一致

这样可以保证：

- 初始化阶段和重启阶段访问同一份 SQLite 数据库文件
- `config.yaml` 写回空路径后，后续启动仍能正确找到数据库
- 避免因打开新空库而导致用户数据缺失、登录失败

### 2. 用户最后登录信息记录

围绕任务 2，我按“模型层 -> 业务层 -> 接口层 -> 前端展示层”的链路完成实现：

- 在 `server/model/system/sys_user.go` 中为 `SysUser` 增加最后登录 IP、最后登录时间字段[用户实体字段添加](./server/model/system/sys_user.go#L35)
- 在 `server/service/system/sys_user.go` 中新增用户最后登录信息更新方法 [用户服务方法添加](./server/service/system/sys_user.go#L63)
- 在 `server/api/v1/system/sys_user.go` 的登录成功流程中记录 `c.ClientIP()` 和当前登录时间 [登录接口添加](./server/api/v1/system/sys_user.go#L120)
- 通过 Gorm 自动迁移将新增字段同步到 `sys_users` 表
- 在 `web/src/view/superAdmin/user/user.vue` 中新增列表列并格式化时间展示 [前端展示添加](./web/src/view/superAdmin/user/user.vue#L41)

当前实现以“最后一次成功登录”为准，不把普通页面刷新或 token 自动续签当作新的登录事件。

## 运行与验收

### 启动方式

前端：

```bash
cd web
npm install
npm run dev
```

后端：

```bash
cd server
go run main.go
```

### 验收要点

- 首次初始化时可选择 SQLite 完成系统初始化
- 当初始化`dbPath`为空时，后端重启后仍可使用原账号继续登录
- 管理员进入“用户管理”页面后，可以看到新增用户的“登录IP”“登录时间”
- 登录时间显示格式为 `YYYY-MM-DD HH:mm`


# SQLite 初始化 DSN 路径不一致导致重启后登录失败Bug

## 问题描述

使用 SQLite 数据库初始化项目后，终止后端进程再次启动，登录时提示"用户名或密码错误"。

## 复现步骤

1. 启动后端，访问 `/init` 进行初始化
2. 选择 SQLite 数据库类型
3. `dbPath` 字段**留空**（前端默认值），只填写 `dbName`（如 `gva`）
4. 完成初始化，登录成功
5. **终止后端进程，重新启动**
6. 使用相同账号密码登录 → 提示"用户名不存在或者密码错误"

## 根因分析

### 初始化阶段和重启阶段 DSN 不一致

**初始化时**的 DSN 生成（[`server/model/system/request/sys_init.go:48`](./server/model/system/request/sys_init.go#L48)）：

```go
func (i *InitDB) SqliteEmptyDsn() string {
    separator := string(os.PathSeparator)
    return i.DBPath + separator + i.DBName + ".db"
}
```

当 `dbPath` 为空字符串时，拼接结果为：
- Windows: `\gva.db` → 当前盘符根目录（如 `C:\gva.db`）
- Linux: `/gva.db` → 系统根目录

**重启时**后端程序会读入`server/config.yaml`中的sqlite.path,DSN 生成代码（[`server/config/gorm_sqlite.go:12`](./server/config/gorm_sqlite.go#L12)）：

```go
func (s *Sqlite) Dsn() string {
    return filepath.Join(s.Path, s.Dbname+".db")
}
```

当 `Path` 为空时，`filepath.Join("", "gva.db")` = `gva.db` → **当前工作目录**

### config.yaml 写回结果

初始化完成后，`WriteConfig` 将空的 `dbPath` 写回 `server/config.yaml`：

```yaml
sqlite:
    path: ""        # 空！
    db-name: "gva"
```
这就导致了当使用前端页面进行初始化数据库时，实际sqlite数据库文件路径与从下一次重启服务时，从 `config.yaml` 读取的路径不一致：
| 阶段 | DSN 值 | 实际文件位置 |
|------|--------|--------------|
| 初始化 | `\gva.db` | `C:\gva.db`（盘符根目录） |
| 重启后 | `gva.db` | `server\gva.db`（当前工作目录） |

**两次打开的是不同的文件！** 重启后打开的是一个没有用户数据的全新空数据库。
导致首次通过前端浏览器访问初始化界面并选择 SQLite，当你的 `dbPath` 字段留空时，初始化的数据库文件路径为 `C:\gva.db`
而下次服务重启后由于 `config.yaml` 中的 `sqlite.path` 为空，导致 DSN 值为 `gva.db`，从而打开的是 `server\gva.db`。
这就导致了重启后登录失败的问题。其根本原因是第一次数据库初始化和第二次重启服务得到的数据库文件路径不一致。

![](./ReademeImg/用户名不存在或密码错误.png)

![](./ReademeImg/表未初始化.png)

## 影响范围

- 仅影响 SQLite 数据库初始化场景
- 仅影响 `dbPath` 字段留空的初始化方式
- MySQL/PostgreSQL 不受影响（它们的 DSN 包含完整的 host/port）

## 修复方案

修改 `server/model/system/request/sys_init.go` 的 `SqliteEmptyDsn` 方法（第一次初始化数据库时）：

```go
func (i *InitDB) SqliteEmptyDsn() string {
    return filepath.Join(i.DBPath, i.DBName+".db")
}
```

与 `server/config/gorm_sqlite.go` 的 `Dsn()` 方法保持一致（重启服务读入`config.yaml`中的sqlite.path路径时）：

```go
func (s *Sqlite) Dsn() string {
    return filepath.Join(s.Path, s.Dbname+".db")
}
```

这样，当在前端页面首次初始化SQlite数据库时，`dbPath` 为空字符串时，`SqliteEmptyDsn（）`方法也会返回`gva.db`路径。与重启时的`Dsn()`方法返回路径保持一致为`gva.db`。
成功解决重启后登录失败的Bug。
