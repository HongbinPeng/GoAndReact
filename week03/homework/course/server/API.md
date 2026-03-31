# 接口文档

> 基础地址：`http://localhost:3000`
> 认证方式：除登录接口外，所有接口需在请求头中携带 `Authorization: Bearer <token>`

## 统一响应格式

所有 JSON 接口均采用以下统一格式返回：

```json
{
  "code": 0,
  "msg": "success",
  "data": ...
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| code | number | 状态码，`0` 表示成功，非 `0` 为 HTTP 错误状态码 |
| msg | string | 提示信息，成功时为 `"success"`，失败时为具体错误描述 |
| data | any | 业务数据，失败时为 `null` |

**错误响应示例**:

```json
{
  "code": 400,
  "msg": "课程名称不能为空",
  "data": null
}
```

---

## 一、认证模块

### 1.1 登录

- **URL**: `POST /api/auth/login`
- **请求体**:

```json
{
  "username": "admin",
  "password": "admin123"
}
```

- **成功响应** (200):

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "user": {
      "id": 1,
      "username": "admin",
      "name": "管理员",
      "role": "admin",
      "avatar": "",
      "created_at": "2024-01-01 00:00:00"
    }
  }
}
```

- **失败响应** (401):

```json
{
  "code": 401,
  "msg": "用户名或密码错误",
  "data": null
}
```

### 1.2 获取当前用户信息

- **URL**: `GET /api/auth/me`
- **成功响应** (200):

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "id": 1,
    "username": "admin",
    "name": "管理员",
    "role": "admin",
    "avatar": "",
    "created_at": "2024-01-01 00:00:00"
  }
}
```

---

## 二、工作台

### 2.1 获取工作台数据

- **URL**: `GET /api/dashboard`
- **成功响应** (200):

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "stats": {
      "totalCourses": 8,
      "publishedCourses": 7,
      "totalStudents": 20,
      "activeStudents": 18
    },
    "charts": {
      "enrollment": [
        { "name": "Git 版本控制", "value": 50 },
        { "name": "Vue 3 实战项目", "value": 45 }
      ],
      "activity": [
        { "date": "2024-03-01", "label": "周一", "students": 8, "duration": 12 },
        { "date": "2024-03-02", "label": "周二", "students": 10, "duration": 15 }
      ],
      "statusDist": [
        { "name": "活跃学生", "value": 18 },
        { "name": "非活跃学生", "value": 2 }
      ],
      "categoryDist": [
        { "name": "前端开发", "value": 3 },
        { "name": "后端开发", "value": 1 }
      ]
    }
  }
}
```

**字段说明**:
- `stats` — 统计概览
  - `totalCourses` — 课程总数
  - `publishedCourses` — 已发布课程数
  - `totalStudents` — 学生总数
  - `activeStudents` — 活跃学生数
- `charts.enrollment` — 课程选课人数（柱状图）
- `charts.activity` — 近 7 天学习活跃度（折线图）
  - `students` — 当天学习人数
  - `duration` — 当天总学习时长（小时）
- `charts.statusDist` — 学生状态分布（饼图）
- `charts.categoryDist` — 课程分类分布（饼图）

---

## 三、课程管理

### 3.1 获取课程列表

- **URL**: `GET /api/courses`
- **查询参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| keyword | string | 否 | 搜索关键字（课程名/讲师） |
| status | string | 否 | 状态筛选：`published` / `draft` |
| category | string | 否 | 分类筛选 |
| sortField | string | 否 | 排序字段，可选值：`student_count` / `lesson_count` / `created_at` / `name` |
| sortOrder | string | 否 | 排序方向：`ascend`（升序） / `descend`（降序） |
| page | number | 否 | 页码，默认 1 |
| pageSize | number | 否 | 每页数量，默认 10 |

