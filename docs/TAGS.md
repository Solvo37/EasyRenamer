# EasyRenamer Tags

EasyRenamer supports a large Advanced Renamer-style tag vocabulary while keeping the program portable and self-contained.

Tags are case-insensitive and are written inside angle brackets:

```text
<Name>-<Inc Nr:001>-<Date Modified:yyyy-mm-dd>
```

The tag engine also supports fallback chains and modifiers. EasyRenamer implements these independently; this document describes EasyRenamer behavior.

## Fallback

Use `|` for a literal fallback value and `||` for another tag:

```text
<Artist|Unknown artist>
<Artist||Performer||Singer|Unknown>
```

## Modifiers

Modifiers are appended with `:`.

| Modifier | Example | Effect |
| --- | --- | --- |
| `default` | `<Artist:default:Unknown>` | Use a value when the tag is empty |
| `alt` | `<Artist:alt:Performer>` | Use another tag when empty |
| `append` | `<Artist:append: - >` | Append text when non-empty |
| `prepend` | `<Artist:prepend:By >` | Prepend text when non-empty |
| `suffix` | `<Name:suffix:_>` | Ensure a suffix |
| `prefix` | `<Name:prefix:_>` | Ensure a prefix |
| `pad` | `<Track:pad:2:0>` | Left-pad |
| `trim` | `<Title:trim>` | Trim whitespace |
| `upper` | `<Artist:upper>` | Uppercase |
| `lower` | `<Artist:lower>` | Lowercase |
| `titlecase` | `<Name:titlecase>` | Title Case |
| `substr` | `<Name:substr:2:5>` | Substring |
| `rsubstr` | `<Name:rsubstr:2:5>` | Reverse substring |
| `insert` | `<Name:insert:_x_:3>` | Insert text |
| `remove` | `<Name:remove:2:4>` | Remove characters |
| `replace` | `<Name:replace:old:new>` | Replace text |
| `word` / `rword` | `<Name:word:2:1>` | Extract words |
| `add` / `subtract` / `multiply` / `divide` | `<Track:add:1>` | Numeric modifier |

## Core tags

- `<Name>` — current name without extension
- `<Ext>` — extension without the dot
- `<FolderName:1>` / `<DirName:1>` — parent folders
- `<Inc Nr:start:step>` — global counter
- `<Inc NrDir:start:step>` — counter reset per directory
- `<Dec Nr:start:step>` — decreasing counter
- `<Inc Alpha:start:step>` — alphabetic counter
- `<Inc Hex:start:step>` — hexadecimal counter
- `<Inc Roman:start:step>` — Roman-numeral counter
- `<Rand:min:max>` — random number
- `<Rand Str:length>` / `<Rand Alpha:length>` — random letters
- `<Num Items:000>`, `<Num Files:000>`, `<Num Dirs:000>`
- `<Word:index:count:separator>`, `<RWord:index:count:separator>`
- `<Substr:pos:count>`, `<RSubstr:pos:count>`
- `<Switch:A:B:C>`
- `<Delimiter:text>` or `<-:text>`
- `<SubFolder:index>`
- `<MediaType>`
- `<File Line:line>`
- `<File Content:pos:count>`
- `<MetaData:fieldname>`

## Date and time

Batch time:

- `<Date:yyyy-mm-dd>`
- `<Time:hh:nn:ss>`
- `<Sec>`, `<Min>`, `<Hour>`, `<Day>`, `<Month>`, `<Year>`
- `<UnixTimestamp>`

File creation time:

- `<Date Created:yyyy-mm-dd>`
- `<Time Created:hh:nn:ss>`
- `<Sec Created>`, `<Min Created>`, `<Hour Created>`, `<Day Created>`, `<Month Created>`, `<Year Created>`, `<UnixTimestamp Created>`

File modified time:

- `<Date Modified:yyyy-mm-dd>`
- `<Time Modified:hh:nn:ss>`
- `<Sec Modified>`, `<Min Modified>`, `<Hour Modified>`, `<Day Modified>`, `<Month Modified>`, `<Year Modified>`, `<UnixTimestamp Modified>`

Supported date pattern parts include `yyyy`, `yy`, `mm`, `mmm`, `mmmm`, `dd`, `ddd`, `dddd`, `hh`, `nn`, and `ss`.

## Image and EXIF

For JPEG/TIFF EasyRenamer reads common EXIF metadata directly, without needing an external helper:

