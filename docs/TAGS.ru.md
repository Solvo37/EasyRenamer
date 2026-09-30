# Теги EasyRenamer

<p align="center"><a href="TAGS.md">English</a> · <strong>Русский</strong> · <a href="TAGS.es.md">Español</a> · <a href="TAGS.zh-CN.md">中文</a></p>

EasyRenamer поддерживает большой набор тегов в стиле Advanced Renamer, оставаясь portable и self-contained.

Теги регистронезависимы и записываются в угловых скобках:

```text
<Name>-<Inc Nr:001>-<Date Modified:yyyy-mm-dd>
```

Движок тегов также поддерживает fallback-цепочки и modifiers. EasyRenamer реализует их самостоятельно; ниже описано поведение именно EasyRenamer.

## Fallback

`|` задаёт буквальное запасное значение, `||` — следующий тег:

```text
<Artist|Unknown artist>
<Artist||Performer||Singer|Unknown>
```

## Modifiers

Modifiers добавляются через `:`.

| Modifier | Пример | Эффект |
| --- | --- | --- |
| `default` | `<Artist:default:Unknown>` | Значение, если тег пуст |
| `alt` | `<Artist:alt:Performer>` | Использовать другой тег, если пусто |
| `append` | `<Artist:append: - >` | Добавить текст справа, если значение не пустое |
| `prepend` | `<Artist:prepend:By >` | Добавить текст слева |
| `suffix` | `<Name:suffix:_>` | Гарантировать суффикс |
| `prefix` | `<Name:prefix:_>` | Гарантировать префикс |
| `pad` | `<Track:pad:2:0>` | Дополнить слева |
| `trim` | `<Title:trim>` | Убрать пробелы по краям |
| `upper` | `<Artist:upper>` | Верхний регистр |
| `lower` | `<Artist:lower>` | Нижний регистр |
| `titlecase` | `<Name:titlecase>` | Title Case |
| `substr` | `<Name:substr:2:5>` | Подстрока |
| `rsubstr` | `<Name:rsubstr:2:5>` | Подстрока справа |
| `insert` | `<Name:insert:_x_:3>` | Вставить текст |
| `remove` | `<Name:remove:2:4>` | Удалить символы |
| `replace` | `<Name:replace:old:new>` | Заменить текст |
| `word` / `rword` | `<Name:word:2:1>` | Извлечь слова |
| `add` / `subtract` / `multiply` / `divide` | `<Track:add:1>` | Числовая операция |

## Основные теги

- `<Name>` — текущее имя без расширения
- `<Ext>` — расширение без точки
- `<FolderName:1>` / `<DirName:1>` — родительские папки
- `<Inc Nr:start:step>` — общий счётчик
- `<Inc NrDir:start:step>` — счётчик отдельно по папкам
- `<Dec Nr:start:step>` — убывающий счётчик
- `<Inc Alpha:start:step>` — буквенный счётчик
- `<Inc Hex:start:step>` — шестнадцатеричный счётчик
- `<Inc Roman:start:step>` — римские числа
- `<Rand:min:max>` — случайное число
- `<Rand Str:length>` / `<Rand Alpha:length>` — случайные строки
- `<Num Items:000>`, `<Num Files:000>`, `<Num Dirs:000>`
- `<Word:index:count:separator>`, `<RWord:index:count:separator>`
- `<Substr:pos:count>`, `<RSubstr:pos:count>`
- `<Switch:A:B:C>`
- `<Delimiter:text>` или `<-:text>`
- `<SubFolder:index>`
- `<MediaType>`
- `<File Line:line>`
- `<File Content:pos:count>`
- `<MetaData:fieldname>`

## Дата и время

Время batch:

- `<Date:yyyy-mm-dd>`
- `<Time:hh:nn:ss>`
- `<Sec>`, `<Min>`, `<Hour>`, `<Day>`, `<Month>`, `<Year>`
- `<UnixTimestamp>`

Время создания файла:

- `<Date Created:yyyy-mm-dd>`
- `<Time Created:hh:nn:ss>`
- `<Sec Created>`, `<Min Created>`, `<Hour Created>`, `<Day Created>`, `<Month Created>`, `<Year Created>`, `<UnixTimestamp Created>`

