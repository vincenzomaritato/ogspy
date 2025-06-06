# Contributing to ogspy

First off, thanks for taking the time to contribute!  
Open-source thrives on community-driven improvements and we appreciate your help.

## Ground rules

1. **Search first** – Please look for existing issues or PRs before opening a new one.
2. **One change per PR** – Keep pull requests focused and self-contained.
3. **Tests & linting** – CI must be green (`make lint test`).
4. **Signed commits** – Use the [DCO](https://developercertificate.org/) sign-off (`git commit -s`).
5. **Respectful communication** – Follow our [Code of Conduct](CODE_OF_CONDUCT.md).

## Development workflow

```bash
# 1. Fork and clone
git clone https://github.com/vincenzomaritato/ogspy.git
cd ogspy
git checkout -b feat/<short-description>

# 2. Hack away!
make lint test        # run linters & unit tests
go vet ./...          # extra static analysis

# 3. Commit (with DCO)
git add .
git commit -s -m "feat: short, imperative description"

# 4. Push and open a PR
git push origin feat/<short-description>
```

## Commit message guidelines

We loosely follow the Conventional Commits spec:

```txt
<type>[optional scope]: <description>

[optional body]
[optional footer(s)]
```

Common type values: feat, fix, docs, test, refactor, chore.
If your change fixes an issue, add `Fixes #123` in the footer.

## Running the test suite

```bash
make test           # unit & integration
make fuzz           # 10-second fuzzing run
```

To inspect coverage:

```bash
go tool cover -html=dist/coverage.out
```

## Adding a new feature

1. Open a Feature Request issue describing the use-case.
2. Wait for maintainer feedback / approval.
3. Implement the feature behind a flag if it alters default behaviour.
4. Update documentation ([README.md](README.md) or docs/).

## Releasing

Only maintainers can tag releases. If your PR is slated for the next release, add the release-note label and include a Changelog entry in [CHANGELOG.md](CHANGELOG.md) under ## [Unreleased].

## Contact

- General questions: open an issue.
- Security reports: never open a public issue—email [hello@vmaritato.com](mailto://hello@vmaritato.com) instead.

Happy hacking! 🎉
