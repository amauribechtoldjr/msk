## Refactoring

### 1. Tmp-file-then-rename (atomic write)

Repeated in: `Session.StoreSession`, `Session.Refresh`, `Config.Save`, `Store.SaveFile`

All do: write to `.tmp` -> rename to final path (with `os.Remove(tmp)` on rename error). `Store.SaveFile` is the most robust version (Sync file + Sync dir). `Config.Save` already has a TODO to reuse `SaveFile`. Extract a shared `atomicWrite(path string, data []byte, perm os.FileMode) error` utility.

### 2. File read + IsNotExist handling

Repeated in: `Session.LoadFile`, `Session.Refresh`, `Config.Load`, `Store.GetFile`

All do: `os.ReadFile(path)` -> if `os.IsNotExist(err)` return custom error, else return raw error. Extract a shared `readFileOrErr(path string, notFoundErr error) ([]byte, error)`.

