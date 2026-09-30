# Etiquetas de EasyRenamer

<p align="center"><a href="TAGS.md">English</a> · <a href="TAGS.ru.md">Русский</a> · <strong>Español</strong> · <a href="TAGS.zh-CN.md">中文</a></p>

EasyRenamer admite un amplio vocabulario de etiquetas al estilo de Advanced Renamer, manteniendo la aplicación portable y autocontenida.

Las etiquetas no distinguen mayúsculas/minúsculas:

```text
<Name>-<Inc Nr:001>-<Date Modified:yyyy-mm-dd>
```

También se admiten cadenas de fallback y modificadores.

## Fallback

Usa `|` para un valor literal alternativo y `||` para otra etiqueta:

```text
<Artist|Unknown artist>
<Artist||Performer||Singer|Unknown>
```

## Modificadores

| Modificador | Ejemplo | Efecto |
| --- | --- | --- |
| `default` | `<Artist:default:Unknown>` | Valor cuando la etiqueta está vacía |
| `alt` | `<Artist:alt:Performer>` | Usa otra etiqueta |
| `append` | `<Artist:append: - >` | Añade texto al final |
| `prepend` | `<Artist:prepend:By >` | Añade texto al principio |
| `suffix` | `<Name:suffix:_>` | Garantiza un sufijo |
| `prefix` | `<Name:prefix:_>` | Garantiza un prefijo |
| `pad` | `<Track:pad:2:0>` | Rellena por la izquierda |
| `trim` | `<Title:trim>` | Recorta espacios |
| `upper` / `lower` / `titlecase` |  | Cambia el uso de mayúsculas |
| `substr` / `rsubstr` |  | Extrae una subcadena |
| `insert` / `remove` / `replace` |  | Edita el valor |
| `word` / `rword` |  | Extrae palabras |
| `add` / `subtract` / `multiply` / `divide` |  | Operación numérica |

## Etiquetas principales

- `<Name>` — nombre actual sin extensión
- `<Ext>` — extensión sin punto
- `<FolderName:1>` / `<DirName:1>` — carpetas superiores
- `<Inc Nr:start:step>` — contador global
- `<Inc NrDir:start:step>` — contador por directorio
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

## Fecha y hora

Hora del lote: `<Date:yyyy-mm-dd>`, `<Time:hh:nn:ss>`, `<Sec>`, `<Min>`, `<Hour>`, `<Day>`, `<Month>`, `<Year>`, `<UnixTimestamp>`.

Creación: `<Date Created:yyyy-mm-dd>`, `<Time Created:hh:nn:ss>` y sus partes.

Modificación: `<Date Modified:yyyy-mm-dd>`, `<Time Modified:hh:nn:ss>` y sus partes.

Partes de formato: `yyyy`, `yy`, `mm`, `mmm`, `mmmm`, `dd`, `ddd`, `dddd`, `hh`, `nn`, `ss`.

## Imagen y EXIF

Para JPEG/TIFF:

- `<Width>`, `<Height>`
- `<Img Year>`, `<Img Month>`, `<Img Day>`
- `<Img Hour>`, `<Img Min>`, `<Img Sec>`, `<Img Subsec>`
- `<Img DateOriginal:pattern>`, `<Img DateCreate:pattern>`
- `<Img TimeOriginal:pattern>`, `<Img TimeCreate:pattern>`
- `<Img DPI>`
- `<Author>`, `<Copyright>`, `<Subject>`, `<Title>`

## GPS

Coordenadas EXIF disponibles sin conexión:

- `<GPS Lat>`, `<GPS Lat Deg>`, `<GPS Lat Min>`, `<GPS Lat Sec>`, `<GPS Lat Dir>`
- `<GPS Lng>`, `<GPS Lng Deg>`, `<GPS Lng Min>`, `<GPS Lng Sec>`, `<GPS Lng Dir>`
- `<GPS Alt>`

EasyRenamer no hace reverse geocoding de ciudad/país/región de forma integrada porque requiere una base geográfica o servicio de red.

## Vídeo

Para MP4/MOV/M4V/3GP: `<Width>`, `<Height>`, `<FrameRate>`, `<Title>`, `<Genre>`, etiquetas Video Date/Time y Duration.

## Audio

MP3 ID3 y FLAC: `<Album>`, `<Artist>`, `<Genre>`, `<Title>`, `<Audio Year>`, `<Track:00>`, `<TrackCount:00>`, `<Disc:00>`, `<DiscCount:00>`.

## Documentos y correo

PDF, Office Open XML, EPUB y EML: `<Pages:000>`, `<Creator>`, `<Author>`, `<Title>`, `<Subject>`, `<Date>`, From/To/Cc/Bcc y variantes Name/Email.

## Tamaño y checksums

- `<Filesize Text>`, `<Filesize B:000>`, `<Filesize Kb:000>`, `<Filesize Mb:000>`, `<Filesize Gb:000>`, `<Filesize Tb:000>`
- `<MD5>`, `<SHA1>`

Los checksums leen el contenido del archivo y son más lentos.

## Windows EXE

- `<Exe Product>`, `<Exe Version>`, `<Exe VersionMajor>`, `<Exe VersionMinor>`, `<Exe FileVersion>`, `<Exe Company>`, `<Description>`

## Metadatos ExifTool-style y CSV

EasyRenamer no incluye ExifTool externo. Los campos comunes se leen con parsers integrados.

`<MetaData:fieldname>` es el punto de extensión para importadores futuros. Las etiquetas de columnas CSV y los campos ExifTool arbitrarios todavía no están completos.

## Compatibilidad

El vocabulario resulta familiar a usuarios de Advanced Renamer, pero la implementación de EasyRenamer es independiente. Este documento describe el comportamiento real de EasyRenamer.
