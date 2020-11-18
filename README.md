# klio

Klio is a crossover between a framework for creating CLI tools and a package manager for installing
them into your project. It can be used as a standalone tool, or as a library which enables you to
create your own CLI.

## How it works

By default klio doesn't include any commands except "get" which enables you to install new ones:

```
klio get hello --from https://raw.githubusercontent.com/g2a-com/klio-example-command/main/registry.yaml
```

Now you can use the newly installed command:

```
klio hello
```

By default "get" adds info about each installed command to the "klio.yaml" file. You can easily
install all dependencies listed in this file by running:

```
klio get
```

## Installation

Currently, you have to compile klio by yourself. Make sure that you have
[golang compiler](https://golang.org/dl/) installed. Next, clone repository and run "go build":

```
git clone git@github.com:g2a-com/klio.git
cd klio
go build ./cmd/klio
```

## Contributing

To contribute to klio, check out [contribution guidelines](./CONTRIBUTING.md).
