# Output handling

For CLI tools printing output to a terminal is the basic way to communicate with a user. Klio injects `KLIO_LOG_LEVEL` environment varianble into each run subcommand. Its value is equal to the currently set log level for Klio command. You can use the variable to dynamically set the log level in your subcommand. If your subcommand is written in Go, you can use the logger exported in `github.com/g2a-com/klio/pkg/log`. 

### Log levels

Log levels used by Klio and their meaning is as follows:

| Log level | Description                                                                              |
| --------- | ---------------------------------------------------------------------------------------- |
| fatal     | Errors causing a command to exit immediately                                             |
| error     | Errors which cause a command to fail, but not immediately                                |
| warn      | Information about unexpected situations and minor errors (not causing a command to fail) |
| info      | Generally useful information (_things happen_)                                           |
| verbose   | More granular but still useful information                                               |
| debug     | Information helpful for command developers                                               |
| spam      | \*Give me **everything\***                                                               |