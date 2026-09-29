# EasyRenamer v0.5.1

Stability and desktop-layout hotfix after v0.5.0.

## Responsiveness

- Preview scans are now cancellable.
- The Preview button becomes **Cancel preview** while a scan is running.
- Live preview cancels stale work instead of allowing expensive scans to pile up.
- Explorer/open-source actions no longer perform file-system checks on the UI thread.
- Opening a source or selected file is dispatched asynchronously so Explorer cannot freeze the EasyRenamer window.

## Window sizing

- The initial window size is calculated from the Windows work area.
- DPI scaling is taken into account.
- The minimum size is reduced.
- The left method/settings workspace is vertically scrollable on smaller screens.
- This avoids first-launch windows extending under the taskbar or placing the close button off-screen.

## Toolbar

- Removed the redundant Batch mode dropdown. EasyRenamer currently has one real batch operation: Rename.

## Dark theme

- Rebuilt the dark palette around neutral charcoal surfaces:
  - window: near #202123;
  - panels: softer raised graphite;
  - fields: slightly lighter than panels;
  - high-contrast white blocks are avoided where Walk/Win32 allows it.
- Existing native dark ListView/ComboBox theming remains enabled.

## Engine

- Added context-aware preview scanning with cancellation checks during recursive directory walking and item processing.
- Existing `Preview` API remains compatible.

## Quality

- Core tests pass.
- Windows build passes.
- Startup smoke test passes.

## Downloads

- **EasyRenamer.exe** — portable Windows x64 executable.
- **EasyRenamer-v0.5.1-windows-x64.zip** — EXE + README + license + documentation.
- **SHA256SUMS.txt** — checksums.