- `<Width>`, `<Height>`
- `<Img Year>`, `<Img Month>`, `<Img Day>`
- `<Img Hour>`, `<Img Min>`, `<Img Sec>`, `<Img Subsec>`
- `<Img DateOriginal:pattern>`, `<Img DateCreate:pattern>`
- `<Img TimeOriginal:pattern>`, `<Img TimeCreate:pattern>`
- `<Img DPI>`
- `<Author>`, `<Copyright>`, `<Subject>`, `<Title>`

## GPS

Common EXIF GPS coordinate tags are supported offline:

- `<GPS Lat>`, `<GPS Lat Deg>`, `<GPS Lat Min>`, `<GPS Lat Sec>`, `<GPS Lat Dir>`
- `<GPS Lng>`, `<GPS Lng Deg>`, `<GPS Lng Min>`, `<GPS Lng Sec>`, `<GPS Lng Dir>`
- `<GPS Alt>`

Country/city/state reverse geocoding is intentionally not performed by EasyRenamer itself because that requires a location database or network service. If those values are supplied as metadata, the generic metadata tag can still use them.

## Video

For MP4/MOV/M4V/3GP the built-in parser can expose common values:

- `<Width>`, `<Height>`
- `<FrameRate>`
- `<Title>`, `<Genre>`
- `<Video Date:pattern>`, `<Video Time:pattern>`
- `<Video Date Year>`, `<Video Date Month>`, `<Video Date Day>`
- `<Video Date Hour>`, `<Video Date Min>`, `<Video Date Sec>`
- `<Duration>`, `<Duration Hour>`, `<Duration Min>`, `<Duration Sec>`

## Audio

Built-in readers cover common MP3 ID3 and FLAC Vorbis fields:

- `<Album>`
- `<Artist>`
- `<Genre>`
- `<Title>`
- `<Audio Year>`
- `<Track:00>`, `<TrackCount:00>`
- `<Disc:00>`, `<DiscCount:00>`

## Documents and e-mail

Common PDF, Office Open XML, EPUB, and EML fields:

- `<Pages:000>`
- `<Creator>`, `<Author>`, `<Title>`, `<Subject>`
- `<Date>`
- `<From>`, `<FromName>`, `<FromEmail>`
- `<To:1>`, `<ToName:1>`, `<ToEmail:1>`
- `<Cc:1>`, `<CcName:1>`, `<CcEmail:1>`
- `<Bcc:1>`, `<BccName:1>`, `<BccEmail:1>`

## File size

- `<Filesize Text>`
- `<Filesize B:000>`
- `<Filesize Kb:000>`
- `<Filesize Mb:000>`
- `<Filesize Gb:000>`
- `<Filesize Tb:000>`

## Checksums

- `<MD5>`
- `<SHA1>`

Checksums require reading the file content and are therefore slower than ordinary name tags.

## Windows executable version info

On Windows, PE version resources are read directly:

- `<Exe Product>`
- `<Exe Version>`
- `<Exe VersionMajor>`
- `<Exe VersionMinor>`
- `<Exe FileVersion>`
- `<Exe Company>`
- `<Description>`

## ExifTool-style and CSV metadata

EasyRenamer is designed as a single portable executable and does not bundle the external ExifTool program. Common image, audio, video, document, GPS, and executable fields are parsed internally.

The generic form `<MetaData:fieldname>` is the extension point for metadata provided by future importers. CSV-column tags and arbitrary ExifTool fields are not yet complete.

## Compatibility references

The tag vocabulary was designed for familiarity with Advanced Renamer users. The public reference pages used during implementation are:

- https://www.advancedrenamer.com/user_guide/tags
- https://www.advancedrenamer.com/user_guide/tags_advanced
- https://www.advancedrenamer.com/user_guide/tags_datetime
- https://www.advancedrenamer.com/user_guide/tags_datetimecreated
- https://www.advancedrenamer.com/user_guide/tags_datetimemodified
- https://www.advancedrenamer.com/user_guide/tags_image
- https://www.advancedrenamer.com/user_guide/tags_video
- https://www.advancedrenamer.com/user_guide/tags_docs
- https://www.advancedrenamer.com/user_guide/tags_audio
- https://www.advancedrenamer.com/user_guide/tags_gps
- https://www.advancedrenamer.com/user_guide/tags_filesize
- https://www.advancedrenamer.com/user_guide/tags_checksum
- https://www.advancedrenamer.com/user_guide/tags_exe
- https://www.advancedrenamer.com/user_guide/tag_modifiers
- https://www.advancedrenamer.com/user_guide/tag_fallback
