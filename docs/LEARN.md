# EasyRenamer — User Guide

<p align="center"><strong>English</strong> · <a href="LEARN.ru.md">Русский</a> · <a href="LEARN.es.md">Español</a> · <a href="LEARN.zh-CN.md">中文</a></p>

Practical guide for EasyRenamer v1.x.

## Quick start

1. Add files with **+ Files**, folders with **+ Folders**, or drag them from Explorer.
2. Enable or disable **Include subfolders** and the extension filter as needed.
3. Build the method chain on the left.
4. Choose the file processing order.
5. Review **New filename** and error status.
6. Click **Start**.
7. Use **Ctrl+Z** or **History** when you need to roll an operation back.

Preview recalculates automatically. **Check** forces a fresh validation of the current task.

## Files and folders

Files from several directories can be processed in the same task.

With folder grouping enabled, every directory is shown as its own group. A group can be collapsed, selected, deselected, removed from the task, opened in Explorer, or have its path copied.

Removing an item from the list **does not delete it from disk**.

## Processing order

Order matters for sequential numbering and the List method.

Available order modes:

- natural filename;
- creation date;
- modification date;
- size;
- extension;
- path;
- added order;
- manual order.

Natural sorting gives:

```text
1.jpg
2.jpg
10.jpg
```

instead of `1.jpg, 10.jpg, 2.jpg`.

**Sort separately inside each folder** keeps folder groups together and sorts the contents of each group independently.

### Manual order

Choose **Manual order** and drag rows. Preview and numbering are recalculated using the new order.

## Per-folder numbering

The tag:

```text
<Inc NrDir:01>
```

restarts the counter for each directory.

```text
Folder A:
1.jpg → 01_1.jpg
2.jpg → 02_2.jpg

Folder B:
a.jpg → 01_a.jpg
b.jpg → 02_b.jpg
```

## Method chain

Methods are applied **top to bottom**. The result of one method becomes the input for the next.

A method can be enabled/disabled, selected for editing, moved up/down, duplicated, deleted, and saved as part of a method set.

## Methods

### New Name

Builds a name from text and tags:

```text
<Inc NrDir:01>_<Name>
```

Type `<` to open tag autocomplete. Use ↑ / ↓, Enter and Esc.

### Replace

Find and replace plain text or regular expressions.

### Renumber

Adds a sequential number as a prefix or suffix. Start, Step, Padding, Separator and per-folder reset are configurable.

### Add text

Adds a prefix and/or suffix.

### Change case

Changes the case of the base filename: lower case, UPPER CASE or Title Case.

### Remove

Removes a number of characters starting at a selected position.

### Remove pattern

Removes text or regular-expression matches.

### Move

Moves a part of the filename to another position.

### Swap

Swaps two parts around a selected separator.

### Trim

Trims surrounding whitespace and can collapse repeated whitespace.

### Timestamp

Adds the file modified time or current batch time to the filename.

### List

Assigns final names line by line.

### List Replace

Applies several replacement rules:

```text
draft => final
IMG_ => product_
```

### Script

Evaluates a safe expression with no direct filesystem or network access.

Variables: `Name`, `Ext`, `FullName`, `Index`, `DirIndex`, `DirName`, `UnixTimestamp`, `ModifiedUnix`.

Functions: `lower()`, `upper()`, `trim()`, `replace()`, `concat()`, `substr()`.

## Tags

Full reference: [TAGS.md](TAGS.md).

Frequently used:

```text
<Name>
<Ext>
<FolderName:1>
<Inc Nr:001>
<Inc NrDir:01>
<Date Modified:yyyy-mm-dd>
<Width>
<Height>
<Artist>
<Album>
<MD5>
```

Tags support fallback chains and modifiers.

## User presets

In **New Name**, the current template can be saved as a user preset. Presets can be applied, renamed, duplicated and deleted. They persist between launches.

## Search and table

Search filters the loaded list by original name, new name and full path. **Errors only** limits the view to problem rows.

Columns can be hidden, reordered and resized. The configuration is persisted.

Row selection supports Ctrl+Click, Shift+Click and Ctrl+A. The file checkbox controls participation in the rename operation independently of row selection.

## Collision policy

Default: **skip conflicting files**.

Available policies:

- skip conflicts;
- add a number automatically;
- stop the operation on conflict.

Destructive overwrite is shown as unavailable until safe backup/restore exists.

## Windows filename validation

Before execution EasyRenamer checks forbidden characters, empty or whitespace-only names, trailing dots/spaces, reserved names such as CON/PRN/AUX/NUL/COM1…COM9/LPT1…LPT9, duplicate destinations, existing targets, path length and missing files.

## Safe rename

Files first receive unique temporary names and only then their final names. This safely supports swaps such as:

```text
A.jpg → B.jpg
B.jpg → A.jpg
```

On Cancel, EasyRenamer stops and attempts to restore the original names without leaving temporary files behind.

## Undo and history

Successful operations store the actual full-path pairs:

```text
old path → new path
```

Undo does not recalculate anything from the current template.

History keeps up to 20 recent operations. An older operation can be rolled back while its files remain at the expected paths and the original names are free.

## Selected-file preview

For a selected file EasyRenamer can show a thumbnail, size, type, image dimensions, original/new name and visual name diff. Heavy image preview data is loaded lazily.

## Keyboard shortcuts

| Shortcut | Action |
| --- | --- |
| `Ctrl + O` | Add files |
| `Ctrl + Shift + O` | Add folders |
| `Ctrl + A` | Select rows |
| `Ctrl + Z` | Undo |
| `Delete` | Remove selected items from the task |
| `Ctrl + Enter` | Start |
| `F5` | Recalculate / check |
| `Esc` | Cancel current operation |

## Metadata

EasyRenamer is self-contained and does not require ExifTool for its common metadata readers.

Built-in readers cover common JPEG/TIFF EXIF and GPS, MP3 ID3, FLAC Vorbis comments, MP4/MOV, PDF, Office Open XML, EPUB, EML and Windows EXE version resources.

## Languages and theme

Supported UI languages: English, Русский, Español and 中文.

Theme modes: System / Light / Dark.

Language and theme changes do not require restarting the application.

## More

Roadmap: [ROADMAP.md](ROADMAP.md)  
Source and releases: [GitHub](https://github.com/Solvo37/easyrenamer)
