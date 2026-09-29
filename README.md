# EasyRenamer

EasyRenamer is an open-source batch file renamer for Windows. The goal is to keep the safety and power of tools like Advanced Renamer while staying free, transparent, portable, and easy to extend.

> Status: early MVP. The core rename engine is usable and tested; the Windows desktop UI is under active development.

## What it can do

- Preview every rename before touching files.
- Rename **all files** or filter by category: Images, Videos, Audio, Documents, Archives, or custom extensions.
- Include subfolders and number files independently inside each folder.
- Natural name ordering: `1`, `2`, `10` instead of `1`, `10`, `2`.
- Template-based names with counters, parent folders, Unix timestamps, random strings, dates, original names, and extensions.
- Text replacement with optional regular expressions.
- Prefix / suffix operations.
- Case conversion.
- Per-row checkboxes so you can exclude individual files from a batch.
- Conflict detection before execution.
- Two-phase transactional rename to avoid collisions.
- Undo the last rename operation.
- Single Windows executable in releases; no installer is required.

## Built-in categories

| Category | Examples |
| --- | --- |
| Images | JPG, PNG, WebP, TIFF, HEIC, AVIF, RAW |
| Videos | MP4, MKV, AVI, MOV, WebM, MTS |
| Audio | MP3, FLAC, WAV, M4A, OGG, Opus |
| Documents | PDF, DOCX, XLSX, PPTX, TXT, CSV, EPUB |
| Archives | ZIP, 7Z, RAR, TAR, GZ, ZST |
| Custom | Any extension list you provide |

## Template tokens

| Token | Meaning |
| --- | --- |
| `<Name>` | Original file name without extension |
| `<Ext>` | Extension without the dot |
| `<Inc:001>` | Global counter (`001`, `002`, ...) |
| `<Inc NrDir:01>` | Counter that restarts in each folder (`01`, `02`, ...) |
| `<DirName:1>` | Immediate parent folder name |
| `<UnixTimestamp>` | One Unix timestamp for the preview batch |
| `<Rand>` | One random digit |
| `<Rand Str:8>` | Random lowercase letters and digits |
| `<Rand Alpha:9>` | Random lowercase letters only |
| `<Date:yyyyMMdd-HHmmss>` | Batch date/time |

### Marketplace photo preset

This is the safe-sort preset that started the project:

```text
<Inc NrDir:01><Rand Alpha:9><UnixTimestamp>-<DirName:1>
```

Example inside folder `24`:

```text
1.jpg  -> 01abcdefghj1786009921-24.jpg
2.jpg  -> 02kqmdfzxwe1786009921-24.jpg
10.jpg -> 03rptncvbsa1786009921-24.jpg
```

The counter follows the current natural filename order inside each directory. `<Rand Alpha:9>` uses letters only so Windows Explorer does not merge the counter with a random digit and reorder `01...` as `016...`.

## Safety model

EasyRenamer treats preview as a plan, not a suggestion:

1. Scan and sort files.
2. Build target names.
3. Reject invalid Windows names and duplicate targets.
4. Move selected sources to unique temporary names.
5. Move temporary names to final names.
6. Roll back automatically if a batch fails midway.
7. Save the last successful operation for Undo.

## Build from source

Requirements:

- Go 1.23+
- Windows for the desktop UI build

```powershell
git clone https://github.com/Solvo37/easyrenamer.git
cd easyrenamer
go test ./internal/...
go build -trimpath -ldflags "-H windowsgui" -o EasyRenamer.exe ./cmd/easyrenamer
```

GitHub Actions builds the Windows executable automatically and creates release artifacts for version tags.

## Project structure

```text
cmd/easyrenamer/       Windows desktop UI
internal/engine/       scanning, categories, templates, preview, rename engine
internal/history/      undo history
internal/version/      build version
docs/                  token reference and roadmap
.github/workflows/     CI and release builds
```

## Roadmap

See [docs/ROADMAP.md](docs/ROADMAP.md). The next big areas are metadata renaming (EXIF/ID3), more composable rename methods, saved presets, folder renaming, CSV import/export, and richer history.

## Contributing

Issues and pull requests are welcome. See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

MIT. See [LICENSE](LICENSE).
