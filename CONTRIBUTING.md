# Contributing to EasyRenamer

Спасибо за интерес к EasyRenamer. Баг-репорты, идеи и pull requests приветствуются.

## Перед началом

1. Проверьте существующие Issues и Pull Requests.
2. Для заметной новой функции лучше сначала описать сценарий использования в Issue.
3. Не добавляйте телеметрию, рекламу, аккаунты или обязательные сетевые зависимости.

## Локальная разработка

Требования:

- Go 1.23+;
- Node.js 20+;
- Wails v2.15;
- Windows для production desktop build.

Базовая проверка движка:

```powershell
go test ./internal/...
```

Frontend:

```powershell
cd frontend
npm install
npm run build
```

Production desktop build:

```powershell
go install github.com/wailsapp/wails/v2/cmd/wails@v2.15.0
wails build -clean -platform windows/amd64 -o EasyRenamer.exe
```

## Pull Request checklist

Перед PR:

- код форматирован;
- `go test ./internal/...` проходит;
- TypeScript build проходит;
- новые UI-строки добавлены во **все четыре** таблицы переводов;
- массовые файловые операции не обходят preview/validation;
- destructive overwrite не включается без безопасного rollback;
- README/LEARN обновлены, если меняется пользовательский workflow;
- бинарники и generated build output не коммитятся.

## Архитектура

```text
app.go                 Wails backend/API
frontend/              React + TypeScript UI
internal/engine/       rename engine
internal/history/      operation journal / Undo
internal/i18n/         translations
docs/                  user documentation
```

UI не должен дублировать критическую rename-логику: порядок обработки, validation и фактические файловые операции должны оставаться в Go-движке.

## Стиль изменений

Предпочтительны небольшие, понятные commits. Не смешивайте крупный feature, массовое форматирование и несвязанный cleanup без необходимости.

## Security

Если проблема может привести к потере или перезаписи пользовательских файлов, не публикуйте опасный proof-of-concept с реальными данными. Опишите сценарий достаточно подробно для воспроизведения на временных тестовых файлах.

## License

Отправляя изменения, вы соглашаетесь на их публикацию по лицензии [MIT](LICENSE).