Время изменения файла:

- `<Date Modified:yyyy-mm-dd>`
- `<Time Modified:hh:nn:ss>`
- `<Sec Modified>`, `<Min Modified>`, `<Hour Modified>`, `<Day Modified>`, `<Month Modified>`, `<Year Modified>`, `<UnixTimestamp Modified>`

Поддерживаются части формата `yyyy`, `yy`, `mm`, `mmm`, `mmmm`, `dd`, `ddd`, `dddd`, `hh`, `nn`, `ss`.

## Изображения и EXIF

Для JPEG/TIFF EasyRenamer читает распространённые EXIF-данные напрямую:

- `<Width>`, `<Height>`
- `<Img Year>`, `<Img Month>`, `<Img Day>`
- `<Img Hour>`, `<Img Min>`, `<Img Sec>`, `<Img Subsec>`
- `<Img DateOriginal:pattern>`, `<Img DateCreate:pattern>`
- `<Img TimeOriginal:pattern>`, `<Img TimeCreate:pattern>`
- `<Img DPI>`
- `<Author>`, `<Copyright>`, `<Subject>`, `<Title>`

## GPS

Офлайн поддерживаются координаты EXIF:

- `<GPS Lat>`, `<GPS Lat Deg>`, `<GPS Lat Min>`, `<GPS Lat Sec>`, `<GPS Lat Dir>`
- `<GPS Lng>`, `<GPS Lng Deg>`, `<GPS Lng Min>`, `<GPS Lng Sec>`, `<GPS Lng Dir>`
- `<GPS Alt>`

Reverse geocoding города/страны/региона намеренно не выполняется встроенно, потому что требует базы мест или сетевого сервиса.

## Видео

Для MP4/MOV/M4V/3GP:

- `<Width>`, `<Height>`
- `<FrameRate>`
- `<Title>`, `<Genre>`
- `<Video Date:pattern>`, `<Video Time:pattern>`
- `<Video Date Year>`, `<Video Date Month>`, `<Video Date Day>`
- `<Video Date Hour>`, `<Video Date Min>`, `<Video Date Sec>`
- `<Duration>`, `<Duration Hour>`, `<Duration Min>`, `<Duration Sec>`

## Аудио

Встроенные readers поддерживают распространённые поля MP3 ID3 и FLAC:

- `<Album>`, `<Artist>`, `<Genre>`, `<Title>`
- `<Audio Year>`
- `<Track:00>`, `<TrackCount:00>`
- `<Disc:00>`, `<DiscCount:00>`

## Документы и e-mail

Для PDF, Office Open XML, EPUB и EML:

- `<Pages:000>`
- `<Creator>`, `<Author>`, `<Title>`, `<Subject>`
- `<Date>`
- `<From>`, `<FromName>`, `<FromEmail>`
- `<To:1>`, `<ToName:1>`, `<ToEmail:1>`
- `<Cc:1>`, `<CcName:1>`, `<CcEmail:1>`
- `<Bcc:1>`, `<BccName:1>`, `<BccEmail:1>`

## Размер файла

- `<Filesize Text>`
- `<Filesize B:000>`
- `<Filesize Kb:000>`
- `<Filesize Mb:000>`
- `<Filesize Gb:000>`
- `<Filesize Tb:000>`

## Контрольные суммы

- `<MD5>`
- `<SHA1>`

Checksums требуют чтения содержимого файла и поэтому медленнее обычных тегов имени.

## Windows EXE

Из PE version resources читаются:

- `<Exe Product>`
- `<Exe Version>`
- `<Exe VersionMajor>`
- `<Exe VersionMinor>`
- `<Exe FileVersion>`
- `<Exe Company>`
- `<Description>`

## ExifTool-style и CSV metadata

EasyRenamer не включает внешний ExifTool. Распространённые image/audio/video/document/GPS/EXE данные читаются встроенными средствами.

`<MetaData:fieldname>` — точка расширения для metadata, которую смогут предоставлять будущие importers. CSV-column tags и произвольные ExifTool fields пока реализованы не полностью.

## Совместимость

Словарь тегов сделан знакомым пользователям Advanced Renamer, но реализация EasyRenamer независимая. Актуальные возможности EasyRenamer описаны именно в этом файле.
