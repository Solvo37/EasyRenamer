# EasyRenamer — Guía de usuario

<p align="center"><a href="LEARN.md">English</a> · <a href="LEARN.ru.md">Русский</a> · <strong>Español</strong> · <a href="LEARN.zh-CN.md">中文</a></p>

Guía práctica para EasyRenamer v1.x.

## Inicio rápido

1. Añade archivos con **+ Archivos**, carpetas con **+ Carpetas** o arrástralos desde el Explorador.
2. Activa o desactiva **Incluir subcarpetas** y el filtro de extensiones.
3. Crea la cadena de métodos a la izquierda.
4. Elige el orden de procesamiento.
5. Revisa **Nuevo nombre** y los errores.
6. Pulsa **Iniciar**.
7. Usa **Ctrl+Z** o **Historial** para revertir una operación.

La vista previa se recalcula automáticamente. **Comprobar** fuerza una validación nueva.

## Archivos y carpetas

Puedes trabajar con archivos de varias carpetas en una sola tarea. Con la agrupación activa, cada carpeta aparece como un grupo independiente que se puede contraer, seleccionar, deseleccionar, quitar de la tarea, abrir en el Explorador o copiar su ruta.

Quitar un elemento de la lista **no lo elimina del disco**.

## Orden de procesamiento

El orden afecta a la numeración y al método List. Puedes ordenar por nombre natural, fecha de creación, fecha de modificación, tamaño, extensión, ruta, orden de añadido o manualmente.

El orden natural produce `1, 2, 10`, no `1, 10, 2`.

**Ordenar por separado dentro de cada carpeta** mantiene los grupos y ordena su contenido de forma independiente.

### Orden manual

Selecciona **Orden manual** y arrastra las filas. La vista previa y la numeración se recalculan.

## Numeración por carpeta

`<Inc NrDir:01>` reinicia el contador en cada directorio.

```text
Carpeta A:
1.jpg → 01_1.jpg
2.jpg → 02_2.jpg

Carpeta B:
a.jpg → 01_a.jpg
b.jpg → 02_b.jpg
```

## Cadena de métodos

Los métodos se aplican **de arriba abajo**. El resultado de uno se convierte en la entrada del siguiente.

Puedes activar/desactivar, editar, mover, duplicar, eliminar y guardar métodos en conjuntos.

## Métodos

- **New Name** — construye el nombre con texto y etiquetas; escribe `<` para abrir el autocompletado.
- **Replace** — buscar y reemplazar texto o regex.
- **Renumber** — añade una secuencia como prefijo/sufijo, con Start, Step, Padding y reinicio por carpeta.
- **Add text** — añade prefijo y/o sufijo.
- **Change case** — lower, UPPER o Title Case.
- **Remove** — elimina caracteres desde una posición.
- **Remove pattern** — elimina texto o coincidencias regex.
- **Move** — mueve una parte del nombre.
- **Swap** — intercambia dos partes alrededor de un separador.
- **Trim** — limpia espacios.
- **Timestamp** — añade al nombre la hora modificada del archivo o la hora del lote.
- **List** — asigna nombres línea por línea.
- **List Replace** — aplica varias reglas `buscar => reemplazar`.
- **Script** — expresión segura sin acceso directo a red o sistema de archivos.

Variables de Script: `Name`, `Ext`, `FullName`, `Index`, `DirIndex`, `DirName`, `UnixTimestamp`, `ModifiedUnix`.

Funciones: `lower()`, `upper()`, `trim()`, `replace()`, `concat()`, `substr()`.

## Etiquetas

Referencia completa: [TAGS.es.md](TAGS.es.md).

Ejemplos: `<Name>`, `<Ext>`, `<FolderName:1>`, `<Inc Nr:001>`, `<Inc NrDir:01>`, `<Date Modified:yyyy-mm-dd>`, `<Width>`, `<Artist>`, `<MD5>`.

Las etiquetas admiten fallback y modificadores.

## Presets de usuario

En **New Name** puedes guardar la plantilla como preset. Se puede aplicar, renombrar, duplicar o eliminar, y se conserva entre ejecuciones.

## Búsqueda y tabla

La búsqueda filtra por nombre original, nombre nuevo y ruta completa. **Solo errores** limita la tabla a elementos problemáticos.

Las columnas se pueden ocultar, reordenar y redimensionar. Ctrl+Click, Shift+Click y Ctrl+A controlan la selección de filas; el checkbox controla si el archivo participa en la operación.

## Política de colisiones

Por defecto se omiten los archivos en conflicto. También puedes añadir un número automáticamente o detener la operación. La sobrescritura destructiva permanece desactivada hasta disponer de backup/restore seguro.

## Validación de nombres de Windows

Se comprueban caracteres prohibidos, nombres vacíos o solo con espacios, punto/espacio final, nombres reservados, destinos duplicados o existentes, longitud de ruta y archivos desaparecidos.

## Renombrado seguro

EasyRenamer usa nombres temporales únicos antes del nombre final, por lo que intercambios como `A.jpg ↔ B.jpg` funcionan con seguridad. Al cancelar, intenta restaurar los nombres originales sin dejar temporales.

## Undo e historial

Las operaciones correctas guardan pares reales `ruta antigua → ruta nueva`. El historial conserva hasta 20 operaciones.

## Vista previa del archivo seleccionado

Puede mostrar miniatura, tamaño, tipo, dimensiones de imagen, nombre original/nuevo y diff visual. Los datos pesados se cargan bajo demanda.

## Atajos

| Atajo | Acción |
| --- | --- |
| `Ctrl + O` | Añadir archivos |
| `Ctrl + Shift + O` | Añadir carpetas |
| `Ctrl + A` | Seleccionar filas |
| `Ctrl + Z` | Undo |
| `Delete` | Quitar de la tarea |
| `Ctrl + Enter` | Iniciar |
| `F5` | Recalcular / comprobar |
| `Esc` | Cancelar |

## Metadatos

Los lectores integrados cubren datos comunes de JPEG/TIFF EXIF/GPS, MP3 ID3, FLAC, MP4/MOV, PDF, Office Open XML, EPUB, EML y recursos de versión de EXE de Windows.

## Idiomas y tema

Idiomas: English, Русский, Español, 中文.  
Temas: System / Light / Dark.

## Más

Roadmap: [ROADMAP.es.md](ROADMAP.es.md)  
Código y releases: [GitHub](https://github.com/Solvo37/easyrenamer)
