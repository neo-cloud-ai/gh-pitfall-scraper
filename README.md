@ -0,0 +1,69 @@
# 🚀 目标：自动抓取「高价值」工程问题

工具会自动筛选：

* 性能劣化（Performance regression）
* GPU OOM/碎片化
* CUDA kernel crash
* NCCL 死锁
* 多机训练异常
* KV Cache/Prefill/Decode 错误
* FlashAttention/FlashDecoding bug
* 推理吞吐瓶颈
* 内存泄漏（memory leak）
* 分布式 hanging

通过关键词 + Github Reactions + Comment 量来打分。

---

# 🧱 **项目结构设计**

```
gh-pitfall-scraper/
│── main.go
│── config.yaml
│── internal/
│     ├── client/
│     │     └── github.go
│     ├── scraper/
│     │     ├── filter.go
│     │     ├── scorer.go
│     │     └── scrape.go
│     └── model/
│           ├── issue.go
│           └── repo.go
│── output/
│     └── issues.json
```

---

# 🎯 **核心能力**

### ✔ 1. 支持多仓库批量抓取

vLLM / sglang / TensorRT-LLM / DeepSpeed / PyTorch / Transformers / Ray Serve

### ✔ 2. 智能“坑价值”评分

评分维度：

* 标题和内容是否匹配技术坑关键词
* 是否有人点赞（reactions）
* 评论量
* 是否被标记为 bug/performance
* 是否 open 状态（未完全解决）

### ✔ 3. 输出 Markdown / JSON 用于写书

自动生成：

```
# Issue Title  
Link:  
标签: crash / perf / oom  
原因总结：  
复现条件：  
可能影响：  
```