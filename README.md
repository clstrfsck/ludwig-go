# The Ludwig Editor

```text
{**********************************************************************}
{                                                                      }
{            L      U   U   DDDD   W      W  IIIII   GGGG              }
{            L      U   U   D   D   W    W     I    G                  }
{            L      U   U   D   D   W ww W     I    G   GG             }
{            L      U   U   D   D    W  W      I    G    G             }
{            LLLLL   UUU    DDDD     W  W    IIIII   GGGG              }
{                                                                      }
{**********************************************************************}
```

## About

Ludwig is a text editor developed at the University of Adelaide.
It is an interactive, screen-oriented text editor.
It may be used to create and modify computer programs, documents
or any other text which consists only of printable characters.

Ludwig may also be used on hardcopy terminals or non-interactively,
but it is primarily an interactive screen editor.

This is a Golang port of the Ludwig code. The original Pascal code is
available here: [cjbarter/ludwig](https://github.com/cjbarter/ludwig).
There is also a C++ port available here:
[clstrfsck/ludwig-c](https://github.com/clstrfsck/ludwig-c).

## Building

The easiest way to build is by using `task`.  The build is a multi-stage build
where the indexed helpfile is embedded into the output binary.  This means the
online help is always available.  The minor downside is that the build process
is slightly more complicated.

Using `task`:

```sh
# Build debug binary
task build
# Build release binary (no symbols)
task build-release
# Run unit tests
task test
# Build everything and run all available tests
task check
# See Taskfile.yml for more
```

The release build will produce a `ludwig` executable which can be copied to
your preferred directory for local binaries, eg `/usr/local/bin`.

If you would prefer to use a different help file than the embedded
documentation, you can set the environment variables `LUD_HELPFILE` and
`LUD_NEWHELPFILE` to point to the locations of the old and new command help
files respectively.

## Coverage

Unit test coverage is quite low right now.  This is being worked on as
refactoring and modernisation continues.

```sh
# Unit test coverage
task coverage
```

## System Tests

There is reasonable system test coverage.  The system tests leverage
Ludwig's batch mode, where a command string is provided on stdin.  The
general approach is:

- The test provides a selection of initial filenames and contents, together
  with expected output files and contents and a command string
- The test framework creates a temporary directory and populates it with the
  supplied files
- The command string is piped into a Ludwig process running in the temporary
  directory
- Once the process completes, the files in the temporary directory are
  collected and compared against expectations

You can clone the
[system tests](https://github.com/clstrfsck/ludwig-system-test) using:

```sh
git clone https://github.com/clstrfsck/ludwig-system-test system-test
# Assuming you have python, pytest and pexpect installed
./system-test/run-system-tests.sh
```

The intention is that the system tests are cloned into a subdirectory of
the main `ludwig-go` project.  If you would like to arrange things differently,
you can use the environment variable `LUDWIG_EXE` to point the tests to
your executable.  Note that this path will need to be an absolute path.

Once the tests are running, you should see a bunch of dots, followed by
something like:

```text
348 passed, 3 skipped in 3.12s
```

Two of the three skipped tests are cases where regular expression patterns
don't match candidate strings in the way I think they should.  The third is
a window related command that is not implemented nor appropriate for batch
mode.  I've left this as a reminder to expand the tests that use pexpect and
a pty for screen / window commands.

I have checked that the system tests run as expected using the original Pascal
version as an oracle, as well as running them against this port.

### More tips on setting up Python

You can use `virtual-env` to easily set up an environment without adding to
your global python packages:

```sh
python3 -m venv .venv
source ./.venv/bin/activate
pip install pytest
pip install pexpect
```

**Please be aware** that the `run-system-tests.sh` script included in the
system tests will automatically source `./.venv/bin/activate` if it is found
in the current directory.

## Usage

There is a `man` file in `ludwig.1`.  You can read it without copying it
anywhere by typing:

```sh
man ./ludwig.1
```

Once in the editor, typing `\h` will bring up the help information, assuming
you have installed the help file in the appropriate spot.  Use `\q` to quit
the editor.
