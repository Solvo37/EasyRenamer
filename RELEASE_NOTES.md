# EasyRenamer v0.4.3

UI polish release based on feedback from v0.4.2.

## Changed

- Dark mode now uses a softer gray palette instead of near-black surfaces.
- Tab pages and preview tables now use the dark surface colors instead of leaving large white areas.
- Edit and combo controls receive more appropriate native Windows dark themes.
- Built-in preset names are localized in English, Russian, Spanish, and Chinese.
- Added a visible drag-and-drop hint above the preview when the source list is empty.

## Drag and drop

Drag and drop was already supported by the window, but v0.4.3 makes it discoverable:

- drag files directly from Explorer;
- drag folders directly from Explorer;
- the hint disappears after sources are added and returns when the list is cleared.

## Quality

- Core tests must pass.
- Windows build must pass.
- The produced EXE must pass the Windows startup smoke test before publishing.

## Downloads

- **EasyRenamer.exe** — portable Windows x64 executable.
- **EasyRenamer-v0.4.3-windows-x64.zip** — EXE + README + license + documentation.
- **SHA256SUMS.txt** — checksums.
