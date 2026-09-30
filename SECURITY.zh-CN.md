# 安全

<p align="center"><a href="SECURITY.md">English</a> · <a href="SECURITY.ru.md">Русский</a> · <a href="SECURITY.es.md">Español</a> · <strong>中文</strong></p>

EasyRenamer 会修改文件系统，因此任何可能损坏用户数据的问题都被视为严重问题。

## 报告问题

普通 bug 可通过 GitHub Issues 提交。

如果问题可能意外覆盖、丢失、损坏或泄露用户文件，请使用可丢弃的测试数据复现，不要公开私人文件、凭据、个人路径或其他敏感信息。

## 安全原则

- 没有安全 backup/restore 时保持破坏性 overwrite 禁用；
- 执行前验证完整 rename 计划；
- 使用唯一临时名称避免内部冲突；
- 成功操作尽可能记录以支持 Undo。

修复不能仅为了“成功执行”而削弱这些保证。
