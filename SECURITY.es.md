# Seguridad

<p align="center"><a href="SECURITY.md">English</a> · <a href="SECURITY.ru.md">Русский</a> · <strong>Español</strong> · <a href="SECURITY.zh-CN.md">中文</a></p>

EasyRenamer modifica el sistema de archivos, por lo que los errores capaces de dañar datos del usuario se consideran serios.

## Informar de un problema

Los bugs normales pueden reportarse mediante GitHub Issues.

Si un problema puede sobrescribir, perder, corromper o exponer archivos inesperadamente, usa datos de prueba desechables y no publiques archivos privados, credenciales o rutas personales.

## Principios de seguridad

- overwrite destructivo desactivado sin backup/restore seguro;
- validación del plan antes de ejecutar;
- nombres temporales únicos para evitar colisiones internas;
- operaciones correctas registradas para Undo cuando sea posible.

Una corrección no debe debilitar estas garantías solo para que una operación termine.
