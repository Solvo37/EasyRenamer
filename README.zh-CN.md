<p align="center"><img src="frontend/public/easyrenamer.svg" width="96" height="96" alt="EasyRenamer"></p>
<h1 align="center">EasyRenamer</h1>

<p align="center"><a href="README.md">English</a> · <a href="README.ru.md">Русский</a> · <a href="README.es.md">Español</a> · <strong>中文</strong></p>

<p align="center">面向 Windows 的快速、安全批量文件重命名工具。<br><strong>免费 · 开源 · 无广告 · 无订阅 · 无遥测</strong></p>

<p align="center">
  <a href="https://github.com/Solvo37/easyrenamer/actions/workflows/ci.yml"><img alt="CI" src="https://github.com/Solvo37/easyrenamer/actions/workflows/ci.yml/badge.svg"></a>
  <a href="https://github.com/Solvo37/easyrenamer/releases/latest"><img alt="Release" src="https://img.shields.io/github/v/release/Solvo37/easyrenamer?display_name=tag"></a>
  <a href="LICENSE"><img alt="License" src="https://img.shields.io/github/license/Solvo37/easyrenamer"></a>
  <img alt="Windows x64" src="https://img.shields.io/badge/Windows-x64-0078D4?logo=windows11&logoColor=white">
</p>

<p align="center">
  <a href="https://github.com/Solvo37/easyrenamer/releases/latest/download/EasyRenamer.exe"><strong>下载 EasyRenamer.exe</strong></a>
  · <a href="docs/LEARN.md">使用指南</a>
  · <a href="docs/TAGS.md">标签参考</a>
  · <a href="docs/ROADMAP.md">Roadmap</a>
</p>

---

## 为什么选择 EasyRenamer

EasyRenamer 面向真实的批量重命名工作：从几个文件到分布在多个文件夹中的大型文件集合。所有最终名称都会在<strong>真正修改前</strong>显示，处理顺序清晰可控，冲突会提前检查，重命名操作采用事务式流程并可回滚。

<strong>添加文件 → 配置方法 → 检查结果 → 执行</strong>

## 主要功能

| 功能 | 作用 |
| --- | --- |
| 实时预览 | 修改模板、方法、排序或筛选条件后自动重新计算名称 |
| 文件夹分组 | 清晰显示文件夹边界和每个文件夹独立编号 |
| 自然排序 | 1、2、10，而不是 1、10、2 |
| 处理顺序 | 名称、创建/修改日期、大小、扩展名、路径、添加顺序或手动拖放 |
| 方法链 | 启用、禁用、复制和重新排序重命名方法 |
| 搜索与错误筛选 | 搜索原始名称、新名称和完整路径 |
| 冲突策略 | 跳过冲突、自动编号或停止操作 |
| 进度 + 取消 | 大型操作保持响应并可安全停止 |
| Undo + 历史记录 | 撤销操作并保存最近 20 次重命名记录 |
| 大型列表 | 虚拟化表格和延迟加载较重的预览数据 |
| 4 种界面语言 | English、Русский、Español、中文 |
| 便携版 | 单个 EXE，无需安装程序 |

## 示例

模板：<code>&lt;Inc NrDir:01&gt;_&lt;Name&gt;</code>

<pre>
1.jpg   → 01_1.jpg
2.jpg   → 02_2.jpg
10.jpg  → 03_10.jpg
</pre>

每个文件夹开始时，<code>NrDir</code> 都会重新从 <code>01</code> 计数。

## 重命名方法

EasyRenamer 支持方法链：

- <strong>New Name</strong> — 使用文本和标签生成名称；
- <strong>Replace</strong> — 查找与替换，支持 regex；
- <strong>Renumber</strong> — 顺序编号，也支持每个文件夹独立编号；
- <strong>Add text</strong> — 前缀和后缀；
- <strong>Change case</strong> — lower / UPPER / Title Case；
- <strong>Remove / Remove pattern</strong> — 按位置、文本或 regex 删除；
- <strong>Move / Swap</strong> — 调整文件名部分顺序；
- <strong>Trim</strong> — 清理多余空格；
- <strong>Timestamp</strong> — 添加文件或批处理时间；
- <strong>List / List Replace</strong> — 按行指定名称或规则；
- <strong>Script</strong> — 安全表达式，不直接访问网络或文件系统。

详细说明：[docs/LEARN.md](docs/LEARN.md)。

## 标签与元数据

模板支持计数器、文件夹、日期、大小、随机值、校验和以及常见元数据：

<pre>
&lt;Name&gt;
&lt;Ext&gt;
&lt;FolderName:1&gt;
&lt;Inc Nr:001&gt;
&lt;Inc NrDir:01&gt;
&lt;Date Modified:yyyy-mm-dd&gt;
&lt;Artist&gt;
&lt;Album&gt;
&lt;Width&gt;
&lt;Height&gt;
&lt;MD5&gt;
</pre>

标签支持 fallback 链和 modifiers。完整参考：[docs/TAGS.md](docs/TAGS.md)。

## 安全性

执行前 EasyRenamer 会检查 Windows 非法名称、禁止字符、CON/NUL/COM1/LPT1 等保留名称、重复目标路径、已存在的目标文件、过长路径和已消失的源文件。

重命名通过唯一临时名称（如 <code>.easyrenamer_tmp_*</code>）分两阶段完成，因此交换名称和仅修改大小写的操作也能安全处理。

<strong>破坏性覆盖功能有意保持禁用</strong>，直到实现可靠的 backup/restore。

## 快捷键

| 快捷键 | 操作 |
| --- | --- |
| Ctrl + O | 添加文件 |
| Ctrl + Shift + O | 添加文件夹 |
| Ctrl + A | 选择行 |
| Ctrl + Z | 撤销最近可用操作 |
| Delete | 从当前任务中移除所选项 |
| Ctrl + Enter | 开始 |
| F5 | 重新计算 / 检查 |
| Esc | 取消当前操作 |

## 语言

应用界面支持 <strong>English、Русский、Español 和 中文</strong>。测试会检查翻译完整性，切换语言无需重启应用。

## 技术栈

Go 1.23+ · Wails v2 · React · TypeScript · Lucide · MIT License。

无需账号、云服务、分析系统或强制网络服务。

## 从源码构建

要求：Go 1.23+、Node.js 20+、Wails v2.15；最终 Windows x64 构建需要 Windows。

<pre>
git clone https://github.com/Solvo37/easyrenamer.git
cd easyrenamer
go test ./internal/...
go install github.com/wailsapp/wails/v2/cmd/wails@v2.15.0
wails build -clean -platform windows/amd64 -o EasyRenamer.exe
</pre>

## 下载

当前版本：[GitHub Releases](https://github.com/Solvo37/easyrenamer/releases/latest)

- EasyRenamer.exe — Windows x64 便携版；
- EasyRenamer-vX.Y.Z-windows-x64.zip — EXE + 文档；
- SHA256SUMS.txt — 校验和。

## 参与贡献

欢迎提交 bug、想法和 pull request。请阅读 [CONTRIBUTING.md](CONTRIBUTING.md)。

## 许可证

[MIT](LICENSE)。
