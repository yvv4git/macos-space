# MACOS-space

CLI utility for macOS disk space analysis.

## Analyse

`analyse` shows disk usage for configured directories or a specific path.

```bash
macos-space analyse
macos-space analyse -p ~/Downloads
macos-space analyse -p ~/Downloads -r -d 1 -f
macos-space analyse -p ~/Downloads -r -d 1 -f -e ~/Downloads/tmp
```

Flags:

- `-c, --config` - path to config file.
- `-p, --path` - scan a specific directory instead of configured directories.
- `-r, --recursive` - show directory tree recursively.
- `-d, --depth` - limit recursive depth, `0` means unlimited.
- `-f, --files` - include files in output.
- `-e, --exclude` - exclude directory/file path; can be repeated. CLI value overrides config `analyse.exclude`.

### Scan modes

- Without `-r`, `analyse` shows the immediate contents of the scanned directory. Files are shown only with `-f`. To report a real size for each child directory, `analyse` walks that child directory tree and sums file sizes, similar to `du`. Directory metadata size is not enough on macOS/APFS. Progress shows top-level entries being calculated; excluded paths are skipped.
- With `-r`, `analyse` walks the directory tree and prints directories (and files with `-f`) as a tree. Directory size is the total size of its subtree. `-d` limits traversal depth.

### Config

```toml
[analyse]
directories = [
  "/Users/name/Downloads",
  "/Users/name/Library/Caches",
]
exclude = [
  "/Users/name/Downloads/tmp",
]
directoryTimeout = "60s"
depth = 0
```

`directoryTimeout` limits time spent calculating one directory size. `exclude` contains paths skipped during scan.

## Дополнительные места

Список системных и временных каталогов вне дома: [doc/system_places.md](doc/system_places.md).
