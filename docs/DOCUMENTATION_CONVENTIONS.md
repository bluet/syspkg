# Documentation Conventions

This document defines the conventions for writing and maintaining documentation in the go-syspkg project.

## Anti-Pattern Marking Convention

To make anti-patterns easily discoverable via search, we use consistent markers in code examples and documentation.

### Visual Markers
- ✅ `// Good` or `// ✅ CORRECT:` for recommended patterns
- ❌ `// Bad` or `// ❌ WRONG:` for anti-patterns

### Searchable End-of-Line Comments

Add searchable keywords at the end of problematic lines to make them discoverable via `rg` or `grep`:

```go
// ❌ WRONG: Example anti-pattern
func badExample() {
    timeout := 5 * time.Minute  // WRONG: arbitrary timeout
    ctx = context.Background()  // BAD: ignoring context
    mgr.Install(packages)       // WRONG: missing context
}

// ✅ CORRECT: Recommended pattern  
func goodExample(ctx context.Context) {
    mgr.Install(ctx, packages, opts)
}
```

### Searchable Keywords

Use these consistent keywords for different types of issues:

- `// WRONG:` - Incorrect implementation patterns
- `// BAD:` - Poor practices or design decisions  
- `// DANGEROUS:` - Security or safety issues
- `// SLOW:` - Performance anti-patterns
- `// BROKEN:` - Code that doesn't work

### Benefits

1. **Discoverable**: `rg "WRONG:"` finds all anti-patterns instantly
2. **Consistent**: Same pattern across all documentation
3. **Educational**: Clear visual distinction between good/bad examples
4. **Maintainable**: Easy to update when patterns change

### Search Commands

Find anti-patterns quickly:

```bash
# Find all wrong patterns
rg "WRONG:" docs/

# Find all bad practices  
rg "BAD:" docs/

# Find security issues
rg "DANGEROUS:" docs/

# Find all anti-pattern markers
rg "WRONG:|BAD:|DANGEROUS:|SLOW:|BROKEN:" docs/
```

## Documentation Structure

### Code Examples
- Always show both ❌ wrong and ✅ correct examples when possible
- Include brief comments explaining why something is wrong
- Add searchable end-of-line markers for specific issues

### Section Headers
- Use clear, descriptive headers
- Include context about when to use patterns
- Explain the "why" behind recommendations

### Cross-References
- Link related documentation sections
- Reference specific implementation examples
- Point to relevant tests or fixtures

## Maintenance

When adding new documentation:

1. **Include anti-patterns**: Show what NOT to do
2. **Add searchable markers**: Use consistent `WRONG:`, `BAD:` etc.
3. **Provide context**: Explain why patterns are problematic  
4. **Link examples**: Point to real code that demonstrates patterns
5. **Update this guide**: Add new conventions as they emerge

## Future Considerations

This convention should be applied to:
- All new documentation
- Code comments in critical areas
- Example code and tutorials
- Error messages and warnings
- API documentation

The goal is to make the codebase self-documenting and easily searchable for both correct patterns and anti-patterns.