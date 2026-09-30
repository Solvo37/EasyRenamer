<p align="center">
  <img src="frontend/public/easyrenamer.svg" width="96" height="96" alt="EasyRenamer">
</p>

<h1 align="center">EasyRenamer</h1>

<p align="center">
  Быстрый и безопасный пакетный переименователь файлов для Windows.<br>
  <strong>Бесплатно · Open Source · Без рекламы · Без подписки · Без телеметрии</strong>
</p>

<p align="center">
  <a href="https://github.com/Solvo37/easyrenamer/actions/workflows/ci.yml"><img alt="CI" src="https://github.com/Solvo37/easyrenamer/actions/workflows/ci.yml/badge.svg"></a>
  <a href="https://github.com/Solvo37/easyrenamer/releases/latest"><img alt="Release" src="https://img.shields.io/github/v/release/Solvo37/easyrenamer?display_name=tag"></a>
  <a href="LICENSE"><img alt="License" src="https://img.shields.io/github/license/Solvo37/easyrenamer"></a>
  <img alt="Windows x64" src="https://img.shields.io/badge/Windows-x64-0078D4?logo=windows11&logoColor=white">
</p>

<p align="center">
  <a href="https://github.com/Solvo37/easyrenamer/releases/latest/download/EasyRenamer.exe"><strong>Скачать EasyRenamer.exe</strong></a>
  ·
  <a href="docs/LEARN.md">Руководство</a>
  ·
  <a href="docs/TAGS.md">Справочник тегов</a>
  ·
  <a href="docs/ROADMAP.md">Roadmap</a>
</p>

---

## Зачем EasyRenamer

EasyRenamer рассчитан на реальную пакетную работу: от нескольких файлов до больших наборов из разных папок. Итоговые имена видны **до запуска**, порядок обработки задаётся явно, конфликты проверяются заранее, а сама операция выполняется транзакционно и может быть отменена.

Основной интерфейс остаётся компактным и одностраничным:

**добавил файлы → настроил методы → проверил результат → запустил**

## Главное

| Возможность | Что даёт |
| --- | --- |
| **Live preview** | Новые имена пересчитываются автоматически при изменении шаблона, методов, сортировки и фильтров |
| **Группировка по папкам** | Видно, где начинается новая группа и почему `<Inc NrDir:01>` сбрасывает счётчик |
| **Natural sorting** | `1, 2, 10` вместо `1, 10, 2` |
| **Порядок обработки** | Имя, дата создания/изменения, размер, расширение, путь, порядок добавления и ручной drag & drop |
| **Цепочка методов** | Методы выполняются сверху вниз; их можно включать, отключать, дублировать и переставлять |
| **Поиск и фильтр ошибок** | Поиск по исходному имени, новому имени и полному пути |
| **Collision policies** | Пропуск конфликтов, автоматическая нумерация или остановка операции |
| **Progress + Cancel** | Большая операция не блокирует интерфейс и может быть безопасно остановлена |
| **Undo + History** | Откат последней операции и журнал последних 20 переименований |
| **Большие списки** | Виртуализированная таблица и ленивое получение тяжёлых preview-данных |
| **4 языка** | English, Русский, Español, 中文 |
| **Portable** | Один EXE, без отдельного установщика |

## Пример

Шаблон:

```text
<Inc NrDir:01>_<Name>
```

Файлы в каждой папке получат собственную последовательность:

```text
1.jpg   → 01_1.jpg
2.jpg   → 02_2.jpg
10.jpg  → 03_10.jpg
```

При добавлении следующей папки счётчик `NrDir` снова начинается с `01`.

## Методы переименования

EasyRenamer поддерживает цепочки из нескольких методов:

- **New Name** — шаблон из текста и тегов;
- **Replace** — поиск и замена, включая regex;
- **Renumber** — нумерация, в том числе отдельно по папкам;
- **Add text** — префикс и суффикс;
- **Change case** — lower / UPPER / Title Case;
- **Remove / Remove pattern** — удаление по позиции, тексту или regex;
- **Move / Swap** — перестановка частей имени;
- **Trim** — очистка лишних пробелов;
- **Timestamp** — дата/время файла или текущего batch;
- **List / List Replace** — имена и правила построчно;
- **Script** — безопасные выражения без доступа к сети или файловой системе.

