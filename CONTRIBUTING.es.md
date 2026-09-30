# Contribuir a EasyRenamer

<p align="center"><a href="CONTRIBUTING.md">English</a> · <a href="CONTRIBUTING.ru.md">Русский</a> · <strong>Español</strong> · <a href="CONTRIBUTING.zh-CN.md">中文</a></p>

Se aceptan bugs, ideas y pull requests.

## Antes de empezar

1. Revisa Issues y Pull Requests existentes.
2. Para una función importante, describe primero el flujo de trabajo real.
3. No añadas telemetría, publicidad, cuentas ni dependencias de red obligatorias.

## Desarrollo local

Requisitos: Go 1.23+, Node.js 20+, Wails v2.15 y Windows para el build desktop final.

```powershell
go test ./internal/...
cd frontend
npm install
npm run build
go install github.com/wailsapp/wails/v2/cmd/wails@v2.15.0
wails build -clean -platform windows/amd64 -o EasyRenamer.exe
```

## Checklist del PR

- código formateado;
- tests de Go y build de TypeScript correctos;
- las nuevas cadenas UI existen en EN/RU/ES/ZH;
- las operaciones de archivos siguen pasando por preview/validation;
- no se activa overwrite destructivo sin rollback seguro;
- la documentación se actualiza si cambia el workflow;
- no se suben binarios ni build output.

## Arquitectura

La lógica crítica de rename, orden, validación y filesystem debe permanecer en el motor Go, no duplicada en UI.

## Seguridad

Los bugs que puedan sobrescribir, perder o corromper archivos deben reproducirse con datos desechables.

## Licencia

Las contribuciones se publican bajo [MIT](LICENSE).
