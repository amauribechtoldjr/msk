## Refactoring

### 1. Session file not found, too generic error

When running 'msk lock' without using "unset MSK_SESSION", app returns an error "session file not found" that doens't help the user to know what to do.

### 2. Add pipe to run tests when creating a pull request on github - DONE

### 3. Add pipeline to release on main

### 4. Refactor config.go to stop using secret.domain struct

### 5. Review silently usage for ALL commands

Also check promptings earlier, I want to move all inside cobra, do not print errors after cobra help anymore!!

### 6. Check config file before prompting master password

### 7. Improve list command to better results layout
