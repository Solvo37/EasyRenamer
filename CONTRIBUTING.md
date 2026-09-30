# Contributing to EasyRenamer

<p align="center"><strong>English</strong> · <a href="CONTRIBUTING.ru.md">Русский</a> · <a href="CONTRIBUTING.es.md">Español</a> · <a href="CONTRIBUTING.zh-CN.md">中文</a></p>

Bug reports, ideas and pull requests are welcome.

## Before you start

1. Check existing Issues and Pull Requests.
2. For a substantial feature, describe the real workflow first.
3. Do not add telemetry, advertising, accounts or mandatory network dependencies.

## Local development

Requirements: Go 1.23+, Node.js 20+, Wails v2.15, and Windows for the final desktop build.

```powershell
go test ./internal/...

cd frontend
npm install
npm run build

go install github.com/wailsapp/wails/v2/cmd/wails@v2.15.0
wails build -clean -platform windows/amd64 -o EasyRenamer.exe
```

## Pull request checklist

- code is formatted;
- `go test ./internal/...` passes;
- TypeScript build passes;
- new UI strings exist in EN / RU / ES / ZH;
- file operations still go through preview and validation;
- destructive overwrite is not enabled without safe rollback;
- user-facing workflow changes update the relevant README/LEARN documentation;
- generated binaries and build output are not committed.

## Architecture

```text
app.go                 Wails backend/API
frontend/              React + TypeScript UI
internal/engine/       rename engine
internal/history/      operation journal / Undo
internal/i18n/         translations
docs/                  user documentation
```

The UI must not duplicate critical rename logic. Processing order, validation and filesystem mutations belong in the Go engine.

## Change style

Prefer small, understandable commits. Avoid mixing a major feature, mass formatting and unrelated cleanup.

## Security

If a problem could overwrite, lose or corrupt user files, reproduce it with disposable test files and do not publish private data.

## License

By contributing, you agree that your changes are distributed under the [MIT License](LICENSE).
