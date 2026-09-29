# EasyRenamer v0.6.0

Major desktop-shell modernization.

## Main window

- Removed the native Win32 menu bar, eliminating the white menu strip in dark mode.
- Replaced it with a dark in-app command bar.
- Language and theme selectors live directly in the app toolbar and still apply without restarting.
- Save/load method-set controls remain inside the working UI.
- Drag/drop preferences remain accessible without the old menu.

## Built-in help

- Added an internal Help window with:
  - Guide;
  - localized Tag reference;
  - About.
- The tag reference no longer opens GitHub.
- Help content is available directly from the EXE.

## Dialogs

- Replaced normal user-facing Win32 message boxes with themed EasyRenamer dialogs.
- Rename confirmation, Undo, preview errors, file/method-set errors and normal information messages now follow the selected theme.
- The native white Undo popup shown in earlier versions is removed from normal operation.

## Localization

- Removed the remaining hard-coded English labels from the main workflow.
- Localized method options, help text, confirmation dialogs, file filters and error titles.
- Russian, English, Spanish and Chinese are covered.
- Added an automated translation-completeness test so a new English UI key cannot silently be missing from another supported language.

## Existing stability work retained

- Cancellable preview scanning.
- Automatic cancellation of stale live-preview scans.
- Async Explorer actions.
- DPI/work-area-aware window sizing.
- Scrollable method/settings workspace.
- Neutral graphite dark palette.

## Quality

- Core tests must pass.
- Windows build must pass.
- Startup smoke test must pass.

## Downloads

- **EasyRenamer.exe** — portable Windows x64 executable.
- **EasyRenamer-v0.6.0-windows-x64.zip** — EXE + README + license + documentation.
- **SHA256SUMS.txt** — checksums.
