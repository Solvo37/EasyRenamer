# EasyRenamer v0.4.2

UX/UI and live-preview release.

## Changed

- Added **System / Light / Dark** theme modes.
- System mode follows the Windows app theme at startup.
- Added dark title-bar and native-control theming on supported Windows versions.
- Refreshed spacing, row heights, window sizing, and control sizing for a less cramped interface.
- Renamed the auto-preview option to **Live preview** for clearer behavior.

## Fixed

- **New Name now updates while typing.**
- Text-based method settings no longer wait for focus loss before recalculating preview.
- Preview recalculation is debounced to avoid rebuilding the file list on every keystroke.
- If another edit happens while preview generation is already running, the latest preview request is queued and recalculated afterwards.
- The preview table is explicitly invalidated after model refresh to prevent stale **New filename** cells that only repainted after mouse movement.

## Quality

- Core tests pass.
- Windows build passes.
- The built executable passes the Windows startup smoke test before release publication.

## Downloads

- **EasyRenamer.exe** — portable Windows x64 executable.
- **EasyRenamer-v0.4.2-windows-x64.zip** — EXE + README + license + documentation.
- **SHA256SUMS.txt** — checksums.