Полный разбор: [docs/LEARN.md](docs/LEARN.md).

## Теги и метаданные

В шаблонах доступны счётчики, папки, даты, размеры, случайные значения, checksums и распространённые метаданные:

```text
<Name>
<Ext>
<FolderName:1>
<Inc Nr:001>
<Inc NrDir:01>
<Date Modified:yyyy-mm-dd>
<Artist>
<Album>
<Width>
<Height>
<MD5>
```

Теги поддерживают fallback и modifiers. Полный список: [docs/TAGS.md](docs/TAGS.md).

В поле **New Name** есть autocomplete: начните ввод с `<`, затем выбирайте тег стрелками и Enter.

## Безопасность

EasyRenamer не выполняет массовый rename наивно по одному файлу.

Перед запуском проверяются:

- пустые и недопустимые имена;
- запрещённые Windows-символы;
- системные имена вроде `CON`, `NUL`, `COM1`, `LPT1`;
- одинаковые итоговые пути;
- уже существующие destination-файлы;
- слишком длинные пути;
- исчезнувшие файлы.

Переименование выполняется в две фазы через уникальные временные имена `.easyrenamer_tmp_*`. Поэтому корректно обрабатываются сценарии вроде:

```text
A.jpg → B.jpg
B.jpg → A.jpg
```

и изменение только регистра:

```text
photo.jpg → PHOTO.jpg
```

**Destructive overwrite намеренно отключён**, пока для него не реализован безопасный backup/restore.

## Горячие клавиши

| Shortcut | Действие |
| --- | --- |
| `Ctrl + O` | Добавить файлы |
| `Ctrl + Shift + O` | Добавить папки |
| `Ctrl + A` | Выделить строки |
| `Ctrl + Z` | Откатить последнюю доступную операцию |
| `Delete` | Убрать выбранное из задачи |
| `Ctrl + Enter` | Запустить |
| `F5` | Пересчитать / проверить |
| `Esc` | Отменить текущую операцию |

## Интерфейс и настройки

EasyRenamer запоминает рабочие настройки интерфейса:

- тему;
- сортировку и направление;
- группировку по папкам;
- collision policy;
- положение splitter'ов;
- ширину, порядок и видимость колонок;
- пользовательские шаблонные пресеты.

Список файлов предыдущей сессии автоматически не восстанавливается.

## Языки

Интерфейс полностью ведётся через общий i18n-слой:

- English
- Русский
- Español
- 中文

Полнота таблиц переводов проверяется тестом. Переключение языка применяется без перезапуска приложения.

## Технологии

- **Go 1.23+** — движок, файловые операции, метаданные и backend;
- **Wails v2** — Windows desktop shell;
- **React + TypeScript** — интерфейс;
- **Lucide** — единый набор UI-иконок;
- **MIT License**.

Никаких аккаунтов, облака, аналитики или обязательного сетевого сервиса.

## Сборка

Требования:

- Go 1.23+;
- Node.js 20+;
- Wails v2.15;
- Windows x64 для финальной desktop-сборки.

```powershell
git clone https://github.com/Solvo37/easyrenamer.git
cd easyrenamer

go test ./internal/...
go install github.com/wailsapp/wails/v2/cmd/wails@v2.15.0
wails build -clean -platform windows/amd64 -o EasyRenamer.exe
```

CI выполняет core tests, production Windows build и smoke-test запуска EXE.

## Структура репозитория

```text
app.go                 Wails backend/API
main.go                desktop entry point
folder_picker_windows.go
frontend/              React + TypeScript UI
internal/engine/       scan, sorting, tags, preview, validation, rename
internal/history/      operation journal and Undo
internal/i18n/         EN / RU / ES / ZH translations
internal/version/      build version
docs/                  guide, tags and roadmap
.github/workflows/     CI and release automation
```

## Скачать

Актуальные сборки находятся в [GitHub Releases](https://github.com/Solvo37/easyrenamer/releases/latest):

- `EasyRenamer.exe` — portable Windows x64;
- `EasyRenamer-vX.Y.Z-windows-x64.zip` — EXE + документация;
- `SHA256SUMS.txt` — контрольные суммы.

## Участие в разработке

Баги, идеи и pull requests приветствуются. Перед изменениями посмотрите [CONTRIBUTING.md](CONTRIBUTING.md).

## Лицензия

[MIT](LICENSE).
