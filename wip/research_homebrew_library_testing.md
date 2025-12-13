# Homebrew Library Formula Testing Patterns

Research on how Homebrew handles library-only formulas (packages with no executables).

## Core Testing Pattern

For library formulas without executables, Homebrew's recommended pattern is to **write and execute a small test program that compiles and links against the library**.

```ruby
test do
  # Write a small C/C++ source file to the temporary test directory
  (testpath/"test.cpp").write <<~CPP
    #include "library-header.h"
    // Test code that uses the library
  CPP

  # Compile and link against the installed library
  system ENV.cxx, "-std=c++11", "test.cpp",
    "-I#{include}", "-L#{lib}", "-llibrary-name", "-o", "test"

  # Run the compiled program
  system "./test"

  # Validate output/files if needed
  assert_match "expected output", (testpath/"output.txt").read
end
```

## Real Formula Examples

### tinyxml2 (Canonical Example)

Homebrew documentation specifically cites tinyxml2 as the canonical example for library testing.

### spdlog (Complete Example)

```ruby
test do
  (testpath/"test.cpp").write <<~CPP
    #include "spdlog/sinks/basic_file_sink.h"
    #include <iostream>
    #include <memory>
    int main()
    {
      try {
        auto console = spdlog::basic_logger_mt("basic_logger",
          "#{testpath}/basic-log.txt");
        console->info("Test");
      }
      catch (const spdlog::spdlog_ex &ex)
      {
        std::cout << "Log init failed: " << ex.what() << std::endl;
        return 1;
      }
    }
  CPP

  system ENV.cxx, "-std=c++11", "test.cpp", "-I#{include}",
    "-L#{Formula["fmt"].opt_lib}", "-lfmt", "-o", "test"
  system "./test"
  assert_path_exists testpath/"basic-log.txt"
  assert_match "Test", (testpath/"basic-log.txt").read
end
```

**Key patterns:**
- Tests actual library functionality (logging to file)
- Handles dependent libraries (fmt) correctly
- Validates output files created by the library
- Includes exception handling validation

### libyaml (Simple C Library)

```ruby
test do
  (testpath/"test.c").write <<~C
    #include <yaml.h>
    #include <stdio.h>

    int main() {
      yaml_parser_t parser;
      if (yaml_parser_initialize(&parser)) {
        printf("SUCCESS: libyaml initialized\\n");
        yaml_parser_delete(&parser);
        return 0;
      }
      return 1;
    }
  C

  system ENV.cc, "test.c", "-L#{lib}", "-I#{include}", "-lyaml", "-o", "test"
  system "./test"
end
```

### OpenSSL (Configuration-Heavy Library)

```ruby
test do
  # Verify required configuration files exist
  assert_path_exists etc/"openssl@3/openssl.cnf"

  # Test core cryptographic operations
  system bin/"openssl", "dgst", "-sha256", "-out", "checksum.txt", "testfile.txt"

  # Test certificate validation (expects failure for bad cert)
  output = pipe_output("#{bin}/openssl verify 2>&1", bad_cert, 2)
  assert_match "verification failed", output
end
```

## How `brew test` Works

### Execution Environment

- `brew test <formula>` invokes the `test do` block
- Automatically creates a temporary directory for test isolation
- Tests run with `HOME` set to the temporary test directory
- Directory is automatically deleted after the test completes
- Access the test directory via `testpath` function

### Test Fixtures

```ruby
# Access pre-built test files
test_fixtures("test.svg")
```

### Assertions Available

```ruby
assert_equal expected, actual
assert_match pattern, string
assert_predicate path, :exist?
assert_path_exists path
```

## keg_only Libraries

### What keg_only Means

```ruby
class LibraryFormula < Formula
  keg_only :provided_by_macos  # For system-provided software
  # OR
  keg_only "reason why not linked system-wide"
end
```

- Formula installed only to Cellar (not symlinked to brew prefix)
- Libraries not in system PATH
- Prevents shadowing system-provided software
- Other Homebrew formulas can depend on it directly

### keg_only Examples

- **OpenSSL** - avoid conflicting with system OpenSSL
- **LibFFI** - portable foreign function interface
- **Libyaml** - YAML parsing (Ruby dependency)

### Building with keg_only Dependencies

Homebrew's `superenv` automatically:
- Injects headers: `-I#{Formula["library"].opt_include}`
- Injects libraries: `-L#{Formula["library"].opt_lib}`
- Sets pkg-config: `PKG_CONFIG_PATH=/usr/local/opt/library/lib/pkgconfig`

## Testing Best Practices

### Good Tests

- Require no user input
- Test basic functionality thoroughly
- Use assertions for validation
- Test meaningful operations

### Bad Tests (Avoid)

- `foo --version` (too trivial for libraries)
- `foo --help` (too trivial)
- Tests requiring user interaction

## Implications for tsuku Library Recipes

For library recipes in tsuku, the `[verify]` section should:

1. **File-based verification** (minimal):
```toml
[verify]
command = "test -f lib/libyaml.a && test -f include/yaml.h"
```

2. **Symbol verification** (better):
```toml
[verify]
command = "nm lib/libyaml.a | grep -q yaml_parser_initialize"
```

3. **Compile test** (best, requires gcc/clang):
```toml
[verify]
command = "echo '#include <yaml.h>\nint main(){yaml_parser_t p;yaml_parser_initialize(&p);return 0;}' | gcc -x c - -L$TSUKU_HOME/lib -I$TSUKU_HOME/include -lyaml -o /tmp/test && /tmp/test"
```

## Sources

- [Homebrew Formula Cookbook](https://docs.brew.sh/Formula-Cookbook)
- [How to Build Software with Homebrew keg_only Dependencies](https://docs.brew.sh/How-to-Build-Software-Outside-Homebrew-with-Homebrew-keg-only-Dependencies)
- [tinyxml2 Formula](https://github.com/Homebrew/homebrew-core/blob/HEAD/Formula/t/tinyxml2.rb)
- [spdlog Formula](https://github.com/Homebrew/homebrew-core/blob/HEAD/Formula/s/spdlog.rb)
