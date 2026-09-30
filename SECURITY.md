# Security

<p align="center"><strong>English</strong> · <a href="SECURITY.ru.md">Русский</a> · <a href="SECURITY.es.md">Español</a> · <a href="SECURITY.zh-CN.md">中文</a></p>

EasyRenamer performs destructive filesystem operations by design, so file-safety bugs are treated seriously.

## Reporting

For ordinary bugs, use GitHub Issues.

If a problem can unexpectedly overwrite, lose, corrupt, or expose user files, describe the issue with disposable test data and avoid publishing private files, credentials, personal paths, or other sensitive information.

## Safety principles

- destructive overwrite is disabled until it can be implemented with safe backup/restore;
- rename plans are validated before execution;
- rename operations use temporary unique names to avoid internal collisions;
- completed operations are journaled for Undo when possible.

No security update should weaken these guarantees merely to make an operation succeed.
