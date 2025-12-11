---
AIGC:
    ContentProducer: Minimax Agent AI
    ContentPropagator: Minimax Agent AI
    Label: AIGC
    ProduceID: "00000000000000000000000000000000"
    PropagateID: "00000000000000000000000000000000"
    ReservedCode1: 304502206c4f57c5e1a017444d978f27e9137e8f4af5b89d5ff0047a405117c7c4816ce8022100cbf23e38f571e7e432b74a132ec9842a8923d2e067ef6487d4fe66e70a36f8db
    ReservedCode2: 3045022100c301daf9b77decb1524f64050afedb9b1f1acf9a2b1f38dbd5cfffabddb2a8650220210ad06327c1507615397b48a4e8a004987fbbd14483496515dfac7322731e7e
---

# gh-pitfall-scraper 文档索引

## 📚 文档分类

### 🚀 快速开始
| 文档 | 描述 | 适合人群 |
|------|------|----------|
| [README.md](README.md) | 项目总体介绍和快速开始 | 所有用户 |
| [config-database-example.yaml](config-database-example.yaml) | 完整配置示例文件 | 配置管理员 |

### 🗄️ 数据库文档
| 文档 | 描述 | 适合人群 |
|------|------|----------|
| [DATABASE_USAGE_GUIDE.md](DATABASE_USAGE_GUIDE.md) | 数据库完整使用指南 | 数据库管理员 |
| [DATABASE_DESIGN_SUMMARY.md](DATABASE_DESIGN_SUMMARY.md) | 数据库架构设计说明 | 架构师 |
| [DATABASE_INTEGRATION_REPORT.md](DATABASE_INTEGRATION_REPORT.md) | 数据库集成实现报告 | 开发者 |

### 🔧 运维指南
| 文档 | 描述 | 适合人群 |
|------|------|----------|
| [DATABASE_BEST_PRACTICES.md](DATABASE_BEST_PRACTICES.md) | 数据库最佳实践和演示 | 运维工程师 |
| [DATABASE_TROUBLESHOOTING.md](DATABASE_TROUBLESHOOTING.md) | 故障排除和常见问题 | 运维工程师 |

## 🎯 按需求查找文档

