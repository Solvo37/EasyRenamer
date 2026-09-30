# 参与 EasyRenamer 开发

<p align="center"><a href="CONTRIBUTING.md">English</a> · <a href="CONTRIBUTING.ru.md">Русский</a> · <a href="CONTRIBUTING.es.md">Español</a> · <strong>中文</strong></p>

欢迎提交 bug、想法和 pull request。

## 开始之前

1. 先查看现有 Issues 和 Pull Requests。
2. 对较大的功能，先说明真实使用场景。
3. 不要加入遥测、广告、账号系统或强制网络依赖。

## 本地开发

要求：Go 1.23+、Node.js 20+、Wails v2.15；最终桌面 build 需要 Windows。

```powershell
go test ./internal/...
cd frontend
npm install
npm run build
go install github.com/wailsapp/wails/v2/cmd/wails@v2.15.0
wails build -clean -platform windows/amd64 -o EasyRenamer.exe
```

## PR 检查项

- 代码已格式化；
- Go tests 和 TypeScript build 通过；
- 新 UI 字符串在 EN/RU/ES/ZH 中都有翻译；
- 文件操作仍经过 preview/validation；
- 没有安全 rollback 时不能启用破坏性 overwrite；
- 用户 workflow 变化同步更新文档；
- 不提交生成的二进制和 build output。

## 架构

关键的 rename、排序、验证和文件系统操作应留在 Go engine 中，不在 UI 中重复实现。

## 安全

可能导致覆盖、丢失或损坏用户文件的问题应使用临时测试文件复现。

## 许可证

提交的更改按 [MIT](LICENSE) 发布。
