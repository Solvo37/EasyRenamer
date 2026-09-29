# Template token reference

EasyRenamer templates are evaluated against every file in the preview batch.

## Core tokens

- `<Name>` — source base name without the extension.
- `<Ext>` — extension without the leading dot. If this token is absent, the original extension is appended automatically.
- `<DirName:1>` — name of the immediate parent directory.
- `<UnixTimestamp>` or `<Unix>` — Unix timestamp captured once when Preview starts.
- `<Date:yyyyMMdd-HHmmss>` — batch date/time using a compact .NET-style pattern.

## Counters

- `<Inc:001>` — global sequence. The number of digits in the spec defines zero padding.
- `<Inc NrDir:01>` — sequence that restarts for each directory.

Examples:

```text
<Inc:001>-<Name>
<Inc NrDir:01>-<Name>
```

## Random tokens

- `<Rand>` — one random digit.
- `<Rand Str:8>` — N random lowercase letters/digits.
- `<Rand Alpha:9>` — N random lowercase letters only.

When a random token follows a numeric counter and Explorer sort order matters, prefer `Rand Alpha` so the counter remains a separate leading numeric run.

## Compatibility goal

EasyRenamer intentionally supports familiar Advanced Renamer-style tokens such as:

```text
<Inc NrDir:01><Rand><Rand Str:8><UnixTimestamp>-<DirName:1>
```

The safer built-in preset uses:

```text
<Inc NrDir:01><Rand Alpha:9><UnixTimestamp>-<DirName:1>
```
