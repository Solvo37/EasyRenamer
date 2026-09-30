# EasyRenamer v1.0.1

Polish release for the v1 desktop experience.

## UI and branding

- Added a lightweight EasyRenamer vector mark used in the title bar and as the app favicon.
- Replaced remaining text-glyph controls with consistent Lucide icons.
- Added localized tooltips for window controls, splitters, view controls and method actions.
- Cleaned obsolete CSS left from the pre-1.0 layout.

## Localization

- Finished the v1 workspace translations for **English, Русский, Español and 中文**.
- Removed RU/EN-only hardcoded labels from the React UI.
- Localized sorting, grouping, collision policies, history, context actions, presets and common validation errors.
- Fixed Functions / Variables labels for languages that do not use an ASCII colon.
- Updated the built-in help text for the current v1 workflow.
- Removed 121 obsolete translation keys that belonged to the legacy UI.
- Active translation keys now match the current frontend/backend usage without missing or unused entries.

## GitHub and documentation

- Rebuilt README as a compact project landing page with logo, badges, direct latest download and current feature overview.
- Updated LEARN and ROADMAP for the shipped v1 workflow.
- Improved CONTRIBUTING.
- Added SECURITY guidance, bug/feature issue forms and a pull-request template.
- Removed the obsolete duplicate token reference.

## Cleanup

- Removed the complete legacy Walk frontend from `cmd/easyrenamer/`.
- Removed old Walk/Win and stale direct x/sys dependencies.
- Kept the Wails + React application as the only production desktop UI.
- Updated ignore rules for generated Wails output.

## Validation

The release pipeline runs:

- `go test ./internal/...`
- production Windows Wails build
- startup smoke test for `EasyRenamer.exe`

## Downloads

- **EasyRenamer.exe** — portable Windows x64 executable.
- **EasyRenamer-v1.0.1-windows-x64.zip** — EXE + README + license + documentation.
- **SHA256SUMS.txt** — checksums.
