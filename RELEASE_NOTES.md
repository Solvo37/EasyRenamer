# EasyRenamer v0.5.0

Major desktop UX rebuild.

## Workspace

- The window is reorganized around a professional batch-renamer workflow.
- Left side: rename method stack, settings for the selected method, and Add method.
- Right side: the large file preview/list area.
- The right side is no longer consumed by a second method-selection UI.
- Start batch is more visually prominent.
- The selected method is preserved when the UI is rebuilt.

## Language and theme

- Language changes apply without restarting EasyRenamer.
- System / Light / Dark theme changes apply without restarting EasyRenamer.
- EasyRenamer recreates its window inside the same running process so all translated labels and Win32 theme surfaces are rebuilt consistently.
- Current sources, method stack, selected method, file filter, custom extensions, recursive setting, and live-preview setting are preserved.
- If files are already loaded, preview is rebuilt automatically after the UI refresh.

## Existing v0.4.5 fixes retained

- The method stack remains the single source of truth.
- Only the selected method settings are shown.
- Extensions are always editable.
- Typing extensions switches to the Custom filter.
- Filter dropdown no longer immediately closes in dark mode.

## Quality

- Core tests must pass.
- Windows build must pass.
- The produced EXE must pass the startup smoke test before publication.

## Downloads

- **EasyRenamer.exe** — portable Windows x64 executable.
- **EasyRenamer-v0.5.0-windows-x64.zip** — EXE + README + license + documentation.
- **SHA256SUMS.txt** — checksums.
