# Системные и временные каталоги

## Кандидаты вне дома

- `/Library/Caches` - кеши системных и сторонних приложений.
- `/Library/Logs` - журналы приложений и служб.
- `/Library/Updates` - загруженные обновления macOS.
- `/private/tmp` - временные файлы, очищаются при перезагрузке.
- `/private/var/tmp` - временные файлы между перезагрузками.
- `/private/var/folders` - временные данные пользовательских сессий.
- `/private/var/log` - журналы системы.
- `/private/var/db/diagnostics` - хранилище unified logging.
- `/Users/Shared` - общие данные пользователей.

## Критерии исключения

- Пути под SIP не добавляются в конфиг: `/System`, `/usr`, `/bin`, `/sbin`.
- Пути, доступные только root, не добавляются в конфиг.
- Внешние монтирования в `/Volumes` не добавляются в конфиг по умолчанию.
- Пути с ошибкой чтения не добавляются в конфиг.

## Проверка размера

```bash
du -sh /Library/Logs /Library/Updates /Users/Shared
du -sh /private/tmp /private/var/tmp /private/var/log /private/var/db/diagnostics
du -sh /Library/Caches/* 2>/dev/null | sort -rh | head -20
du -sh /private/var/folders/* 2>/dev/null | sort -rh | head -20
```

- `/Library/Caches` и `/private/var/folders` не идут в конфиг из-за ошибок чтения.
- Ошибки чтения скрыты в командах через `2>/dev/null`.

## Удаление

- Удалять что-либо только после проверки размера и назначения каталога.
- Не удалять вручную `/Library/Updates`, `/private/var/db/diagnostics` и `/private/var/log`.
- Для `/private/tmp` и `/private/var/tmp` предпочтительна перезагрузка.
- Команды удаления в этой документации не приводятся.
