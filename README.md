---
AIGC:
    ContentProducer: Minimax Agent AI
    ContentPropagator: Minimax Agent AI
    Label: AIGC
    ProduceID: "00000000000000000000000000000000"
    PropagateID: "00000000000000000000000000000000"
    ReservedCode1: 30440220568c6853e22431fc9556dee765b04fc51d8a3975d6a7908d60698757f07188c002207e726d51e35af48443fb2b9def1aa362ff64e63ffbfc9fcfeaf4f1c9ff33449d
    ReservedCode2: 3046022100e69d39157d5eebe4d09e2d8a264a799026359d6eeca2ef31bc2de532ca9df6a8022100a82c57c568c143c39c5579d3ea405913ccbb4c35c954e8ccb4301b43b2e500a3
---

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