# Quickbuild (qb)
`qb` is a zero-configuration build system to very quickly build C/C++ projects on Linux and Windows.

<p align="center">
  <img src="https://4o4.nl/qb.gif" />
</p>

## Example
Let's say you have a folder containing some source files:

```c++
// main.cpp
#include "test.h"
int main() {
  test();
  return 0;
}

// test.h
void test();

// test.cpp
#include <cstdio>
void test() {
  printf("Hello, world\n");
}
```

You run `qb` in this directory:

```
~/qbtest $ qb
22:30:40.363 | main.cpp
22:30:40.364 | test.cpp
22:30:40.456 | 👏 qbtest
22:30:40.456 | ⏳ compile 53.9738ms, link 39.1138ms
```

And you run the resulting binary:
```
~/qbtest $ ./qbtest
Hello, world
```

## Optional configuration
Since `qb` is meant to be a zero configuration tool, you don't have to do any configuration to get going quickly. It will do its best to find appropriate defaults for your setup, you just run `qb` and it builds.

If you **do** want a little bit more control over what happens, you can either use command line flags or create a configuration file in your source folder.

### Command line options
```
qb [--name name] [--type (exe|dll|lib)] [--static] [--debug]
```

#### `--name`
Sets the name of the project and controls the output filename. You should not provide any file extension here as it will be added automatically.

If no name is passed, the name of the current directory will be used.

For example, `--name foo` will produce a binary `foo` on Linux, and a binary `foo.exe` on Windows.

#### `--type`
Sets the type of the project, which can be an executable or a (dynamic) library. This is specified using the keywords `exe`, `dll`, or `lib`.

For example, to create a dynamic library, you would pass `--type dll`.

#### `--static`
Links statically in order to create a standalone binary that does not perform any loading of dynamic libraries.

#### `--debug`
Produces debug information for the resulting binary. On Windows that means a `.pdb` file, and on Linux that means embedding debug information into the binary itself so that it can be used with gdb.

### Configuration file
It's possible to create a `qb.toml` file (in the folder you're running `qb`) to specify your configuration options as well. This is handy if you build a lot but don't want to pass the command line options every time.

The configuration file works the same as the command line options, except they are in a toml file. Values not defined in the file will remain as their defaults. For example, to make `qb` always build as a dynamic library with the name `libfoo`, you would put this in `qb.toml`:

```toml
name = "libfoo"
type = "dll"
```
