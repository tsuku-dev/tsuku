# Library Validation on Linux

Research on validating portable C libraries (libyaml, libffi, openssl, zlib) in Linux environments.

## Tools Summary

| Tool | Purpose | Common Flags |
|------|---------|--------------|
| `file` | Identify file type | (none) |
| `ar` | Archive tool (static libs) | `-t` (list contents) |
| `nm` | Display symbols | `-u` (undefined), `-g` (global) |
| `readelf` | ELF format inspection | `-h` (header), `-S` (sections), `-d` (dynamic) |
| `objdump` | Object file tool | `-p` (private headers), `-h` (section headers) |
| `ldd` | Show shared library deps | `-v` (verbose) - WARNING: may execute code |

## Validating Static Libraries (.a files)

```bash
# Verify it's actually a static library archive
file libmylib.a
# Expected: current ar archive

# List object files in the archive
ar -t libmylib.a

# List all symbols
nm libmylib.a

# Count total symbols (health check)
nm libmylib.a | wc -l

# Check for undefined symbols (dependencies)
nm -u libmylib.a

# Verify symbol table index exists
nm -s libmylib.a | head -5
```

## Validating Shared Libraries (.so files)

```bash
# Verify it's a valid shared library
file libmylib.so
# Expected: ELF 64-bit LSB shared object, x86-64, dynamically linked

# Show ELF header information
readelf -h libmylib.so

# List required shared libraries (SAFE - no execution)
objdump -p libmylib.so | grep NEEDED

# Alternative using readelf
readelf -d libmylib.so | grep NEEDED

# List exported symbols
nm libmylib.so | grep " T " | head -20

# Check SONAME
readelf -d libmylib.so | grep SONAME
```

## Comprehensive Validation Script

```bash
#!/bin/bash
validate_library() {
    local lib=$1
    local errors=0

    echo "Validating: $lib"

    # 1. File checks
    [[ -f "$lib" ]] || { echo "ERROR: File not found"; return 1; }
    [[ -r "$lib" ]] || { echo "ERROR: File not readable"; return 1; }

    # 2. File type check
    if ! file "$lib" | grep -qE "(archive|shared object)"; then
        echo "ERROR: Invalid file type"
        ((errors++))
    fi

    # 3. ELF format validation
    if ! readelf -h "$lib" > /dev/null 2>&1; then
        echo "ERROR: Invalid ELF format"
        ((errors++))
    fi

    # 4. Architecture check
    local arch=$(readelf -h "$lib" 2>/dev/null | grep "Machine:" | awk '{print $NF}')
    echo "  Architecture: $arch"

    # 5. Symbol table check
    if [[ "$lib" == *.a ]]; then
        local sym_count=$(nm "$lib" 2>/dev/null | wc -l)
        echo "  Symbols: $sym_count"
        [[ $sym_count -gt 0 ]] || { echo "WARNING: No symbols found"; ((errors++)); }
    else
        local sym_count=$(readelf -s "$lib" 2>/dev/null | grep -c "FUNC\|OBJECT")
        echo "  Exported symbols: $sym_count"
    fi

    # 6. Section integrity
    if ! readelf -S "$lib" > /dev/null 2>&1; then
        echo "ERROR: Cannot read section headers"
        ((errors++))
    fi

    return $errors
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

# Compile against shared library
gcc /tmp/test_lib.c -L/path/to/lib -I/path/to/include -lyaml -o /tmp/test_lib

# Run with library path
LD_LIBRARY_PATH=/path/to/lib /tmp/test_lib

# Compile against static library
gcc /tmp/test_lib.c /path/to/libyaml.a -I/path/to/include -o /tmp/test_lib_static
./tmp/test_lib_static
```

## Library-Specific Examples

### libyaml
```bash
nm /path/to/libyaml.so | grep "yaml_parser_initialize"
nm /path/to/libyaml.so | grep "yaml_emitter_initialize"
```

### openssl
```bash
nm libssl.so | grep "SSL_new"
nm libcrypto.so | grep "EVP_"
objdump -p libssl.so | grep NEEDED  # Check zlib dependency
```

### libffi
```bash
nm libffi.so | grep "ffi_call"
nm libffi.so | grep "ffi_prep_cif"
```

### zlib
```bash
nm libz.so | grep -E "compress|uncompress|inflate|deflate"
```

## Package Manager Validation

### RPM (Red Hat/Fedora)
```bash
rpm -V package-name           # Verify package integrity
rpm -qR package-name          # Check dependencies
```

### DEB (Debian/Ubuntu)
```bash
dpkg --verify package-name    # Verify integrity (MD5)
dpkg -L package-name          # List package files
apt-cache depends package-name # Check dependencies
```
