# EasyRenamer v0.6.1

UX polish based on screenshots from v0.6.0.

## Method controls

- Reflowed selected-method actions into a 2×3 grid.
- Buttons now have enough room for Russian, Spanish, English, and Chinese labels.
- Save / Load use shorter localized labels so they do not get clipped in the left pane.

## Tags

- Replaced the tag-category tab strip with a category selector.
- Removed the native tag-table header that could render dark text on a dark surface.
- Added explicit readable Tag / Meaning headings using the app palette.
- The same browser is used in New Name and in the built-in Help window.

## Live preview

- Removed the separate **Live preview** checkbox.
- Live preview is now the default behavior and does not require a separate mode.
- Manual Preview / Cancel preview remains available for explicit rescans and heavy folders.

## Multiple folders

- Replaced the old single-folder tree dialog.
- **Folders** now opens the modern Windows IFileOpenDialog in folder mode with multi-select enabled.
- Ctrl / Shift selection can be used to add several folders in one operation.
- All selected folders are added to the current source list.

## Quality

- Translation completeness tests remain enabled.
- Core tests must pass.
- Windows build must pass.
- Startup smoke test must pass.

## Downloads

- **EasyRenamer.exe** — portable Windows x64 executable.
- **EasyRenamer-v0.6.1-windows-x64.zip** — EXE + README + license + documentation.
- **SHA256SUMS.txt** — checksums.
