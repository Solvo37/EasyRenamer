# EasyRenamer v0.8.0

This release replaces the release UI layer instead of continuing to patch the old Win32/Walk forms.

## New desktop technology

- The production UI is now **Wails v2 + React + TypeScript**.
- The existing Go rename engine remains the source of truth for preview, validation, metadata, transactional rename and Undo.
- The old Walk frontend remains in the repository temporarily as a legacy fallback, but it is no longer the release shell.
- Wails v2 was chosen because it is the stable Wails line; v3 is still beta.

## Mockup-driven shell

- Frameless custom title bar with EasyRenamer branding and native window controls.
- Large primary command bar matching the approved mockup.
- Dedicated filter row.
- Card-based rename-method sidebar.
- Modern settings card below the selected method.
- Large file workspace with sticky table header.
- Selected-file thumbnail and metadata preview.
- Original → new filename comparison card.
- Full-width bottom status bar.
- Responsive layout for smaller Windows displays.

## Modern interaction

- CSS-based dark/light/system themes apply instantly.
- Language switching applies instantly using the existing Go translation tables.
- Native Wails drag & drop accepts files and folders.
- Folder drops use a custom in-app choice dialog.
- Native multi-file picker and multi-folder picker remain available.
- Ctrl+Z performs Undo.
- Save/load method sets remain supported.
- Local image thumbnails are shown only for the currently selected file.

## Rename methods

The React settings editor supports the existing engine methods:

- New Name / template tags
- Replace
- Renumber
- Change case
- Remove
- Remove pattern
- Add text
- Move
- List
- List Replace
- Swap
- Trim
- Timestamp
- Script expressions

The method chain on the left remains the single source of truth.

## Tags

- Built-in tag catalog.
- Category selector.
- Search by raw token or localized description.
- Tags insert directly at the caret in the New Name field.
- Functions and Script variables are available alongside tags.

## Build and stability

- GitHub Actions now builds the Wails production executable.
- React/TypeScript frontend is compiled as part of the release.
- Core Go tests pass.
- Windows production build passes.
- The produced EXE passes the startup smoke test.

## Downloads

- **EasyRenamer.exe** — portable Windows x64 executable.
- **EasyRenamer-v0.8.0-windows-x64.zip** — EXE + README + license + documentation.
- **SHA256SUMS.txt** — checksums.