- **成功响应** (200):

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "list": [
      {
        "id": 1,
        "name": "React 基础入门",
        "description": "从零开始学习 React 框架...",
        "instructor": "张老师",
        "cover": "",
        "category": "前端开发",
        "status": "published",
        "student_count": 32,
        "lesson_count": 12,
        "created_at": "2024-01-01 00:00:00",
        "updated_at": "2024-01-01 00:00:00"
      }
    ],
    "total": 8,
    "page": 1,
    "pageSize": 10
  }
}
```

### 3.2 获取课程分类列表

- **URL**: `GET /api/courses/categories`
- **成功响应** (200):

```json
{
  "code": 0,
  "msg": "success",
  "data": ["前端开发", "后端开发", "数据库", "数据科学", "工具", "运维"]
}
```

### 3.3 获取课程详情

- **URL**: `GET /api/courses/:id`
- **成功响应** (200): `data` 结构同列表项

### 3.4 创建课程

- **URL**: `POST /api/courses`
- **请求体**:

```json
{
  "name": "课程名称",
  "description": "课程描述",
  "instructor": "讲师名",
  "category": "前端开发",
  "status": "draft",
  "lesson_count": 10
}
```

- **成功响应** (201): `data` 为创建的课程对象

### 3.5 更新课程

- **URL**: `PUT /api/courses/:id`
- **请求体**: 同创建，所有字段可选
- **成功响应** (200): `data` 为更新后的课程对象

### 3.6 删除课程

- **URL**: `DELETE /api/courses/:id`
- **成功响应** (200):

```json
{
  "code": 0,
  "msg": "删除成功",
  "data": null
}
```

### 3.7 切换课程状态

- **URL**: `PATCH /api/courses/:id/status`
- **说明**: 在 `published` 和 `draft` 之间切换
- **成功响应** (200): `data` 为更新后的课程对象

---

## 四、学生管理

### 4.1 获取学生列表

- **URL**: `GET /api/students`
- **查询参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| keyword | string | 否 | 搜索关键字（姓名/学号） |
| className | string | 否 | 班级筛选 |
| status | string | 否 | 状态：`active` / `inactive` |
| courseId | number | 否 | 按选课筛选 |
| page | number | 否 | 页码，默认 1 |
| pageSize | number | 否 | 每页数量，默认 10 |

- **成功响应** (200):

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "list": [
      {
        "id": 1,
        "name": "陈明远",
        "student_no": "20240001",
        "class_name": "前端2401班",
        "phone": "13800000001",
        "email": "student1@example.com",
        "status": "active",
        "course_ids": [1, 3, 5],
        "created_at": "2024-01-01 00:00:00",
        "updated_at": "2024-01-01 00:00:00"
      }
    ],
    "total": 20,
    "page": 1,
    "pageSize": 10
  }
}
```

### 4.2 获取班级列表

- **URL**: `GET /api/students/classes`
- **成功响应** (200):

```json
{
  "code": 0,
  "msg": "success",
  "data": ["前端2401班", "前端2402班", "后端2401班", "全栈2401班"]
}
```

### 4.3 获取学生详情

- **URL**: `GET /api/students/:id`
- **成功响应** (200):

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "id": 1,
    "name": "陈明远",
    "student_no": "20240001",
    "class_name": "前端2401班",
    "phone": "13800000001",
    "email": "student1@example.com",
    "status": "active",
    "course_ids": [1, 3, 5],
    "enrolledCourses": [
      { "id": 1, "name": "React 基础入门" },
      { "id": 3, "name": "Vue 3 实战项目" },
      { "id": 5, "name": "MySQL 数据库设计" }
    ],
    "created_at": "2024-01-01 00:00:00",
    "updated_at": "2024-01-01 00:00:00"
  }
}
```

### 4.4 创建学生

<font color=e32d40>**TODO: 按照下面要求完善接口**</font>

- **URL**: `POST /api/students`
- **请求体**:

```json
{
  "name": "学生姓名",
  "student_no": "20240021",
  "class_name": "前端2401班",
  "phone": "13800000000",
  "email": "student@example.com",
  "status": "active",
  "course_ids": [1, 2]
}
```

- **成功响应** (201): `data` 为创建的学生对象

### 4.5 更新学生

<font color=e32d40>**TODO: 按照下面要求完善接口**</font>

- **URL**: `PUT /api/students/:id`
- **请求体**: 同创建，所有字段可选
- **成功响应** (200): `data` 为更新后的学生对象

### 4.6 删除学生

<font color=e32d40>**TODO: 按照下面要求完善接口**</font>

- **URL**: `DELETE /api/students/:id`
- **成功响应** (200):

```json
{
  "code": 0,
  "msg": "删除成功",
  "data": null
}
```

---

## 五、学习总结

### 5.1 获取学习总结内容

- **URL**: `GET /api/summary`
- **成功响应** (200):

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "content": "# 前端开发学习总结\n\n## 一、React 基础\n..."
  }
}
```

- **说明**: 返回 Markdown 格式的文本内容，前端需解析并渲染

---

## 六、静态资源

### 6.1 获取静态文件

- **URL**: `GET /api/static/:path`
- **认证**: 无需认证
- **说明**: 返回 `server/data/` 目录下的静态资源文件，支持子目录路径。主要用于学习总结中 Markdown 引用的图片。
- **支持的文件类型**: `.png`、`.jpg`、`.jpeg`、`.gif`、`.webp`、`.svg`
- **示例**: `GET /api/static/assets/flower.png`
- **成功响应** (200): 返回对应 MIME 类型的文件内容（二进制，不使用统一 JSON 格式）
- **失败响应**:

| 状态码 | 响应示例 |
|--------|---------|
| 400 | `{ "code": 400, "msg": "非法路径", "data": null }` |
| 403 | `{ "code": 403, "msg": "不支持的文件类型", "data": null }` |
| 404 | `{ "code": 404, "msg": "文件不存在", "data": null }` |

---

## 通用错误响应

| 状态码 | 说明 |
|--------|------|
| 400 | 请求参数错误 |
| 401 | 未认证或令牌过期 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |

错误响应格式:

```json
{
  "code": 400,
  "msg": "错误描述信息",
  "data": null
}
```
