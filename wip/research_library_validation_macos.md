# Library Validation on macOS

Research on validating portable C libraries (libyaml, libffi, openssl, zlib) in macOS environments.

## Tools Summary

| Tool | Purpose | Common Flags |
|------|---------|--------------|
| `file` | Identify file type | (none) |
| `otool` | Object file tool | `-L` (deps), `-D` (dylib ID), `-l` (load cmds) |
| `nm` | Display symbols | `-u` (undefined), `-gC` (global demangle) |
| `lipo` | Universal binary tool | `-info`, `-archs`, `-detailed_info` |
| `codesign` | Code signing | `-v` (verify), `-dv` (display verbose) |
| `install_name_tool` | Modify install names | `-change`, `-rpath`, `-id` |
| `ar` | Archive tool (static libs) | `-t` (list contents) |

## Validating Static Libraries (.a files)

```bash
# Identify file type
file libexample.a
# Output: libexample.a: current ar archive random library

# Check if universal (fat) binary
lipo -info libexample.a
# Output: Architectures in the fat file: libexample.a are: x86_64 arm64

# List object files in archive
ar -t libexample.a

# Display symbol table
nm -gC libexample.a  # -g = global, -C = demangle C++

# Check for undefined symbols
nm -u libexample.a

# Get detailed architecture info
lipo -detailed_info libexample.a
```

## Validating Dynamic Libraries (.dylib files)

```bash
# Check file type
file libexample.dylib
# Output: Mach-O 64-bit dynamically linked shared library arm64

# Get the library's install name (ID_DYLIB)
otool -D libexample.dylib

# List all dependencies
otool -L libexample.dylib

# Check for @rpath references
otool -l libexample.dylib | grep -A3 "LC_RPATH"

# Show runtime search paths
otool -l libexample.dylib | grep -B2 -A5 "LC_RPATH"

# Check architectures
lipo -archs libexample.dylib
# Output: x86_64 arm64

# Verify code signature
codesign -v libexample.dylib
# Success: "valid on disk"

# Display signature details
codesign -dv --verbose=4 libexample.dylib
```

## Comprehensive Validation Script

```bash
#!/bin/bash
validate_c_library() {
    local lib_path="$1"

    # Input validation
    if [[ ! -f "$lib_path" ]]; then
        echo "ERROR: Library file not found: $lib_path"
        return 1
    fi

    local lib_name=$(basename "$lib_path")
    echo "Validating: $lib_name"
    echo "============================================"

    # File type detection
    local file_output=$(file "$lib_path")

    if echo "$file_output" | grep -q "ar archive"; then
        echo "Detected: Static Library (.a)"

        # Check archive integrity
        if ! ar -t "$lib_path" > /dev/null 2>&1; then
            echo "  ERROR: Archive is corrupted"
            return 1
        fi

        local obj_count=$(ar -t "$lib_path" | wc -l)
        echo "  Object files: $obj_count"

        local symbol_count=$(nm "$lib_path" 2>/dev/null | wc -l)
        echo "  Total symbols: $symbol_count"

    elif echo "$file_output" | grep -q "dynamically linked shared library"; then
        echo "Detected: Dynamic Library (.dylib)"

        # Get install name
        local install_name=$(otool -D "$lib_path" 2>/dev/null | tail -1)
        echo "  Install name: $install_name"

        # List dependencies
        echo "  Dependencies:"
        otool -L "$lib_path" | tail -n +2 | while read -r dep; do
            echo "    - $(echo "$dep" | awk '{print $1}')"
        done

        # Code signature
        if codesign -v "$lib_path" 2>&1 | grep -q "valid on disk"; then
            echo "  Code signature: Valid"
        else
            echo "  Code signature: Unsigned or invalid"
        fi
    fi

    # Architecture check
    echo ""
    echo "Architectures:"
    lipo -info "$lib_path" 2>/dev/null || echo "  (single architecture)"

    return 0
}
```

## Quick Reference Commands

```bash
# Universal quick check
quick_check_lib() {
    local lib="$1"
    echo "=== File Type ===" && file "$lib"
    echo "" && echo "=== Architectures ===" && lipo -info "$lib" 2>/dev/null
    echo "" && echo "=== Dependencies ===" && otool -L "$lib" 2>/dev/null | head -5
    echo "" && echo "=== Signature ===" && codesign -v "$lib" 2>&1 | head -1
}
```

## Minimal Compile Test (Best Validation)

```bash
# Create test program
cat > /tmp/test_lib.c << 'EOF'
#include <yaml.h>
int main() {
    yaml_parser_t parser;
    yaml_parser_initialize(&parser);
    yaml_parser_delete(&parser);
    return 0;
}
EOF

# Compile and link
clang -o /tmp/test_lib /tmp/test_lib.c \
    -I/path/to/include \
    -L/path/to/lib \
    -lyaml

# Run the test
/tmp/test_lib && echo "SUCCESS" || echo "FAILED"
```

## Runtime Debugging

```bash
# Show all libraries as they load
DYLD_PRINT_LIBRARIES=1 ./my_program

# Show runtime statistics
DYLD_PRINT_STATISTICS=1 ./my_program
```

## Homebrew Integration

```bash
# Check linkage after installing
brew linkage <formula_name>

# Show all libraries a keg uses
brew linkage <formula_name> --reverse

# Find undeclared dependencies
brew linkage <formula_name> --strict

# Only report missing libraries
brew linkage <formula_name> --test
```

## Architecture Considerations

For Apple Silicon (M1/M2/M3) compatibility, libraries should include both architectures:

```bash
# Check for universal binary
lipo -info libexample.dylib
# Should show: x86_64 arm64

# Extract specific architecture
lipo libexample.dylib -thin arm64 -output libexample_arm64.dylib

# Create universal binary from two architectures
lipo -create libexample_x86_64.dylib libexample_arm64.dylib -output libexample.dylib
```
