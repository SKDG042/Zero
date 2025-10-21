# Zero Agent

施工中ing

## 快速开始

### 1. 克隆仓库

```bash
git clone https://github.com/SKDG042/Zero.git
cd Zero
```

### 2. 安装依赖

```bash
go mod tidy
```

### 3. 运行项目

```bash
go run cmd/main.go
```

### 4. 打包项目

```bash
# Windows
go build -o Zero.exe ./cmd

# Linux/macOS
go build -o Zero ./cmd
```

## 项目结构

```
Zero/
├── cmd/           # 主程序入口
├── ui/            # GUI 界面
├── llm/           # LLM 客户端
├── session/       # 会话管理
├── devops/        # DevOps 工具(coze)
├── utils/         # 工具函数
├── assets/        # 资源文件
└── docs/          # 文档
```

## 依赖

- Fyne v2.7.0

## 许可证

- MIT
