---
AIGC:
    ContentProducer: Minimax Agent AI
    ContentPropagator: Minimax Agent AI
    Label: AIGC
    ProduceID: "00000000000000000000000000000000"
    PropagateID: "00000000000000000000000000000000"
    ReservedCode1: 3046022100c1fb96c9a8e4d7a4ff5522afd870698e3b56c69819e292647a9f871bc64e9fc1022100fbd0450248a7cc5a176e7c49a96426b508db0d0b62e444278ec5925bc6f0d6d3
    ReservedCode2: 30450221008fa477cb541d124d51882d379610ac669b19144aebac8ccb1eb583c62effeb1e02206f1b375233fd09da7d3ff9a89bce65954d946c2e86750adf76c6c811af7850ab
---

# 示例输出文件

## 示例: summary.md

```markdown
# GitHub Issues 踩坑报告摘要

## 📊 统计概览

- **抓取时间**: 2025-12-11 23:51:39
- **总计问题数**: 156
- **涉及仓库数**: 7

## 🏢 仓库统计

- **vllm-project/vllm**: 32 个高价值问题
- **sgl-project/sglang**: 28 个高价值问题
- **NVIDIA/TensorRT-LLM**: 25 个高价值问题
- **microsoft/DeepSpeed**: 22 个高价值问题
- **pytorch/pytorch**: 19 个高价值问题
- **huggingface/transformers**: 17 个高价值问题
- **ray-project/ray**: 13 个高价值问题

## 🎯 高价值问题类别分布

- **性能问题**: 45 个问题
- **GPU内存问题**: 38 个问题
- **分布式训练**: 31 个问题
- **模型推理**: 24 个问题
- **崩溃错误**: 12 个问题
- **内存泄漏**: 6 个问题

## 📋 详细报告

- [vllm-project_vllm](./vllm-project_vllm.md)
- [sgl-project_sglang](./sgl-project_sglang.md)
- [NVIDIA_TensorRT-LLM](./NVIDIA_TensorRT-LLM.md)
- [microsoft_DeepSpeed](./microsoft_DeepSpeed.md)
- [pytorch_pytorch](./pytorch_pytorch.md)
- [huggingface_transformers](./huggingface_transformers.md)
- [ray-project_ray](./ray-project_ray.md)

---

*报告由 gh-pitfall-scraper 自动生成*
```

## 示例: vllm-project_vllm.md

```markdown
# vllm-project/vllm - 高价值工程问题报告

## 📈 问题概览

- **生成时间**: 2025-12-11 23:51:39
- **问题总数**: 32
- **平均评分**: 67.5

---

## 1. Performance regression in GPU memory usage after v0.4.0

**链接**: [https://github.com/vllm-project/vllm/issues/2847](https://github.com/vllm-project/vllm/issues/2847)  
**评分**: 85.0/100  
**状态**: open  
**创建时间**: 2025-12-01  
**更新时间**: 2025-12-09  

**标签**: [bug](https://github.com/vllm-project/vllm/labels/bug) [performance](https://github.com/vllm-project/vllm/labels/performance)

**评分理由**:
- 关键词匹配: 25.0分
- 模式匹配: 20.0分
- 标签匹配: 15.0分
- 状态评分: 10.0分
- 活跃度评分: 15.0分

**问题描述**:
```
After upgrading to v0.4.0, we're seeing significant memory usage increase during inference. The GPU memory consumption has increased by approximately 30% compared to v0.3.0 when processing the same batch size and model configuration.

Environment:
- CUDA 12.1
- A100 80GB
- Model: llama-2-70b
- Batch size: 16
- Sequence length: 2048

This is blocking our production deployment...
```

---

## 2. CUDA kernel crash when using flash attention with large batch sizes

**链接**: [https://github.com/vllm-project/vllm/issues/2756](https://github.com/vllm-project/vllm/issues/2756)  
**评分**: 82.0/100  
**状态**: open  
**创建时间**: 2025-11-28  
**更新时间**: 2025-12-10  

**标签**: [critical](https://github.com/vllm-project/vllm/labels/critical) [cuda](https://github.com/vllm-project/vllm/labels/cuda)

**评分理由**:
- 关键词匹配: 30.0分
- 模式匹配: 25.0分
- 标签匹配: 12.0分
- 状态评分: 10.0分
- 活跃度评分: 5.0分

**问题描述**:
```
The application crashes with CUDA error when batch size exceeds 32 when using Flash Attention. Error message:

CUDA kernel launch failed: misaligned address
CUDA error: misaligned address

This happens consistently with:
- Model: mixtral-8x7b-instruct
- Flash attention enabled
- Batch size > 32
- Sequence length > 1024
```

---

## 3. Memory leak in distributed training mode

**链接**: [https://github.com/vllm-project/vllm/issues/2691](https://github.com/vllm-project/vllm/issues/2691)  
**评分**: 78.0/100  
**状态**: open  
**创建时间**: 2025-11-25  
**更新时间**: 2025-12-08  

**标签**: [bug](https://github.com/vllm-project/vllm/labels/bug) [distributed](https://github.com/vllm-project/vllm/labels/distributed)

**评分理由**:
- 关键词匹配: 28.0分
- 模式匹配: 20.0分
- 标签匹配: 10.0分
- 状态评分: 10.0分
- 活跃度评分: 10.0分

**问题描述**:
```
Memory usage keeps increasing during multi-node training. The memory leak is more severe when using tensor parallelism across multiple GPUs.

Environment:
- 4x A100 80GB
- NCCL backend
- Ray cluster
- Model: llama-70b

After 2-3 hours of training, we see consistent 2-3GB memory increase per GPU...
```

---

*报告由 gh-pitfall-scraper 生成于 2025-12-11 23:51:39*
```

## 示例: summary.json

```json
{
  "generated_at": "2025-12-11T23:51:39Z",
  "total_repos": 7,
  "total_issues": 156,
  "repository_stats": {
    "vllm-project/vllm": {
      "issue_count": 32,
      "avg_score": 67.5
    },
    "sgl-project/sglang": {
      "issue_count": 28,
      "avg_score": 64.2
    },
    "NVIDIA/TensorRT-LLM": {
      "issue_count": 25,
      "avg_score": 71.8
    },
    "microsoft/DeepSpeed": {
      "issue_count": 22,
      "avg_score": 69.3
    },
    "pytorch/pytorch": {
      "issue_count": 19,
      "avg_score": 58.7
    },
    "huggingface/transformers": {
      "issue_count": 17,
      "avg_score": 55.1
    },
    "ray-project/ray": {
      "issue_count": 13,
      "avg_score": 62.4
    }
  }
}
```

## 示例: scraping_summary.txt

```
GitHub Issues 踩坑内容抓取报告
=====================================

抓取时间: 2025-12-11 23:51:39
配置文件: config.yaml

仓库统计:
- vllm-project/vllm: 32/89 问题 (过滤率: 36.0%)
- sgl-project/sglang: 28/76 问题 (过滤率: 36.8%)
- NVIDIA/TensorRT-LLM: 25/67 问题 (过滤率: 37.3%)
- microsoft/DeepSpeed: 22/58 问题 (过滤率: 37.9%)
- pytorch/pytorch: 19/45 问题 (过滤率: 42.2%)
- huggingface/transformers: 17/39 问题 (过滤率: 43.6%)
- ray-project/ray: 13/31 问题 (过滤率: 41.9%)

总计: 7 个仓库, 405 个问题, 156 个高价值问题
过滤率: 38.5%

工具: gh-pitfall-scraper
```