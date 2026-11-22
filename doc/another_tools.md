# Another tools

## Go module cache

```bash
go clean -modcache
```

## Docker

```bash
docker system prune -a
```

## VS Code extensions

Check sizes and large extensions:

```bash
du -sh ~/.vscode/extensions/* | sort -rh | head -20
code --list-extensions --show-versions
```

Remove unused extensions via CLI/UI (close VS Code before manual cleanup):

```bash
code --uninstall-extension <publisher.extension>
```

After updates, `~/.vscode/extensions` may keep old versions of the same extension in separate folders. Check for duplicates and remove only outdated directories, keeping the current version:

```bash
# example: remove an old version after verifying the current one
rm -rf ~/.vscode/extensions/<publisher.extension-old-version>
```

After cleanup, restart VS Code; if issues occur, reinstall the extension.

## Cache

```bash
rm -rf ~/.cache/*
```
