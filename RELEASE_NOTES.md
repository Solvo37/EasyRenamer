# EasyRenamer v0.4.1

Hotfix release for the Windows startup crash in v0.4.0.

## Fixed

- Fixed `value out of range` during startup in the Walk `NumberEdit` controls.
- The app now shows a visible startup error dialog if window creation fails instead of silently exiting.
- CI and the release workflow now launch the freshly built Windows executable for a startup smoke test before accepting or publishing it.

## Background

v0.4.0 compiled and passed core tests, but several numeric editor defaults were assigned through Walk declarative properties during window initialization. On a real Windows launch this could conflict with the control's initialized range and terminate the app before the main window appeared.

v0.4.1 removes those declarative default-value assignments. Method defaults are still applied when the corresponding method is selected.

## Downloads

- **EasyRenamer.exe** — portable Windows x64 executable.
- **EasyRenamer-v0.4.1-windows-x64.zip** — EXE + README + license + documentation.
- **SHA256SUMS.txt** — checksums.
