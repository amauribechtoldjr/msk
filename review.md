## SESSION.go

### Create(Vault)

- Generate token
- VAULT -> createSession(generated Token)
- Create binary session file
- Store binary file
- Return token

### Load(Token, Vault)

- Read file
- Unmarshall binary session file
- VAULT -> loadSession(token, nonce, cipherData)
- Return master key

### Refresh

- Read file
- Generate new binary with new expiry date
- Store new binary file
- Return nil

### Destroy

- Remove file
- Return nil

## CONFIG.go

### NewConfig(path)

- If path provided, use it directly
- Get OS user config dir
- Return Config with default path (configDir/msk/config.msk)

### Exists()

- Stat file
- Return bool

### Load(Vault)

- Read file
- VAULT -> DecryptSecret(data)
- Validate config name
- Return vault path

### Save(Vault, vaultPath)

- MkdirAll config dir
- Create domain.Secret with config name + vaultPath as password
- VAULT -> EncryptSecret(secret)
- Write tmp file
- Rename tmp -> final
- Return nil

### DefaultVaultPath()

- Get user home dir
- Return home/.msk/vault

## STORE.go

### NewStore(dir)

- MkdirAll dir
- Return Store

### SaveFile(encryptedFile, name)

- Build file path from name
- Write tmp file (OpenFile + Write + Sync + Close)
- Rename tmp -> final
- Sync parent dir
- Return nil

### GetFile(name)

- Read file
- Handle IsNotExist -> ErrNotFound
- Return data

### DeleteFile(name)

- Stat file
- Check not a dir
- Remove file
- Return nil

### FileExists(name)

- Stat file
- Return bool

### GetFiles()

- ReadDir
- Filter: skip dirs, keep only .msk files
- Return names

## VAULT.go

### ConfigMK(mk)

- Wrap mk bytes in memguard buffer
- Seal into enclave
- Store in struct

### DestroyMK()

- Purge all memguard memory
- Nil out mk reference

### EncryptSecret(secret)

- Generate random salt
- Open mk from enclave
- Derive secret key (mk + salt)
- Marshal secret to bytes
- sealGCM(key, bytes) -> nonce, cipherData
- MarshalFile(salt, nonce, cipherData)
- Return encrypted bytes

### DecryptSecret(cipherData)

- UnmarshalFile -> salt, nonce, data
- Open mk from enclave
- Derive secret key (mk + salt)
- openGCM(nonce, key, data) -> plaintext
- UnmarshalSecret(plaintext) -> secret
- Return secret

### CreateSession(token)

- Open mk from enclave
- Hash token -> session key
- sealGCM(key, mk) -> nonce, cipherData
- Return nonce, cipherData

### LoadSession(token, nonce, cipherData)

- Hash token -> session key
- openGCM(nonce, key, cipherData) -> mk
- Return mk

### sealGCM(key, bytes) [private]

- NewCipher(key) -> block
- NewGCM(block) -> gcm
- Generate random nonce
- gcm.Seal -> cipherData
- Return nonce, cipherData

### openGCM(nonce, key, data) [private]

- NewCipher(key) -> block
- NewGCM(block) -> gcm
- gcm.Open -> plaintext
- Return plaintext

## FORMAT.go

### MarshalSecret(secret)

- Allocate buffer (nameLen + name + passLen + password)
- Write name length (uint16 BE) + name bytes
- Write password length (uint16 BE) + password bytes
- Return buffer

### UnmarshalSecret(data)

- Read name length (uint16 BE)
- Read name bytes
- Read password length (uint16 BE)
- Read password bytes (copy)
- Validate bounds at each step -> ErrCorruptedFile
- Return Secret

### MarshalFile(salt, nonce, data)

- Validate salt size and nonce size
- Allocate buffer (header + data)
- Write magic ("MSK") + version + salt + nonce + data
- Return buffer

### UnmarshalFile(data)

- Validate min length (header size)
- Validate magic value
- Validate file version
- Extract salt, nonce, secret from fixed offsets
- Return salt, nonce, secret

### getBufferLength(secret) [private]

- Return nameLen(2) + name + passLen(2) + password

## ROOT.go (cli)

### NewMSKCmd(vault)

- PersistentPreRunE:
  - Skip if command is in ignored list (version, help, unlock, lock, config)
  - If MSK_SESSION env set:
    - session.New()
    - session.Load(token, vault) -> mk
  - Else:
    - PromptMasterPassword -> mk
  - vault.ConfigMK(mk)
  - config.NewConfig("") -> conf
  - conf.Load(vault) -> vaultPath
  - storage.NewStore(vaultPath) -> store
  - app.NewMSKService(store, vault) -> holder.Service

- PersistentPostRunE:
  - If holder.Service != nil -> DestroyMK()

- RunE:
  - If --version flag -> print version, exit

- Register subcommands: add, get, delete, list, update, config, version, unlock, lock

---

## Refactoring

### 1. Tmp-file-then-rename (atomic write)

Repeated in: `Session.Create`, `Session.Refresh`, `Config.Save`, `Store.SaveFile`

All do: write to `.tmp` -> rename to final path (with `os.Remove(tmp)` on rename error). `Store.SaveFile` is the most robust version (Sync file + Sync dir). Extract a shared `atomicWrite(path string, data []byte, perm os.FileMode) error` utility.

### 2. File read + IsNotExist handling

Repeated in: `Session.Load`, `Config.Load`, `Store.GetFile`

All do: `os.ReadFile(path)` -> if `os.IsNotExist(err)` return custom error, else return raw error. Extract a shared `readFileOrErr(path string, notFoundErr error) ([]byte, error)`.

### 3. File exists check (stat + bool)

Repeated in: `Config.Exists()`, `Store.FileExists()`

Both do: `os.Stat(path)` -> return true/false. Extract a shared `fileExists(path string) (bool, error)`.

### 4. MK enclave open pattern

Repeated in: `Vault.EncryptSecret`, `Vault.DecryptSecret`, `Vault.CreateSession`

All do: nil-check mk -> `mk.Open()` -> `defer lockedBuffer.Destroy()`. Extract a private `openMK() (*memguard.LockedBuffer, error)` method on MSKVault.

### 5. Secret key derivation setup

Repeated in: `Vault.EncryptSecret`, `Vault.DecryptSecret`

Both do: `SecretKeyDeriver{}` -> `getSecretKey(mk, salt)` -> `defer wipe.Bytes(key)`. Could be inlined into the `openMK` flow or extracted as `deriveKey(mk, salt) ([]byte, error)`.

### 6. Config dir resolution

Repeated in: `session.New()`, `config.NewConfig("")`

Both call `os.UserConfigDir()` and join under `"msk/"`. Extract a shared `mskConfigPath(filename string) (string, error)`.

### 7. Config.Load abuses DecryptSecret

`Config.Load` uses `Vault.DecryptSecret` but treats the result as a vault path (not a real secret). The `Password` field holds the vault path and `Name` holds a config identifier. Consider a dedicated config encrypt/decrypt or a more generic data encrypt/decrypt on the vault.
