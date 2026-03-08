## Refactoring

### 1. Tmp-file-then-rename (atomic write)

Repeated in: `Session.StoreSession`, `Session.Refresh`, `Config.Save`, `Store.SaveFile`

All do: write to `.tmp` -> rename to final path (with `os.Remove(tmp)` on rename error). `Store.SaveFile` is the most robust version (Sync file + Sync dir). `Config.Save` already has a TODO to reuse `SaveFile`. Extract a shared `atomicWrite(path string, data []byte, perm os.FileMode) error` utility.

### 2. Session file not found to generic error

"session file not found" doens't help the user to know what to do, he needs to run "unset MSK_SESSION" to make it work again
