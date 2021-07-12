# Output handling

For CLI tools printing things to a terminal is the essential way to communicate with a user. Klio
aims to make it easier for you to implement this mode of communication without using
language-specific frameworks.

Each command run by Klio may write either to stdout or stderr. By default, Klio captures these
outputs and processes them line by line by prefixing lines with a log-level. You can control how
outputs are handled by writing special control sequences which are interpreted by Klio.

This default mode of processing outputs is called "line mode." There is also an alternative raw mode
that disables most of the output processing and simply passes things through.

## Controlling output using escape sequences

A command can control how its output is handled by sending [ANSI escape codes][]. Klio uses APC
sequences that are stripped from the output by most of the terminal emulators. Each such escape
sequence has the following format:

```text
\033_<command-name> <json-encoded-parameters>\033\\
```

Keep in mind that each stream (stdout and stderr) is handled separately. Sending control sequence to
stdout is not going to affect stderr.

A more detailed description of this format is placed
[at the end of this document](#escape-sequences-grammar). But if you don't want to dive into the
nitty-gritty details of grammar, simply follow the examples or use one of the libraries.

## Line mode

By default, Klio is working in this mode. In the line mode each line is prefixed with log level and
tags:

```text
[INFO][TAG][ANOTHER TAG] original output printed by command
```

Lines that have log-level lower than minimum specified by the user (by default info) are stripped
from the output.

### Log levels

Klio assigns each line with a log level. If you don't change it, by default, all lines printed to
the stdout have the "info" level and lines printed to the stderr have the "error" level. You may
change log level by sending `klio_log_level` command, but you have to use one of the following log
levels:

| Log level | Description                                                                              |
| --------- | ---------------------------------------------------------------------------------------- |
| fatal     | Errors causing a command to exit immediately                                             |
| error     | Errors which cause a command to fail, but not immediately                                |
| warn      | Information about unexpected situations and minor errors (not causing a command to fail) |
| info      | Generally useful information (_things happen_)                                           |
| verbose   | More granular but still useful information                                               |
| debug     | Information helpful for command developers                                               |
| spam      | \*Give me **everything\***                                                               |

**Examples**

- `\033_klio_log_level "spam"\033\\` – sets log level to spam
- `\033_klio_log_level "warn"\033\\` – sets log level to warn

### Tags

Using `klio_tags` command you may specify multiple tags which will be added to each line. You may
use these tags to distinguish various steps of command execution.

**Examples**

- `\033_klio_tags ["foo", "bar"]\033\\` – sets tags to `[FOO][BAR]`
- `\033_klio_tags []\033\\` – disables tags

## Raw mode

Raw mode neither buffers nor modifies output, it simply passes it through unchanged. Use it if you
want to implement some spinner, progress bar or even [text-based user interface][]. In order to
enable raw mode, use `klio_mode` command.

**Examples**

- `\033_klio_mode "raw"\033\\` – enables raw mode
- `\033_klio_mode "line"\033\\` – goes back to line mode

## Resetting

You can reset mode, log level, and tags to default using `klio_reset` command.

**Examples**

- `\033_klio_reset\033\\` - resets to default settings

## Escape sequences grammar

```abnf
control-sequence = escape underscore command escape backslash

; Commands
command           = log-level-command / tags-command / mode-command / reset-command
log-level-command = "klio_log_level" space quotation-mark log-level quotation-mark
tags-command      = "klio_tags" space strings-array
mode-command      = "klio_mode" space quotation-mark mode quotation-mark
reset-command     = "klio_reset"

; Enums
log-level = "spam" / "debug" / "verbose" / "info" / "warn" / "error" / "fatal"
mode      = "line" / "raw"

; Supported subset of JSON
strings-array   = begin-array [ string *( value-separator string ) ] end-array
string          = quotation-mark *char quotation-mark
char            = %x20-21 / %x23-5B / %x5D-10FFFF / backslash (%x22 / %x5C / %x2F / %x62 / %x66 / %x6E / %x72 / %x74 / %x75 4hex-digit)
begin-array     = *space "[" *space
end-array       = *space "]" *space
value-separator = *space "," *space

; Other
escape          = %x1B
backslash       = "\"
underscore      = "_"
space           = " "
quotation-mark  = %x22
hex-digit       = %x30–39 / %x41-46 / %x61-66
```

[text-based user interface]: https://en.wikipedia.org/wiki/Text-based_user_interface
[ANSI escape codes]: https://en.wikipedia.org/wiki/ANSI_escape_code

