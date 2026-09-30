<p align="center"><img src="frontend/public/easyrenamer.svg" width="96" height="96" alt="EasyRenamer"></p>
<h1 align="center">EasyRenamer</h1>

<p align="center"><a href="README.md">English</a> · <a href="README.ru.md">Русский</a> · <strong>Español</strong> · <a href="README.zh-CN.md">中文</a></p>

<p align="center">Renombrado masivo de archivos rápido y seguro para Windows.<br><strong>Gratis · Código abierto · Sin anuncios · Sin suscripción · Sin telemetría</strong></p>

<p align="center">
  <a href="https://github.com/Solvo37/easyrenamer/actions/workflows/ci.yml"><img alt="CI" src="https://github.com/Solvo37/easyrenamer/actions/workflows/ci.yml/badge.svg"></a>
  <a href="https://github.com/Solvo37/easyrenamer/releases/latest"><img alt="Release" src="https://img.shields.io/github/v/release/Solvo37/easyrenamer?display_name=tag"></a>
  <a href="LICENSE"><img alt="License" src="https://img.shields.io/github/license/Solvo37/easyrenamer"></a>
  <img alt="Windows x64" src="https://img.shields.io/badge/Windows-x64-0078D4?logo=windows11&logoColor=white">
</p>

<p align="center">
  <a href="https://github.com/Solvo37/easyrenamer/releases/latest/download/EasyRenamer.exe"><strong>Descargar EasyRenamer.exe</strong></a>
  · <a href="docs/LEARN.md">Guía</a>
  · <a href="docs/TAGS.md">Referencia de etiquetas</a>
  · <a href="docs/ROADMAP.md">Roadmap</a>
</p>

---

## Por qué EasyRenamer

EasyRenamer está pensado para trabajo real de renombrado por lotes: desde unos pocos archivos hasta grandes conjuntos distribuidos entre varias carpetas. Los nombres finales se muestran <strong>antes de modificar nada</strong>, el orden de procesamiento es explícito, los conflictos se comprueban con antelación y las operaciones son transaccionales y reversibles.

<strong>Añadir archivos → configurar métodos → revisar el resultado → ejecutar</strong>

## Funciones principales

| Función | Qué aporta |
| --- | --- |
| Vista previa en vivo | Recalcula automáticamente los nombres al cambiar plantillas, métodos, orden o filtros |
| Agrupación por carpetas | Límites claros entre carpetas y numeración independiente por carpeta |
| Orden natural | 1, 2, 10 en lugar de 1, 10, 2 |
| Orden de procesamiento | Nombre, fechas, tamaño, extensión, ruta, orden de añadido o drag & drop manual |
| Cadena de métodos | Activar, desactivar, duplicar y reordenar métodos |
| Búsqueda y errores | Buscar nombre original, nombre nuevo y ruta completa |
| Colisiones | Omitir conflictos, numerarlos automáticamente o detener la operación |
| Progreso + Cancelar | Las operaciones grandes siguen siendo responsivas y se pueden detener |
| Undo + Historial | Deshacer y conservar las últimas 20 operaciones |
| Listas grandes | Tabla virtualizada y carga diferida de datos pesados |
| 4 idiomas | English, Русский, Español, 中文 |
| Portable | Un solo EXE, sin instalador obligatorio |

## Ejemplo

Plantilla: <code>&lt;Inc NrDir:01&gt;_&lt;Name&gt;</code>

<pre>
1.jpg   → 01_1.jpg
2.jpg   → 02_2.jpg
10.jpg  → 03_10.jpg
</pre>

El contador <code>NrDir</code> vuelve a <code>01</code> en cada carpeta.

## Métodos de renombrado

EasyRenamer permite encadenar métodos:

- <strong>New Name</strong> — nombre a partir de texto y etiquetas;
- <strong>Replace</strong> — buscar y sustituir, incluido regex;
- <strong>Renumber</strong> — numeración secuencial, también por carpeta;
- <strong>Add text</strong> — prefijo y sufijo;
- <strong>Change case</strong> — lower / UPPER / Title Case;
- <strong>Remove / Remove pattern</strong> — eliminar por posición, texto o regex;
- <strong>Move / Swap</strong> — reorganizar partes del nombre;
- <strong>Trim</strong> — limpiar espacios;
- <strong>Timestamp</strong> — añadir fecha/hora;
- <strong>List / List Replace</strong> — nombres y reglas línea por línea;
- <strong>Script</strong> — expresiones seguras sin acceso directo a red o sistema de archivos.

Guía detallada: [docs/LEARN.md](docs/LEARN.md).

## Etiquetas y metadatos

Las plantillas admiten contadores, carpetas, fechas, tamaños, valores aleatorios, checksums y metadatos comunes:

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

Las etiquetas soportan fallback y modificadores. Referencia completa: [docs/TAGS.md](docs/TAGS.md).

## Seguridad

Antes de ejecutar, EasyRenamer comprueba nombres inválidos de Windows, caracteres prohibidos, nombres reservados como CON/NUL/COM1/LPT1, destinos duplicados, archivos de destino ya existentes, rutas demasiado largas y archivos de origen ausentes.

El renombrado se realiza en dos fases mediante nombres temporales únicos como <code>.easyrenamer_tmp_*</code>, por lo que intercambios y cambios solo de mayúsculas/minúsculas son seguros.

<strong>La sobrescritura destructiva está desactivada intencionadamente</strong> hasta disponer de backup/restore seguro.

## Atajos

| Atajo | Acción |
| --- | --- |
| Ctrl + O | Añadir archivos |
| Ctrl + Shift + O | Añadir carpetas |
| Ctrl + A | Seleccionar filas |
| Ctrl + Z | Deshacer la última operación disponible |
| Delete | Quitar elementos de la tarea |
| Ctrl + Enter | Ejecutar |
| F5 | Recalcular / comprobar |
| Esc | Cancelar la operación |

## Idiomas

La interfaz admite <strong>English, Русский, Español y 中文</strong>. Los tests comprueban la integridad de las traducciones y cambiar de idioma no requiere reiniciar.

## Tecnología

Go 1.23+ · Wails v2 · React · TypeScript · Lucide · Licencia MIT.

Sin cuentas, nube, analítica ni servicios de red obligatorios.

## Compilar desde el código fuente

Requisitos: Go 1.23+, Node.js 20+, Wails v2.15 y Windows x64 para el build final.

<pre>
git clone https://github.com/Solvo37/easyrenamer.git
cd easyrenamer
go test ./internal/...
go install github.com/wailsapp/wails/v2/cmd/wails@v2.15.0
wails build -clean -platform windows/amd64 -o EasyRenamer.exe
</pre>

## Descargas

Versiones actuales: [GitHub Releases](https://github.com/Solvo37/easyrenamer/releases/latest)

- EasyRenamer.exe — Windows x64 portable;
- EasyRenamer-vX.Y.Z-windows-x64.zip — EXE + documentación;
- SHA256SUMS.txt — checksums.

## Contribuir

Se aceptan bugs, ideas y pull requests. Consulta [CONTRIBUTING.md](CONTRIBUTING.md).

## Licencia

[MIT](LICENSE).
