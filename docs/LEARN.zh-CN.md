# EasyRenamer — 使用指南

<p align="center"><a href="LEARN.md">English</a> · <a href="LEARN.ru.md">Русский</a> · <a href="LEARN.es.md">Español</a> · <strong>中文</strong></p>

EasyRenamer v1.x 的实用指南。

## 快速开始

1. 使用 **+ 文件**、**+ 文件夹**，或从资源管理器拖入项目。
2. 根据需要启用/禁用 **包含子文件夹** 和扩展名筛选。
3. 在左侧建立方法链。
4. 选择文件处理顺序。
5. 检查 **新文件名** 和错误状态。
6. 点击 **开始**。
7. 需要回滚时使用 **Ctrl+Z** 或 **历史记录**。

预览会自动重新计算，**检查**按钮会强制重新验证当前任务。

## 文件与文件夹

一个任务可以同时处理多个目录中的文件。启用文件夹分组后，每个目录会显示为独立分组，可以折叠、整组选中、取消选择、从任务中移除、在资源管理器中打开或复制路径。

从列表移除项目**不会删除磁盘上的文件**。

## 处理顺序

处理顺序会影响连续编号和 List 方法。支持自然名称、创建日期、修改日期、大小、扩展名、路径、添加顺序和手动顺序。

自然排序得到 `1, 2, 10`，而不是 `1, 10, 2`。

**每个文件夹内分别排序**会保持文件夹分组，并独立排序每组内容。

### 手动顺序

选择 **手动顺序** 后拖动行，预览和编号会按新顺序重新计算。

## 按文件夹编号

`<Inc NrDir:01>` 会在每个目录中重新开始计数。

```text
文件夹 A:
1.jpg → 01_1.jpg
2.jpg → 02_2.jpg

文件夹 B:
a.jpg → 01_a.jpg
b.jpg → 02_b.jpg
```

## 方法链

方法按**从上到下**的顺序执行。一个方法的结果会成为下一个方法的输入。

方法可以启用/禁用、编辑、上移/下移、复制、删除并保存到方法集。

## 方法

- **New Name** — 使用文本和标签生成名称；输入 `<` 可打开标签自动补全。
- **Replace** — 普通文本或正则表达式查找替换。
- **Renumber** — 添加连续编号，可配置 Start、Step、Padding 和按文件夹重置。
- **Add text** — 添加前缀/后缀。
- **Change case** — lower、UPPER、Title Case。
- **Remove** — 从指定位置删除字符。
- **Remove pattern** — 删除文本或 regex 匹配。
- **Move** — 移动文件名中的一部分。
- **Swap** — 交换分隔符两侧内容。
- **Trim** — 清理空格。
- **Timestamp** — 把文件修改时间或当前批处理时间加入文件名。
- **List** — 按行指定最终名称。
- **List Replace** — 应用多条 `查找 => 替换` 规则。
- **Script** — 安全表达式，不直接访问文件系统或网络。

Script 变量：`Name`、`Ext`、`FullName`、`Index`、`DirIndex`、`DirName`、`UnixTimestamp`、`ModifiedUnix`。

函数：`lower()`、`upper()`、`trim()`、`replace()`、`concat()`、`substr()`。

## 标签

完整参考：[TAGS.zh-CN.md](TAGS.zh-CN.md)。

常用标签包括 `<Name>`、`<Ext>`、`<FolderName:1>`、`<Inc Nr:001>`、`<Inc NrDir:01>`、`<Date Modified:yyyy-mm-dd>`、`<Width>`、`<Artist>`、`<MD5>`。

标签支持 fallback 链和 modifiers。

## 用户预设

在 **New Name** 中可以把当前模板保存为用户预设，随后可以应用、重命名、复制和删除，并会在下次启动时保留。

## 搜索与表格

搜索支持原始名称、新名称和完整路径。**仅错误**只显示问题项。

列可以隐藏、调整顺序和宽度。Ctrl+Click、Shift+Click、Ctrl+A 控制行选择；文件 checkbox 独立控制是否参与重命名。

## 冲突策略

默认跳过冲突文件。也可以自动添加编号或在发生冲突时停止。破坏性覆盖保持禁用，直到实现安全的 backup/restore。

## Windows 文件名验证

执行前会检查非法字符、空名称/全空格名称、结尾点号或空格、Windows 保留名称、重复或已存在目标、路径长度和已消失的源文件。

## 安全重命名

EasyRenamer 先给文件分配唯一临时名称，再改为最终名称，因此 `A.jpg ↔ B.jpg` 等交换场景可以安全处理。取消操作时会尝试恢复原始名称并清理临时文件。

## Undo 与历史记录

成功操作会保存真实的 `旧路径 → 新路径`。历史记录最多保存 20 次操作。

## 选中文件预览

可以显示缩略图、大小、类型、图片尺寸、原始/新名称以及名称差异。较重的数据按需加载。

## 快捷键

| 快捷键 | 操作 |
| --- | --- |
| `Ctrl + O` | 添加文件 |
| `Ctrl + Shift + O` | 添加文件夹 |
| `Ctrl + A` | 选择行 |
| `Ctrl + Z` | Undo |
| `Delete` | 从任务中移除 |
| `Ctrl + Enter` | 开始 |
| `F5` | 重新计算 / 检查 |
| `Esc` | 取消 |

## 元数据

内置读取器覆盖常见 JPEG/TIFF EXIF/GPS、MP3 ID3、FLAC、MP4/MOV、PDF、Office Open XML、EPUB、EML 和 Windows EXE 版本资源。

## 语言与主题

语言：English、Русский、Español、中文。  
主题：System / Light / Dark。

## 更多

Roadmap：[ROADMAP.zh-CN.md](ROADMAP.zh-CN.md)  
源码与 Releases：[GitHub](https://github.com/Solvo37/easyrenamer)
