# Quickbuild (qb)
`qb` is a zero-configuration build system to very quickly build C/C++ projects on Linux, Windows, and MacOS.

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

## Installing
To install `qb`, you can [download a release](https://github.com/codecat/qb/releases) and make sure it exists in your `PATH`.

If you have [Go](https://golang.org/) installed, you can also install the latest development version of `qb` by simply running `go get github.com/codecat/qb`.

## Commands
You can pass a number of commands to `qb`.

### `qb run`
Runs the binary after building it.

### `qb clean`
Cleans all output files that qb could generate.

## Optional configuration
Since `qb` is meant to be a zero configuration tool, you don't have to do any configuration to get going quickly. It will do its best to find appropriate defaults for your setup, you just run `qb` and it builds.

If you **do** want a little bit more control over what happens, you can either use command line flags or create a configuration file in your source folder.

### Command line options
```
qb [--name name]
   [--type <exe|dll|lib>]
   [--pkg name]
   [--static]
   [--debug]
   [--verbose]
   [--strict]
   [--exceptions <std|all|min>]
   [--optimize <default|none|size|speed>]
   [--cppstd <latest|20|17|14>]
   [--cstd <latest|17|11>]
   [--include <path>]
   [--define <define>]
```

#### `--name`
Sets the name of the project and controls the output filename. You should not provide any file extension here as it will be added automatically.

If no name is passed, the name of the current directory will be used.

For example, `--name foo` will produce a binary `foo` on Linux, and `foo.exe` on Windows.

#### `--type`
Sets the type of the project, which can be an executable or a (dynamic) library. This is specified using the keywords `exe`, `dll`, or `lib`.

For example, to create a dynamic library, you would pass `--type dll`.

#### `--pkg`
Adds a package to link to by its name. `qb` will try to resolve the package by itself, using a variety of sources. Listed here are the sources, in the order that they will be searched for:

1. **Local configuration**: If you have a `qb.toml` file, this will check for packages defined there.
   ```toml
   [package.sfml]
   includes = [ "D:\\Libs\\SFML-2.5.1\\include\\" ]
   linkdirs = [ "D:\\Libs\\SFML-2.5.1\\lib\\" ]
   links = [
     "sfml-main.lib",
     "sfml-graphics-s.lib",
     "sfml-system-s.lib",
     "sfml-window-s.lib",
     "opengl32.lib",
     "winmm.lib",
     "gdi32.lib",
   ]
   defines = [ "SFML_STATIC" ]
   ```
2. **pkgconfig**: If you have `pkg-config` installed on your system, it will be checking for packages from there.
3. Nothing else yet, but the following is planned: global configuration (like local, but system-wide), and vcpkg (for Windows).

For example, to link with SFML, we can add `--pkg sfml`, as long as `sfml` can be resolved by one of the package sources.

Additionally, if [Conan](https://conan.io/) is installed, it may be used as a way to manage packages. If a `conanfile.txt` exists, it will run `conan install .` (unless `conanbuildfile.txt` already exists). Then `conanbuildfile.txt` is used to properly compile & link to any dependencies in the Conanfile.

#### `--static`
Links statically in order to create a standalone binary that does not perform any loading of dynamic libraries.

#### `--debug`
Produces debug information for the resulting binary. On Windows that means a `.pdb` file, on Linux that means embedding debug information into the binary itself so that it can be used with gdb, and on Mac that means a `.dSYM` bundle.

#### `--verbose`
Makes it so that all compiler and linker commands will be printed to the log. Useful for debugging `qb` itself.

#### `--strict`
Makes the compiler more strict with its warnings.

#### `--exceptions`
Sets the way that the compiler's runtime will handle exceptions. Can either be `standard` (`std`), `all`, or `minimal` (`min`). The default is `standard`.

This only makes a difference on Windows, where setting this to `all` will allow the runtime to catch certain access violation and other exceptions. When it's `minimal` or `min`, the minimal amount of exception handling will be done, which is similar to `all`, but there is no stack unwinding.

#### `--optimize`
Sets whether to use optimization. Can either be `default`, `none`, `size`, or `speed`. The default is `default`.

When this option is set to `default`, whether the binary will be optimized is defined by whether it's a debug build or not. For example, when building with `qb --debug`, you will get an unoptimized binary, but by building without any options (by just running `qb`) it will produce an optimized build.

#### `--cppstd`
Sets which C++ standard to use. Can either be `latest`, `20`, `17`, or `14`. The default is `latest`.

#### `--cstd`
Sets which C standard to use. Can either be `latest`, `17`, or `11`. The default is `latest`.

#### `--include`
Adds a directory to the include path. For example, to add the folders `foo` and `bar` to the include path, you would run `qb --include foo --include bar`.

#### `--define`
Adds a precompiler definition. For example, to define `FOO` and `BAR` in the preprocessor when compiling, you would run `qb --define FOO --define BAR`.

### Configuration file
It's possible to create a `qb.toml` file (in the folder you're running `qb`) to specify your configuration options as well. This is handy if you build a lot but don't want to pass the command line options every time.

The configuration file works the same as the command line options, except they are in a toml file. Values not defined in the file will remain as their defaults. For example, to make `qb` always build as a dynamic library with the name `libfoo`, you would put this in `qb.toml`:

```toml
name = "libfoo"
type = "dll"
```

To make a statically linked debug binary, you can put this in the configuration file:

```toml
static = true
debug = true
```
