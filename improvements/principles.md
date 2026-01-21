# Key Principles

## Security Principles

1. **Defense in Depth**: Multiple layers of security
2. **Fail Secure**: If something fails, fail in a secure way
3. **Minimize Attack Surface**: Don't store/transmit more than necessary
4. **Principle of Least Privilege**: Give minimum access needed
5. **Secure by Default**: Secure defaults, require opt-in for weaker options

## Go Design Patterns

1. **Error Handling**: Errors are values, handle them explicitly
2. **Interfaces**: Program to interfaces, not implementations
3. **Context**: Use for cancellation, timeouts, request-scoped values
4. **Composition**: Prefer composition over inheritance
5. **Package Design**: Clear boundaries, minimal exports
