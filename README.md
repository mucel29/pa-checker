# PA CHECKER

## License

This project is using the `LICENSE PENDING` License.

## Contributors

* Asavoae Cosmin-Stefan
* Gatej Stefan-Alexandru
* Neamu Ciprian-Valentin
* Potop Horia-Ioan

## Dependencies

* `valgrind`
* `cppcheck`
* `git`

## Features

- [x] Parallel test running

- [x] Configuration
  - [x] Configurable tests
  - [x] Configurable modules
  - [x] User configuration
  - [x] Macros

- [x] Modules
  - [x] Module dependency checks
  - [x] Diff module
  - [x] Memory module _(valgrind backend)_
  - [x] Style module _(cppcheck backend)_
  - [x] Commit module _(git backend)_

- [x] Interface
  - [x] Basic - full module dump
  - [x] Interactive
    - [x] Live reload
    - [x] Module output visualization
      - [x] Side-by-side diff visualization
      - [x] Memory leak information

  
- [x] OS Compatibility
  - [x] `Linux / WSL` - full support
  - [x] `OSX` - full support _(at least from the tests)_
  - [ ] `Windows` - partial support _(no backend for the memory module)_

## Overview

### Running the checker

#### Basic
```bash
./checker
```

#### Interactive
```bash
./checker -i
```

### Navigating the interactive interface

* Use the `arrow keys` to navigate around
* Press `TAB` to switch between navigation and current section
* Press `ESC` to exit a fullscreen page
* Press `ESC` or `Ctrl+C` while on the main page to exit the program
* Press `~` to trigger a test run _(or modify the executable)_
* `Mouse` should be fully supported

### Configuration

Inside `config.json` or the `Options` tab you can modify the following options:

* `Executable Path` - the executable that will be used to run the tests
* `Source Path` - the project root directory
* `Input Path` - the directory containing the input files
* `Output Path` - the directory where the test output will be stored
* `Ref Path` - the directory containing the reference files
* `Forward Path` - the directory where the `stdout` & `stderr` of each test will be stored
* `Valgrind` - whether to run the tests using valgrind or not _(disable for faster iteration)_
* `Tutorial` - display the tutorial again _(disabled afterward)_

## Contributing

### Project Structure

* root
  * `bin` - Windows & Linux compiled binaries 
  * `res` - project resources: config files
  * `src` - project source files
    * `checker-modules`
    * `display`
    * `menu`
    * `manager`
    * `utils`
  * `main.go` - project entrypoint
  * `Makefile` - use this to compile the project

### Building the checker

To build the checker simply run
```bash
make build-linux # ELF executable

make build-windows # Win32 executable

make build-macos # OSX executable
```

### Formatting

Before committing any changes, run
```bash
make vet

make lint
```

### Commit format

* Keep in mind that Andra recommended that the commits be in english.
* The commits must be signed (`git commit -s`)
* The commit messages should have the following structure

```
MODULE: <concise title>

*detailed description* (around 75 characters per line)
```
> Example commit message
> ```
> ref: Added order checks
> 
> Lorem ipsum odor amet, consectetuer adipiscing elit. Neque magna platea
> ornare a maecenas aptent tincidunt. Tellus dolor maecenas congue pharetra
> leo himenaeos dis curabitur. Accumsan venenatis eget ipsum enim montes
> volutpat quisque. Diam finibus leo mattis fames efficitur.
> ```
