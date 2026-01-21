# Testing Your Improvements

## Memory Safety Testing

Use Process Hacker to verify memory cleanup:

1. Run your MSK program
2. Perform operations (add/get secrets)
3. Use Process Hacker to search memory for test passwords
4. Verify sensitive data is NOT found after operations complete

## Security Testing Checklist

- [ ] Master key not found in memory after command completion
- [ ] Passwords not found in memory after encryption
- [ ] Derived keys cleared after use
- [ ] Input validation prevents path traversal
- [ ] File permissions are correct (0o600 for files, 0o700 for directories)
- [ ] Argon2 parameters are strong enough
