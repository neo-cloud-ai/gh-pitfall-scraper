# gh-pitfall-scraper

一个专门用于从 GitHub Issues 中筛选和抓取「高价值」工程问题的自动化工具，特别专注于 AI/ML 框架和分布式训练领域的性能问题、崩溃问题和关键工程故障。

## 🎯 项目目标

自动抓取「高价值」工程问题，工具会自动筛选：

### 📋 支持的 Issue 类型
- **性能劣化** (Performance regression)
- **GPU OOM/碎片化**
- **CUDA kernel crash**
- **NCCL 死锁**
- **多机训练异常**
- **KV Cache/Prefill/Decode 错误**
- **FlashAttention/FlashDecoding bug**
- **推理吞吐瓶颈**
- **内存泄漏**
- **分布式 hanging**

## 🏗️ 智能评分系统

通过以下维度进行「坑价值」评分：
- **关键词匹配**: 标题和内容匹配技术坑关键词
- **GitHub Reactions**: 点赞数
- **评论量**: Issue 的讨论活跃度
- **标签**: bug/performance 标记
- **状态**: open 状态
- **时效性**: 最近更新的问题优先级更高

## 🌐 多仓库支持

支持多仓库批量抓取，目标仓库包括：
- **vLLM** (vllm-project)
- **sglang** (sgl-project)  
- **TensorRT-LLM** (NVIDIA)
- **DeepSpeed** (microsoft)
- **PyTorch** (pytorch)
- **Transformers** (huggingface)
- **Ray Serve**

## 📊 输出格式

- **JSON 格式**: 便于程序化处理
- **Markdown 格式**: 用于写书和文档编写

每个输出的 Issue 包含：
- Issue Title
- 链接
- 标签 (crash/perf/oom)
- 原因总结
- 复现条件
- 可能影响等信息

## 🚀 快速开始

### 1. 环境要求

- Go 1.21 或更高版本
- GitHub Personal Access Token

### 2. 安装

```bash
git clone https://github.com/neo-cloud-ai/gh-pitfall-scraper.git
cd gh-pitfall-scraper
go mod tidy
```

### 3. 配置

编辑 `config.yaml` 文件，设置您的 GitHub Token：

```yaml
github_token: "ghp_your_github_token_here"

repos:
  - owner: vllm-project
    name: vllm
  - owner: sgl-project
    name: sglang
  - owner: NVIDIA
    name: TensorRT-LLM
  - owner: microsoft
    name: DeepSpeed
  - owner: pytorch
    name: pytorch
  - owner: huggingface
    name: transformers

keywords:
  - "performance"
  - "regression"
  - "latency"
  - "throughput"
  - "OOM"
  - "memory leak"
  - "CUDA"
  - "kernel"
  - "NCCL"
  - "hang"
  - "deadlock"
  - "kv cache"
```

### 4. 运行

```bash
go run main.go
```

程序将自动：
1. 读取配置文件
2. 抓取指定仓库的 Issues
3. 应用智能评分算法
4. 过滤高价值问题
5. 输出结果到 `output/issues.json`

## 📁 项目结构

```
gh-pitfall-scraper/
├── main.go                    # 主程序入口
├── config.yaml               # 配置文件
├── .gitignore               # Git 忽略文件
├── go.mod                   # Go 模块文件
└── internal/
    └── scraper/
        ├── github.go         # GitHub API 客户端
        ├── scorer.go         # 智能评分系统
        ├── filter.go         # 问题过滤逻辑
        ├── scrape.go         # 数据抓取逻辑
        └── scraper_test.go   # 单元测试
```

## 🔧 核心模块

### 1. GitHub 客户端 (github.go)
- 处理 GitHub API 认证
- 批量获取 Issues 数据
- 错误处理和重试机制

### 2. 智能评分系统 (scorer.go)
- 多维度评分算法
- 关键词权重配置
- 社区参与度评估

### 3. 过滤系统 (filter.go)
- 基础过滤条件
- 高级过滤规则
- 黑名单/白名单机制

### 4. 抓取引擎 (scrape.go)
- 仓库批量抓取
- 并发控制
- 统计信息生成

## 🧪 测试

运行单元测试：

```bash
go test ./internal/scraper/...
```

运行所有测试：

```bash
go test ./...
```

## 📈 应用场景

### ✅ 适用场景
- **技术写作**: 收集工程问题用于技术书籍编写
- **技术调研**: 了解 AI/ML 框架的常见问题和解决方案
- **技术选型**: 评估不同框架的稳定性和成熟度
- **知识管理**: 建立技术问题的知识库和最佳实践

### 🎯 目标用户
- 技术写作者
- AI/ML 工程师
- 技术团队负责人
- 技术架构师

## 📊 输出示例

```json
[
  {
    "id": 12345,
    "number": 678,
    "title": "Performance regression in vLLM inference after CUDA upgrade",
    "url": "https://github.com/vllm-project/vllm/issues/678",
    "state": "open",
    "labels": [
      {
        "name": "bug",
        "color": "d73a4a",
        "description": "Something isn't working"
      }
    ],
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-20T14:22:00Z",
    "keywords": ["performance", "regression", "CUDA"],
    "score": 28.5,
    "comments": 25,
    "reactions": 42,
    "assignee": "developer_name",
    "milestone": "v0.12.0",
    "body": "Detailed issue description...",
    "repo_owner": "vllm-project",
    "repo_name": "vllm"
  }
]
```

## 🔮 未来计划

- [ ] Web 界面支持
- [ ] 定时任务功能
- [ ] 更多 AI/ML 框架支持
- [ ] 机器学习优化评分算法
- [ ] API 服务模式
- [ ] 导出多种格式 (CSV, PDF)
- [ ] 问题分类和标签系统

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

MIT License

## ⚠️ 注意事项

1. **Rate Limiting**: GitHub API 有请求限制，请合理设置并发数
2. **Token Security**: 不要将 GitHub Token 提交到版本控制系统
3. **Data Usage**: 抓取的数据仅用于学习和研究目的
4. **API Terms**: 请遵守 GitHub API 的使用条款

---

**Built with ❤️ for the AI/ML community**