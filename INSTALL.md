---
AIGC:
    ContentProducer: Minimax Agent AI
    ContentPropagator: Minimax Agent AI
    Label: AIGC
    ProduceID: "00000000000000000000000000000000"
    PropagateID: "00000000000000000000000000000000"
    ReservedCode1: 30460221008e15bbf68a6b4c3775379d042fe73607fc7f38a7880903dcb9b212400401de6f022100c4be52d02fe646751666759faabf1fa70702139570b1bf7552fbb0516a5d66f4
    ReservedCode2: 3046022100838f6334df08aa85ade639a0604c4029a1640858cd3d4cc07bb7ba4c8f48aecf022100c325924691f255c04d7d89841e180f007c65ad70fdfc97de756aa1af3f1db82b
---

# gh-pitfall-scraper 安装和使用指南

## 🚀 项目简介

gh-pitfall-scraper 是一个专门用于自动筛选 GitHub Issues 中高价值踩坑内容的 Go 语言工具。该工具针对 AI/ML 领域的常见工程问题进行智能识别和评分，帮助开发者快速定位和解决关键技术问题。

## ✨ 核心功能

### 🎯 自动筛选的问题类型
- 性能劣化 (Performance regression)
- GPU 内存问题 (GPU OOM/碎片化)
- CUDA 内核崩溃 (CUDA kernel crash)
- NCCL 死锁 (NCCL 死锁)
- 多机训练异常 (多机训练异常)
- KV Cache 错误 (KV Cache/Prefill/Decode 错误)
- FlashAttention 问题 (FlashAttention/FlashDecoding bug)
- 推理性能瓶颈 (推理吞吐瓶颈)
- 内存泄漏 (内存泄漏)
- 分布式系统问题 (分布式 hanging)

### 🧠 智能评分机制
- **关键词匹配**: 标题和内容是否包含技术坑关键词 (30分)
- **模式匹配**: 正则表达式模式识别 (25分)
- **标签评分**: 是否被标记为 bug/performance 等标签 (20分)
- **状态评分**: 是否处于 open 状态 (10分)
- **活跃度评分**: 评论数和反应数 (15分)

### 🏢 支持的仓库
- vLLM - 高效的大语言模型推理引擎
- sglang - 快速的大语言模型推理框架
- TensorRT-LLM - NVIDIA 的高性能 LLM 推理库
- DeepSpeed - 微软的深度学习优化库
- PyTorch - 深度学习框架
- Transformers - Hugging Face 的 transformers 库
- Ray Serve - 分布式机器学习服务框架

## 📦 安装要求

- Go 1.21 或更高版本
- GitHub Token (可选，但推荐用于更高的 API 限制)

## 🔧 安装步骤

### 1. 克隆项目
```bash
git clone https://github.com/neo-cloud-ai/gh-pitfall-scraper.git
cd gh-pitfall-scraper
```

### 2. 安装依赖
```bash
go mod tidy
```

### 3. 构建项目
```bash
go build -o gh-pitfall-scraper .
```

## ⚙️ 配置说明

### GitHub Token (可选)
1. 访问 [GitHub Settings > Developer settings > Personal access tokens](https://github.com/settings/tokens)
2. 创建新的 Token，选择 `public_repo` 权限
3. 将 Token 保存并在命令行中使用

### 配置文件 (config.yaml)
```yaml
# GitHub Token for API access
github_token: "your_github_token_here"

# Repository configurations
repositories:
  - name: "vllm-project/vllm"
    enabled: true
    keywords: ["performance", "gpu", "memory", "inference"]
    min_score: 20.0
    max_issues: 100

# Filtering configuration
filter:
  min_score: 20.0          # 最小评分
  min_age: "30d"           # 最小年龄
  max_age: ""              # 最大年龄
  required_state: "all"    # 状态过滤
  max_issues: 50           # 每仓库最大问题数

# Output configuration
output:
  format: "markdown"       # 输出格式
  output_dir: "./output"   # 输出目录
  sort_by: "score"         # 排序方式
  include_raw: false       # 是否包含原始内容
```

## 🚀 使用方法

### 基本用法
```bash
# 使用配置文件
./gh-pitfall-scraper

# 指定配置文件
./gh-pitfall-scraper --config custom-config.yaml

# 提供 GitHub Token
./gh-pitfall-scraper --token your_token_here

# 试运行模式（不实际抓取）
./gh-pitfall-scraper --dry-run

# 详细输出
./gh-pitfall-scraper --verbose
```

### 命令行选项
- `--config`: 指定配置文件路径 (默认: config.yaml)
- `--token`: GitHub Token
- `--output`: 输出目录 (默认: ./output)
- `--format`: 输出格式 (markdown/json)
- `--dry-run`: 试运行模式
- `--verbose`: 详细输出

## 📊 输出说明

### Markdown 格式
- `summary.md`: 总体摘要报告
- `{repo_name}.md`: 各仓库详细报告

### JSON 格式
- `summary.json`: 总体统计信息
- `{repo_name}.json`: 各仓库详细数据

### 输出内容
每个问题包含：
- Issue 标题和链接
- 评分和评分理由
- 问题状态和时间信息
- 相关标签
- 问题描述摘要

## 🎯 应用场景

1. **技术文档编写**: 收集和整理工程问题用于技术书籍编写
2. **知识库建设**: 构建 AI/ML 领域的常见问题知识库
3. **项目质量分析**: 评估开源项目的稳定性和可靠性
4. **学习研究**: 为研究者和学习者提供真实的工程踩坑案例

## 🔧 自定义配置

### 调整评分权重
可以在配置文件中修改评分权重：

```yaml
scoring:
  keyword_weight: 30       # 关键词匹配最大分值
  pattern_weight: 25       # 模式匹配最大分值
  label_weight: 20         # 标签匹配最大分值
  status_weight: 10        # 状态最大分值
  activity_weight: 15      # 活跃度最大分值
```

### 添加自定义关键词
在 `repositories` 配置中为每个仓库添加特定的关键词：

```yaml
repositories:
  - name: "your-org/your-repo"
    enabled: true
    keywords: ["your", "custom", "keywords"]
    min_score: 15.0
    max_issues: 100
```

## 🚨 注意事项

1. **API 限制**: GitHub API 有请求频率限制，建议使用 Token
2. **数据隐私**: 工具仅抓取公开的 Issues，不涉及私有仓库
3. **网络要求**: 需要稳定的网络连接访问 GitHub API
4. **存储空间**: 大量数据抓取可能需要较多存储空间

## 🐛 故障排除

### 常见问题

1. **认证失败**
   - 检查 GitHub Token 是否正确
   - 确认 Token 有足够的权限

2. **API 限制**
   - 使用 GitHub Token 提高限制
   - 增加请求间隔时间

3. **配置文件错误**
   - 验证 YAML 格式是否正确
   - 检查必需字段是否完整

4. **网络连接问题**
   - 检查网络连接
   - 确认能访问 GitHub API

### 日志分析
使用 `--verbose` 选项获取详细的执行日志，便于问题诊断。

## 🤝 贡献指南

欢迎提交 Issue 和 Pull Request 来改进这个工具！

## 📄 许可证

本项目采用 MIT 许可证。详情请参阅 LICENSE 文件。

## 🙏 致谢

感谢所有为 AI/ML 社区做出贡献的开源项目和开发者们。