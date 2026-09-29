# Roadmap

EasyRenamer is intended to grow from a focused batch renamer into a serious open-source alternative for Windows power users.

## v0.1 — Foundation

- [x] Native Windows GUI
- [x] Multiple files and folders as sources
- [x] File categories and custom extensions
- [x] Recursive scan
- [x] Natural sorting
- [x] Template engine
- [x] Per-folder and global counters
- [x] Random/date/parent-folder tokens
- [x] Replace + regex
- [x] Prefix / suffix
- [x] Case conversion
- [x] Preview validation
- [x] Select individual files
- [x] Transactional rename
- [x] Undo last operation
- [x] CI and single-exe release workflow

## v0.2 — Rename methods and desktop workflow

- [x] Chained rename methods (method stack)
- [x] Move / reorder methods in the stack
- [x] Enable/disable individual methods without deleting them
- [x] Drag reorder methods in the stack
- [x] List method with load/save/populate
- [x] List Replace with multiple pairs
- [x] Swap by separator and occurrence
- [x] Safe Script expression method
- [x] Remove characters by position
- [x] Move filename fragments by position
- [x] Remove text / regex patterns
- [x] Timestamp method using file modified time or batch time
- [x] Trim / normalize whitespace
- [ ] Extension-specific transformations
- [ ] Search/replace presets
- [x] Dedicated numbering method: start, step, padding, per-folder/global
- [ ] Optional transliteration
- [ ] Filename sanitization profiles

## v0.3 — Metadata

- [x] Common EXIF date / dimensions / GPS tokens
- [x] Common MP3 ID3 / FLAC artist / album / track tokens
- [x] Common MP4 / MOV video metadata tokens
- [x] File creation/modification date tokens
- [x] Checksum tags (MD5 / SHA1)

## v0.4 — Workflow

- [x] Save and load method stacks / presets
- [ ] Rename history browser (not only last operation)
- [ ] CSV import/export
- [x] Drag-and-drop files and folders
- [ ] Folder renaming
- [ ] Conflict-resolution helpers
- [ ] Dry-run export as JSON/CSV

## v0.5 — Polish

- [x] Rebuilt workspace: methods + selected settings on the left, file list on the right
- [x] Live language switching without restarting the process
- [x] Live System / Light / Dark switching without restarting the process
- [x] Preserve current sources, methods, filter and selected method across UI rebuild

- [x] System / light / dark theme
- [x] Localization (English / Russian / Spanish / Chinese)
- [ ] Search/filter inside preview
- [x] File size and image dimensions in preview table
- [x] Debounced live preview while editing method settings
- [x] Explicit preview repaint after model refresh
- [x] Localized built-in preset names
- [x] Visible drag-and-drop affordance
- [x] Softer dark palette and dark tab/table surfaces
- [x] Localized tabbed tag browser in New Name
- [x] Drag/drop choice dialog for files, folders and subfolders
- [x] Rememberable drag/drop behavior
- [x] Native dark ListView and ComboBox popup theming
- [ ] Keyboard shortcuts
- [ ] Context-menu integration (optional)
- [ ] Signed Windows releases when infrastructure is available
- [ ] Portable settings mode

## v0.4 — Public release

- [x] Portable GitHub Release workflow
- [x] SHA-256 release checksums
- [x] Built-in tag reference
- [x] Built-in Learn / user guide
- [x] Common document / email / EPUB metadata tags
- [x] Windows EXE version-resource tags
- [ ] Arbitrary ExifTool fields without an external dependency
- [ ] CSV column tag importer
- [ ] Offline GPS country/city/state reverse-geocoding database

## v0.4.5 — UX clarity

- [x] Settings-only right pane: selected left-side method is the single source of truth
- [x] Right editor can no longer change method type
- [x] Always-enabled extension filter
- [x] Typing extensions selects the Custom category
- [x] ComboBox dropdown no longer closes because the owner window is re-themed while opening
