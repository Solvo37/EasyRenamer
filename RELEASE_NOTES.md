# EasyRenamer v0.4.5

UX clarity hotfix.

## Method workflow

- The method stack on the left is now the single source of truth.
- The right pane no longer acts like a second method selector.
- Only the settings page for the currently selected left-side method is shown.
- Editing fields on the right can no longer change a method from one type to another.
- The settings title explicitly names the selected method.

This removes the previous ambiguity where users could add **New Name** on the left but also see clickable **New Name / Replace / Move / ...** tabs on the right.

## Filters

- The **Extensions** field is now always editable.
- It starts empty instead of showing a disabled example value.
- Typing extensions such as `jpg, png, psd` automatically switches the category to **Custom**.
- Choosing another category still works normally and does not erase the typed extensions.

## ComboBox dropdown

- Fixed a dark-theme interaction where opening the Filter dropdown could immediately close it.
- Popup theming now targets the transient dropdown only and no longer re-themes the entire owner window while the dropdown is opening.

## Quality

- Core tests pass.
- Windows build passes.
- The produced EXE passes the startup smoke test before publication.

## Downloads

- **EasyRenamer.exe** — portable Windows x64 executable.
- **EasyRenamer-v0.4.5-windows-x64.zip** — EXE + README + license + documentation.
- **SHA256SUMS.txt** — checksums.
