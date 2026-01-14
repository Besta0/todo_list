# Todo List CLI

一个使用 Go 语言实现的简单、高效的命令行待办事项管理工具。

## 特性

- ✅ 添加待办事项
- 📋 查看所有待办事项
- ✓ 标记任务为已完成
- 🗑️ 删除待办事项
- 💾 自动持久化到本地文件
- 🎯 简洁的命令行界面
- 🔒 数据完整性保证

## 安装

### 从源码构建

确保已安装 Go 1.24.5 或更高版本：

```bash
# 克隆仓库
git clone <repository-url>
cd todolist

# 构建可执行文件
go build -o todolist ./cmd/todolist

# (可选) 将可执行文件移动到 PATH 中
sudo mv todolist /usr/local/bin/
```

### 依赖

- Go 1.24.5+
- [gopter](https://github.com/leanovate/gopter) - 属性测试库（仅开发时需要）

## 使用方法

### 基本命令

```bash
# 显示帮助信息
todolist help

# 添加新任务
todolist add <任务描述>

# 查看所有任务
todolist list

# 标记任务为已完成
todolist done <任务ID>

# 删除任务
todolist delete <任务ID>
```

### 使用示例

#### 1. 添加任务

```bash
$ todolist add "学习 Go 语言"
✓ Task added: [1] 学习 Go 语言

$ todolist add "完成项目文档"
✓ Task added: [2] 完成项目文档

$ todolist add "准备周会演示"
✓ Task added: [3] 准备周会演示
```

#### 2. 查看任务列表

```bash
$ todolist list
Your tasks:
[ ] [1] 学习 Go 语言 (created: 2026-01-14 10:30:00)
[ ] [2] 完成项目文档 (created: 2026-01-14 10:31:15)
[ ] [3] 准备周会演示 (created: 2026-01-14 10:32:00)
```

#### 3. 完成任务

```bash
$ todolist done 1
✓ Task 1 marked as completed

$ todolist list
Your tasks:
[✓] [1] 学习 Go 语言 (created: 2026-01-14 10:30:00)
[ ] [2] 完成项目文档 (created: 2026-01-14 10:31:15)
[ ] [3] 准备周会演示 (created: 2026-01-14 10:32:00)
```

#### 4. 删除任务

```bash
$ todolist delete 2
✓ Task 2 deleted

$ todolist list
Your tasks:
[✓] [1] 学习 Go 语言 (created: 2026-01-14 10:30:00)
[ ] [3] 准备周会演示 (created: 2026-01-14 10:32:00)
```

#### 5. 空列表提示

```bash
$ todolist list
No tasks found. Add a task with: todolist add <description>
```

### 常见使用场景

#### 日常任务管理

```bash
# 早上添加今天的任务
todolist add "回复邮件"
todolist add "参加团队会议"
todolist add "代码审查"

# 查看待办事项
todolist list

# 完成任务后标记
todolist done 1
todolist done 2

# 删除不需要的任务
todolist delete 3
```

#### 项目任务跟踪

```bash
# 添加项目相关任务
todolist add "设计数据库架构"
todolist add "实现 API 接口"
todolist add "编写单元测试"
todolist add "部署到测试环境"

# 按顺序完成
todolist done 1
todolist done 2
# ... 继续工作
```

## 数据存储

所有任务数据自动保存到 `~/.todolist.json` 文件中。数据格式为 JSON，便于备份和迁移。

### 数据文件示例

```json
{
  "tasks": [
    {
      "id": 1,
      "description": "学习 Go 语言",
      "completed": true,
      "created_at": "2026-01-14T10:30:00Z"
    },
    {
      "id": 3,
      "description": "准备周会演示",
      "completed": false,
      "created_at": "2026-01-14T10:32:00Z"
    }
  ],
  "next_id": 4
}
```

### 备份和恢复

```bash
# 备份任务数据
cp ~/.todolist.json ~/.todolist.backup.json

# 恢复任务数据
cp ~/.todolist.backup.json ~/.todolist.json

# 清空所有任务（重新开始）
rm ~/.todolist.json
```

## 错误处理

程序会对常见错误提供清晰的提示：

```bash
# 空任务描述
$ todolist add ""
Error: task description cannot be empty

# 无效的任务 ID
$ todolist done 999
Error: task not found

# 无效的命令
$ todolist invalid
Error: invalid command

Use 'todolist help' for usage information.
```

## 项目结构

```
.
├── cmd/
│   └── todolist/          # CLI 入口点
│       └── main.go
├── internal/
│   ├── cli/               # 命令行解析和执行
│   │   └── cli.go
│   ├── errors/            # 错误定义
│   │   └── errors.go
│   ├── models/            # 数据模型
│   │   ├── models.go
│   │   └── models_test.go
│   ├── storage/           # 存储层
│   │   ├── storage.go
│   │   └── storage_test.go
│   └── todolist/          # 业务逻辑层
│       ├── todolist.go
│       └── todolist_test.go
├── go.mod
├── go.sum
└── README.md
```

## 架构设计

系统采用三层架构，确保代码的可维护性和可测试性：

### 1. CLI Layer (`internal/cli`)
- 命令行参数解析
- 用户输入验证
- 输出格式化
- 错误消息显示

### 2. Business Logic Layer (`internal/todolist`)
- 任务管理核心逻辑
- 业务规则验证
- 数据完整性保证
- 与存储层交互

### 3. Storage Layer (`internal/storage`)
- JSON 文件读写
- 数据序列化/反序列化
- 原子写入保证
- 错误处理

### 数据流

```
用户输入 → CLI 解析 → 业务逻辑 → 存储层 → 文件系统
                ↓           ↓          ↓
            验证命令    执行操作    持久化数据
```

## 开发

### 运行测试

```bash
# 运行所有测试
go test ./...

# 运行测试并显示覆盖率
go test -cover ./...

# 运行特定包的测试
go test ./internal/todolist

# 运行属性测试（详细输出）
go test -v ./internal/todolist -run Property
```

### 测试策略

项目采用双重测试方法：

#### 单元测试
- 验证特定功能和边缘情况
- 测试错误处理路径
- 验证组件集成

#### 属性测试（Property-Based Testing）
- 使用 [gopter](https://github.com/leanovate/gopter) 库
- 验证通用正确性属性
- 每个属性测试运行 100+ 次迭代
- 覆盖 13 个核心正确性属性

### 构建

```bash
# 开发构建
go build -o todolist ./cmd/todolist

# 生产构建（优化）
go build -ldflags="-s -w" -o todolist ./cmd/todolist

# 跨平台构建
GOOS=linux GOARCH=amd64 go build -o todolist-linux ./cmd/todolist
GOOS=darwin GOARCH=amd64 go build -o todolist-macos ./cmd/todolist
GOOS=windows GOARCH=amd64 go build -o todolist.exe ./cmd/todolist
```

## 设计原则

- **简单性**: 使用 Go 标准库，避免不必要的依赖
- **可测试性**: 核心逻辑与 I/O 操作分离
- **数据完整性**: 所有操作立即持久化，使用原子写入
- **用户友好**: 清晰的命令和错误提示
- **可维护性**: 清晰的分层架构和代码组织

## 贡献

欢迎提交 Issue 和 Pull Request！

## 许可证

MIT License

## 作者

Todo List CLI 项目团队
