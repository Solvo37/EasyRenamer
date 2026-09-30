<p align="center"><img src="frontend/public/easyrenamer.svg" width="96" height="96" alt="EasyRenamer"></p>
<h1 align="center">EasyRenamer</h1>

<p align="center"><strong>English</strong> · <a href="README.ru.md">Русский</a> · <a href="README.es.md">Español</a> · <a href="README.zh-CN.md">中文</a></p>

<p align="center">Fast and safe batch file renaming for Windows.<br><strong>Free · Open Source · No ads · No subscription · No telemetry</strong></p>

<p align="center">
  <a href="https://github.com/Solvo37/easyrenamer/actions/workflows/ci.yml"><img alt="CI" src="https://github.com/Solvo37/easyrenamer/actions/workflows/ci.yml/badge.svg"></a>
  <a href="https://github.com/Solvo37/easyrenamer/releases/latest"><img alt="Release" src="https://img.shields.io/github/v/release/Solvo37/easyrenamer?display_name=tag"></a>
  <a href="LICENSE"><img alt="License" src="https://img.shields.io/github/license/Solvo37/easyrenamer"></a>
  <img alt="Windows x64" src="https://img.shields.io/badge/Windows-x64-0078D4?logo=windows11&logoColor=white">
</p>

<p align="center">
  <a href="https://github.com/Solvo37/easyrenamer/releases/latest/download/EasyRenamer.exe"><strong>Download EasyRenamer.exe</strong></a>
  · <a href="docs/LEARN.md">User guide</a>
  · <a href="docs/TAGS.md">Tag reference</a>
  · <a href="docs/ROADMAP.md">Roadmap</a>
</p>

---

## Why EasyRenamer

EasyRenamer is built for real batch-renaming work, from a few files to large sets spread across multiple folders. Final names are visible <strong>before anything is changed</strong>, processing order is explicit, conflicts are checked in advance, and rename operations are transactional and reversible.

<strong>Add files → configure methods → review the result → run</strong>

## Highlights

| Feature | What it gives you |
| --- | --- |
| Live preview | New names update automatically when templates, methods, sorting or filters change |
| Folder grouping | Clear folder boundaries and predictable per-folder numbering |
| Natural sorting | 1, 2, 10 instead of 1, 10, 2 |
| Processing order | Name, creation/modification date, size, extension, path, added order or manual drag & drop |
| Method chain | Enable, disable, duplicate and reorder rename methods |
| Search & error filter | Search original names, new names and full paths |
| Collision policies | Skip conflicts, auto-number them, or stop the operation |
| Progress + Cancel | Large operations stay responsive and can be stopped safely |
| Undo + History | Undo and a journal of the latest 20 rename operations |
| Large lists | Virtualized table and lazy heavy preview data |
| 4 UI languages | English, Русский, Español, 中文 |
| Portable | One EXE, no installer required |

## Example

Template: <code>&lt;Inc NrDir:01&gt;_&lt;Name&gt;</code>

<pre>
1.jpg   → 01_1.jpg
2.jpg   → 02_2.jpg
10.jpg  → 03_10.jpg
</pre>

The <code>NrDir</code> counter starts again from <code>01</code> for each folder.

## Rename methods

EasyRenamer supports chained methods:

- <strong>New Name</strong> — names from text and tags;
- <strong>Replace</strong> — find and replace, including regex;
- <strong>Renumber</strong> — sequential numbering, including per-folder numbering;
- <strong>Add text</strong> — prefix and suffix;
- <strong>Change case</strong> — lower / UPPER / Title Case;
- <strong>Remove / Remove pattern</strong> — remove by position, text or regex;
- <strong>Move / Swap</strong> — rearrange filename parts;
- <strong>Trim</strong> — clean up whitespace;
- <strong>Timestamp</strong> — add file or batch date/time;
- <strong>List / List Replace</strong> — names and rules line by line;
- <strong>Script</strong> — safe expressions without direct filesystem or network access.

Detailed guide: [docs/LEARN.md](docs/LEARN.md).

## Tags and metadata

Templates support counters, folders, dates, sizes, random values, checksums and common metadata, for example:

<pre>
&lt;Name&gt;
&lt;Ext&gt;
&lt;FolderName:1&gt;
&lt;Inc Nr:001&gt;
&lt;Inc NrDir:01&gt;
&lt;Date Modified:yyyy-mm-dd&gt;
&lt;Artist&gt;
&lt;Album&gt;
&lt;Width&gt;
&lt;Height&gt;
&lt;MD5&gt;
</pre>

Tags support fallback chains and modifiers. Full reference: [docs/TAGS.md](docs/TAGS.md).

## Safety

Before execution EasyRenamer checks invalid Windows names, forbidden characters, reserved names such as CON/NUL/COM1/LPT1, duplicate destination paths, existing targets, overly long paths and missing source files.

Renaming runs in two phases through unique temporary names such as <code>.easyrenamer_tmp_*</code>, so swaps and case-only renames are safe.

<strong>Destructive overwrite is intentionally disabled</strong> until safe backup/restore is implemented.

## Shortcuts

| Shortcut | Action |
| --- | --- |
| Ctrl + O | Add files |
| Ctrl + Shift + O | Add folders |
| Ctrl + A | Select rows |
| Ctrl + Z | Undo latest available operation |
| Delete | Remove selected items from the task |
| Ctrl + Enter | Start |
| F5 | Recalculate / check |
| Esc | Cancel current operation |

## Languages

The application UI supports <strong>English, Русский, Español and 中文</strong>. Translation completeness is covered by tests and language switching does not require a restart.

## Technology

Go 1.23+ · Wails v2 · React · TypeScript · Lucide · MIT License.

No accounts, cloud backend, analytics or mandatory network service.

## Build from source

Requirements: Go 1.23+, Node.js 20+, Wails v2.15, Windows x64 for the final desktop build.

<pre>
git clone https://github.com/Solvo37/easyrenamer.git
cd easyrenamer
go test ./internal/...
go install github.com/wailsapp/wails/v2/cmd/wails@v2.15.0
wails build -clean -platform windows/amd64 -o EasyRenamer.exe
</pre>

## Downloads

Current builds: [GitHub Releases](https://github.com/Solvo37/easyrenamer/releases/latest)

- EasyRenamer.exe — portable Windows x64;
- EasyRenamer-vX.Y.Z-windows-x64.zip — EXE + documentation;
- SHA256SUMS.txt — checksums.

## Contributing

Bug reports, ideas and pull requests are welcome. See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

[MIT](LICENSE).
