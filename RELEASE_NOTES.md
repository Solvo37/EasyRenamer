# EasyRenamer v0.4.4

Power-user UX release moving the interface closer to the workflow of established desktop batch renamers.

## New Name / tags

- Replaced the raw tag ComboBox with a tabbed tag browser.
- Categories: Default, Advanced, Time, Image, Video, Audio, Documents, GPS, File.
- Every entry shows the raw tag plus a localized human-readable meaning.
- Double-click inserts the selected tag at the caret.
- English, Russian, Spanish, and Chinese tag labels are included.
- Direct link to the full tag documentation remains available.

## Drag and drop

- Dropping items now opens a clear choice dialog.
- Choose files and folders, only files, or only folders.
- Choose whether folder scanning includes subfolders.
- Optionally remember the selected behavior.
- The saved choice can be changed again from the File menu.

## Dark mode

- Native ListView backgrounds/text are explicitly themed.
- ComboBox popup lists are re-themed when opened.
- This removes the most obvious white list areas inside dark mode while keeping the softer gray palette from v0.4.3.

## Quality

- Core tests must pass.
- Windows build must pass.
- The produced EXE must pass the Windows startup smoke test before publication.

## Downloads

- **EasyRenamer.exe** — portable Windows x64 executable.
- **EasyRenamer-v0.4.4-windows-x64.zip** — EXE + README + license + documentation.
- **SHA256SUMS.txt** — checksums.