### 我是新手，想要...
- **快速体验项目** → 阅读 [README.md](README.md#快速开始)
- **配置第一个数据库** → 查看 [config-database-example.yaml](config-database-example.yaml)
- **了解数据库功能** → 阅读 [DATABASE_USAGE_GUIDE.md](DATABASE_USAGE_GUIDE.md#快速开始)

### 我需要管理数据库...
- **日常操作指南** → [DATABASE_USAGE_GUIDE.md](DATABASE_USAGE_GUIDE.md#数据库操作)
- **备份和恢复** → [DATABASE_USAGE_GUIDE.md](DATABASE_USAGE_GUIDE.md#备份恢复)
- **性能优化** → [DATABASE_USAGE_GUIDE.md](DATABASE_USAGE_GUIDE.md#性能优化)
- **监控和维护** → [DATABASE_USAGE_GUIDE.md](DATABASE_USAGE_GUIDE.md#监控维护)

### 我遇到了问题...
- **数据库连接问题** → [DATABASE_TROUBLESHOOTING.md](DATABASE_TROUBLESHOOTING.md#数据库连接问题)
- **性能问题** → [DATABASE_TROUBLESHOOTING.md](DATABASE_TROUBLESHOOTING.md#性能问题排查)
- **数据损坏** → [DATABASE_TROUBLESHOOTING.md](DATABASE_TROUBLESHOOTING.md#数据问题处理)
- **配置错误** → [DATABASE_TROUBLESHOOTING.md](DATABASE_TROUBLESHOOTING.md#配置问题诊断)

### 我想要优化性能...
- **配置调优** → [DATABASE_BEST_PRACTICES.md](DATABASE_BEST_PRACTICES.md#配置最佳实践)
- **查询优化** → [DATABASE_USAGE_GUIDE.md](DATABASE_USAGE_GUIDE.md#性能优化)
- **索引优化** → [DATABASE_USAGE_GUIDE.md](DATABASE_USAGE_GUIDE.md#索引优化)

### 我需要部署到生产环境...
- **部署准备** → [DATABASE_BEST_PRACTICES.md](DATABASE_BEST_PRACTICES.md#生产环境部署)
- **安全配置** → [DATABASE_BEST_PRACTICES.md](DATABASE_BEST_PRACTICES.md#安全配置)
- **监控告警** → [DATABASE_BEST_PRACTICES.md](DATABASE_BEST_PRACTICES.md#监控告警)

### 我想了解技术细节...
- **数据库架构** → [DATABASE_DESIGN_SUMMARY.md](DATABASE_DESIGN_SUMMARY.md)
- **集成实现** → [DATABASE_INTEGRATION_REPORT.md](DATABASE_INTEGRATION_REPORT.md)
- **设计模式** → [DATABASE_DESIGN_SUMMARY.md](DATABASE_DESIGN_SUMMARY.md#核心表设计)

## 🔍 快速查找

### 按主题搜索
- **配置**: `config-database-example.yaml`
- **初始化**: `DATABASE_USAGE_GUIDE.md#快速开始`
- **备份**: `DATABASE_USAGE_GUIDE.md#备份恢复`
- **迁移**: `DATABASE_USAGE_GUIDE.md#数据库迁移`
- **监控**: `DATABASE_USAGE_GUIDE.md#监控维护`
- **故障**: `DATABASE_TROUBLESHOOTING.md`
- **性能**: `DATABASE_BEST_PRACTICES.md#性能调优`
- **安全**: `DATABASE_BEST_PRACTICES.md#安全实践`

### 按文件类型搜索
- **配置**: `*.yaml`
- **文档**: `*.md`
- **脚本**: `*.sh`

### 按功能模块搜索
- **数据库操作**: `DATABASE_USAGE_GUIDE.md#数据库操作`
- **管理工具**: `DATABASE_USAGE_GUIDE.md#数据库管理工具`
- **迁移工具**: `DATABASE_USAGE_GUIDE.md#数据库迁移管理`
- **监控工具**: `DATABASE_USAGE_GUIDE.md#监控维护`

## 📋 常用操作快速参考

### 日常操作
```bash
# 查看状态
./gh-pitfall-scraper --health

# 查看统计
./gh-pitfall-scraper --stats

# 创建备份
./database/db-manager.sh backup

# 清理数据
./database/db-manager.sh cleanup
```

### 故障处理
```bash
# 快速诊断
./quick-diagnosis.sh

# 详细诊断
./detailed-diagnosis.sh

# 紧急恢复
./emergency-response.sh
```

### 维护任务
```bash
# 每日维护
./auto-maintenance.sh now

# 设置自动维护
./auto-maintenance.sh setup

# 健康检查
./health-check-scheduler.sh once
```

## 📞 获取帮助

### 社区支持
- **GitHub Issues**: 报告 Bug 和功能请求
- **GitHub Discussions**: 社区讨论和问答
- **Wiki**: 社区维护的补充文档

### 文档反馈
如果您发现文档问题或有改进建议，请：
1. 提交 Issue 描述问题
2. 发起 Pull Request 贡献内容
3. 参与文档讨论

### 技术支持
- 查看 [常见问题](DATABASE_TROUBLESHOOTING.md#常见问题及解决方案)
- 运行 [诊断脚本](DATABASE_TROUBLESHOOTING.md#快速诊断)
- 检查 [错误代码](DATABASE_TROUBLESHOOTING.md#错误代码说明)

---

*本文档索引帮助您快速找到需要的文档。如果无法找到相关信息，请查看相关主题文档或提交 Issue。*