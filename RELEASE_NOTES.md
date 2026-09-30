# EasyRenamer v0.7.0

Major UI release driven directly by the supplied modern desktop mockup.

## New shell

- Added a product header with the ER badge, app version, subtitle and localized tagline.
- Reworked the dark palette into a blue/graphite hierarchy closer to the mockup.
- Enlarged the main command bar and made Add files / Start the primary actions.
- Filter and workspace sections now use distinct card-like surfaces.

## Rename methods

- Method rows are taller and easier to scan.
- Each method shows a localized short description plus a chevron.
- The left working pane is wider and behaves more like a modern sidebar.
- Selected-method settings remain immediately below the method stack.

## Files workspace

- The main file table now focuses on source name, new name, path, size, type and status.
- Source rows show the associated file icon.
- Ready/conflict states use localized green/red/muted status styling.
- The Files heading now shows the current item count.

## Selected-file preview

- Added a detail card below the file table.
- It shows the selected filename, path, dimensions when available, file size and type.
- The right side compares the original filename with the computed new filename.
- Selecting another row updates this card immediately.

## Status bar

- Added a full-width bottom status bar.
- Shows ready/error state, collision state and the current file/selection summary.

## Tags

- Added search inside the built-in tag browser.
- Search matches both raw tag syntax and localized descriptions.

## Existing behavior retained

- Always-on live preview with cancellation.
- Native multi-select folder picker.
- Live language/theme switching.
- Built-in Help and tag reference.
- Transactional rename and Undo.
- Four-language UI with translation completeness tests.

## Quality

- Core tests must pass.
- Windows build must pass.
- Startup smoke test must pass.

## Downloads

- **EasyRenamer.exe** — portable Windows x64 executable.
- **EasyRenamer-v0.7.0-windows-x64.zip** — EXE + README + license + documentation.
- **SHA256SUMS.txt** — checksums.
