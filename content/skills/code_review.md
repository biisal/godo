# Example Skill: Code Review 

When reviewing Go code, you should prioritize:
1. Error handling: Ensure errors are checked and returned or logged appropriately.
2. Naming conventions: Use camelCase for variables and PascalCase for exported identifiers.
3. Concurrency: Avoid simple goroutine leaks, ensure proper synchronization using waitgroups or channels.
4. Simplicity: Suggest removing over-engineered abstractions.

Remember this skill when asked to review or write new Go code!
