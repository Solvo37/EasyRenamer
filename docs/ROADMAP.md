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

## v0.5.1 — Stability

- [x] Cancellable preview scans
- [x] Cancel stale live-preview work instead of queueing heavy scans
- [x] Preview button switches to Cancel preview while scanning
- [x] Explorer/open-source actions moved off the UI thread
- [x] Window size derived from Windows work area and DPI
- [x] Scrollable left method/settings pane for smaller screens
- [x] Remove the redundant single-option Batch mode selector
- [x] Rebuilt neutral dark palette

## v0.5.2 — Flat workspace

- [x] Replace framed left-side GroupBoxes with flat sections
- [x] Bold section headings for Methods / Settings / Add method / Files
- [x] Compact localized Add method selector
- [x] Remove 14-button Add method grid
- [x] Remove ambiguous Open source button from filter row

## v0.6.0 — Modern shell

- [x] Remove native Win32 menu strip from the main window
- [x] Dark in-app command bar with live language/theme selectors
- [x] Built-in Guide / Tags / About help
- [x] No external GitHub link required for tag documentation
- [x] Themed in-app info/error/confirmation dialogs
- [x] Localize remaining hard-coded UI text in all four languages
- [x] Translation completeness test for Russian / Spanish / Chinese
- [x] Save/load method-set controls inside the main workspace
- [x] Keep drag/drop settings accessible without the native menu

## v0.6.1 — UX polish

- [x] Native multi-select folder picker with Ctrl/Shift selection
- [x] Always-on live preview; remove confusing Live preview checkbox
- [x] Replace tag-category tab strip with category selector
- [x] Hide dark-unreadable native tag table header
- [x] Custom readable tag headings in dark mode
- [x] Reflow method action buttons into a 2×3 grid
- [x] Short localized Save / Load labels so controls do not clip

## v0.7.0 — Mockup-driven shell

- [x] Product header with ER badge, version, subtitle and tagline
- [x] Blue/graphite mockup-style dark palette
- [x] Larger command bar with primary Add files and Start actions
- [x] Card-style filter bar and working panels
- [x] Method rows with localized descriptions and chevrons
- [x] Wider/taller method stack matching the supplied mockup
- [x] Simplified file table: source, new name, path, size, type, status
- [x] File icons in source-name rows
- [x] Localized status coloring
- [x] Selected-file detail card with metadata and original → new comparison
- [x] Full-width bottom status bar
- [x] File count in the Files heading
- [x] Search inside the built-in tag browser

## v0.8.0 — New UI layer

- [x] Replace release UI technology: Walk → Wails v2 + React + TypeScript
- [x] Keep the existing Go rename engine instead of rewriting core logic
- [x] Frameless custom Windows title bar
- [x] Mockup-driven command bar / filter bar / sidebar / file workspace
- [x] CSS-based dark/light/system themes
- [x] Live language switching using existing Go translation tables
- [x] React method stack with all existing rename method editors
- [x] Modern built-in tag browser with search
- [x] Native file drop integration through Wails
- [x] Multi-file and multi-folder pickers
- [x] Selected-file thumbnail/metadata preview
- [x] Original → new filename comparison
- [x] Custom modal confirmations / drop workflow / help
- [x] Ctrl+Z Undo
- [x] Production Wails build + Windows startup smoke test
- [ ] Remove legacy Walk frontend after the new shell has been field-tested
