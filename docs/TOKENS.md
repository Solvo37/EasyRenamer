# Template token reference

EasyRenamer templates are evaluated for every item in the preview batch. When several rename methods are used, `<Name>` and `<Ext>` refer to the name produced by the previous method.

## Core tokens

- `<Name>` — current base name without the extension.
- `<Ext>` — current extension without the leading dot. If this token is absent, the current extension is appended automatically.
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

## Compatibility

EasyRenamer supports familiar Advanced Renamer-style tokens, including counters, parent-directory names, dates and random-string tokens. This makes it easier to move existing naming templates into EasyRenamer while still using the preview and transactional execution model.
