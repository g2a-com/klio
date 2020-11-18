# Contributing

If you are new to the klio, try to pick one of the issues labelled as [good first issue][] or [help
wanted][].

## Codding style guidelines

This project is written in go, so remember to run `go fmt` before commit.

## Commit message guidelines

This repository follows [Conventional Commits][] specification, so each commit should be structured
as follows:

```
<type>: <subject>
<BLANK LINE>
[optional body]
<BLANK LINE>
[optional footer(s)]
```

The header (first line) is mandatory, rest is optional.

### Type

Type must be picked from the following list:

- **chore**: changes which do not affect code and do not match to other types
- **ci**: changes to CI configuration files and scripts
- **docs**: documentation _only_ changes
- **feat**: a new feature
- **fix**: a bug fix
- **refactor**: a code change that neither fixes a bug nor adds a feature
- **revert**: a change which reverts some previous commit
- **style**: changes that do not affect the meaning of the code
- **test**: changes which affect only tests

If the commit introduces some breaking changes, append an exclamation mark (!) after a type.

### Subject

Subject is a terse description of the changes introduced by commit. Remember to:

- keep it short, this makes it easier to read on GitHub;
- use the imperative, present tense ("change" not "changed", "add" not "added");
- don't capitalize first letter;
- don't add a dot (.) at the end.

Keep in the mind that commit subjects are used to generate [the changelog](./CHANGELOG.md), so take
a few moments to make them meaningful.

### Body

If you want to add additional information, you may provide a longer description after the subject.

### Footer

If commit introduces breaking changes, add `BREAKING CHANGE: <description>` footer which explains
how exactly it breaks compatibility.

Footer can also [link to an issue][], simply add:

```
Resolves #<issue number>
```

If you are linking more than one issue, place each link in a separate line, eg:

```
Resolves #123
Fixes #456
```

<!-- prettier-ignore-start -->
[good first issue]: https://github.com/g2a-com/klio/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22
[help wanted]: https://github.com/g2a-com/klio/issues?q=is%3Aissue+is%3Aopen+label%3A%22help+wanted%22
[conventional commits]: https://www.conventionalcommits.org/en/v1.0.0/
[link to an issue]: https://docs.github.com/en/free-pro-team@latest/github/managing-your-work-on-github/linking-a-pull-request-to-an-issue#linking-a-pull-request-to-an-issue-using-a-keyword
<!-- prettier-ignore-end -->
