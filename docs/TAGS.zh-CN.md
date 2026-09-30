# EasyRenamer 标签

<p align="center"><a href="TAGS.md">English</a> · <a href="TAGS.ru.md">Русский</a> · <a href="TAGS.es.md">Español</a> · <strong>中文</strong></p>

EasyRenamer 支持大量与 Advanced Renamer 风格相近的标签，同时保持 portable 和 self-contained。

标签不区分大小写：

```text
<Name>-<Inc Nr:001>-<Date Modified:yyyy-mm-dd>
```

标签引擎也支持 fallback 链和 modifiers。

## Fallback

`|` 表示文字备用值，`||` 表示下一个标签：

```text
<Artist|Unknown artist>
<Artist||Performer||Singer|Unknown>
```

## Modifiers

| Modifier | 示例 | 效果 |
| --- | --- | --- |
| `default` | `<Artist:default:Unknown>` | 标签为空时使用默认值 |
| `alt` | `<Artist:alt:Performer>` | 为空时使用另一个标签 |
| `append` | `<Artist:append: - >` | 非空时追加文本 |
| `prepend` | `<Artist:prepend:By >` | 前置文本 |
| `suffix` / `prefix` |  | 保证后缀/前缀 |
| `pad` | `<Track:pad:2:0>` | 左侧补位 |
| `trim` | `<Title:trim>` | 去除首尾空格 |
| `upper` / `lower` / `titlecase` |  | 大小写转换 |
| `substr` / `rsubstr` |  | 提取子串 |
| `insert` / `remove` / `replace` |  | 修改值 |
| `word` / `rword` |  | 提取单词 |
| `add` / `subtract` / `multiply` / `divide` |  | 数值运算 |

## 核心标签

- `<Name>` — 当前名称（不含扩展名）
- `<Ext>` — 扩展名（不含点）
- `<FolderName:1>` / `<DirName:1>` — 父文件夹
- `<Inc Nr:start:step>` — 全局计数器
- `<Inc NrDir:start:step>` — 每个目录独立计数
- `<Dec Nr:start:step>`
- `<Inc Alpha:start:step>`
- `<Inc Hex:start:step>`
- `<Inc Roman:start:step>`
- `<Rand:min:max>`
- `<Rand Str:length>` / `<Rand Alpha:length>`
- `<Num Items:000>`, `<Num Files:000>`, `<Num Dirs:000>`
- `<Word:index:count:separator>`, `<RWord:index:count:separator>`
- `<Substr:pos:count>`, `<RSubstr:pos:count>`
- `<Switch:A:B:C>`
- `<Delimiter:text>` / `<-:text>`
- `<SubFolder:index>`
- `<MediaType>`
- `<File Line:line>`
- `<File Content:pos:count>`
- `<MetaData:fieldname>`

## 日期与时间

批处理时间：`<Date:yyyy-mm-dd>`、`<Time:hh:nn:ss>`、`<Sec>`、`<Min>`、`<Hour>`、`<Day>`、`<Month>`、`<Year>`、`<UnixTimestamp>`。

文件创建时间：`<Date Created:yyyy-mm-dd>`、`<Time Created:hh:nn:ss>` 及各时间部分。

文件修改时间：`<Date Modified:yyyy-mm-dd>`、`<Time Modified:hh:nn:ss>` 及各时间部分。

支持 `yyyy`、`yy`、`mm`、`mmm`、`mmmm`、`dd`、`ddd`、`dddd`、`hh`、`nn`、`ss`。

## 图片与 EXIF

JPEG/TIFF 支持：

- `<Width>`、`<Height>`
- `<Img Year>`、`<Img Month>`、`<Img Day>`
- `<Img Hour>`、`<Img Min>`、`<Img Sec>`、`<Img Subsec>`
- `<Img DateOriginal:pattern>`、`<Img DateCreate:pattern>`
- `<Img TimeOriginal:pattern>`、`<Img TimeCreate:pattern>`
- `<Img DPI>`
- `<Author>`、`<Copyright>`、`<Subject>`、`<Title>`

## GPS

离线支持 EXIF 坐标：

- `<GPS Lat>` 及 Lat Deg/Min/Sec/Dir
- `<GPS Lng>` 及 Lng Deg/Min/Sec/Dir
- `<GPS Alt>`

EasyRenamer 不内置城市/国家/地区 reverse geocoding，因为这需要地理数据库或网络服务。

## 视频

MP4/MOV/M4V/3GP 支持 `<Width>`、`<Height>`、`<FrameRate>`、`<Title>`、`<Genre>`、Video Date/Time 和 Duration 标签。

## 音频

MP3 ID3 和 FLAC 支持 `<Album>`、`<Artist>`、`<Genre>`、`<Title>`、`<Audio Year>`、`<Track:00>`、`<TrackCount:00>`、`<Disc:00>`、`<DiscCount:00>`。

## 文档与邮件

PDF、Office Open XML、EPUB、EML 支持 `<Pages:000>`、`<Creator>`、`<Author>`、`<Title>`、`<Subject>`、`<Date>` 以及 From/To/Cc/Bcc 系列标签。

## 文件大小与校验和

- `<Filesize Text>`、`<Filesize B:000>`、`<Filesize Kb:000>`、`<Filesize Mb:000>`、`<Filesize Gb:000>`、`<Filesize Tb:000>`
- `<MD5>`、`<SHA1>`

校验和需要读取文件内容，因此比普通名称标签更慢。

## Windows EXE

支持 `<Exe Product>`、`<Exe Version>`、`<Exe VersionMajor>`、`<Exe VersionMinor>`、`<Exe FileVersion>`、`<Exe Company>`、`<Description>`。

## ExifTool-style 与 CSV 元数据

EasyRenamer 不打包外部 ExifTool，常见元数据由内置 parser 读取。

`<MetaData:fieldname>` 是未来 importer 的扩展点。CSV 列标签和任意 ExifTool 字段目前还没有完整实现。

## 兼容性

标签词汇对 Advanced Renamer 用户较熟悉，但 EasyRenamer 的实现是独立的。本文件描述的是 EasyRenamer 的实际行为。
