# EasyRenamer Roadmap

<p align="center"><strong>English</strong> · <a href="ROADMAP.ru.md">Русский</a> · <a href="ROADMAP.es.md">Español</a> · <a href="ROADMAP.zh-CN.md">中文</a></p>

This roadmap reflects the project after **v1.0.x**. Dates are intentionally not promised; stability and file safety come before calendar targets.

## ✅ v1.0 — shipped

The core desktop workflow is complete:

- Wails + React production UI;
- multiple folders in one task;
- file/folder drag & drop;
- folder grouping;
- natural sorting and explicit processing order;
- per-folder numbering;
- method chains;
- automatic preview;
- conflict and Windows filename validation;
- safe two-phase rename;
- progress + cancel;
- Undo and recent history;
- file search;
- tag autocomplete;
- context menus and shortcuts;
- configurable columns;
- user presets;
- table virtualization;
- EN / RU / ES / ZH UI.

## v1.1 — reliability and parity

High-value next work:

- true filesystem Timestamp method (Created / Modified / Accessed);
- method conditions;
- richer Apply To support for name / extension / both;
- stronger Renumber for existing numbers;
- richer Remove and Change Case modes;
- list import/export;
- CSV data import and CSV tags;
- folder rename mode;
- safe Copy / Move batch modes;
- file-pair workflows;
- operation-result summary;
- full window geometry persistence;
- accessibility and focus polish.

## v1.2 — power features

- metadata details panel and tag discovery;
- broader native metadata readers;
- optional ExifTool integration;
- richer Script runtime;
- CLI automation;
- export list to text / CSV / JSON / HTML;
- advanced collision rules;
- operation profiles and preset import/export;
- additional sort/filter rules.

## Later

- metadata writer;
- GPS reverse geocoding as an optional network feature;
- signed Windows binaries when infrastructure is available;
- macOS support if it can be maintained without hurting the Windows workflow.

## Non-goals

EasyRenamer is not intended to become a file manager, cloud service, account-based product, subscription product, advertising platform, or telemetry-heavy application.

Core workflow:

```text
Add files / folders
→ configure methods
→ verify output
→ run
→ undo if needed
```
