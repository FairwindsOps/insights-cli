# Insights-cli

[![FairwindsOps](https://circleci.com/gh/FairwindsOps/insights-cli.svg?style=svg)](https://circleci.com/gh/FairwindsOps/insights-cli)
[![Apache 2.0 license](https://img.shields.io/badge/license-Apache2-brightgreen.svg)](https://opensource.org/licenses/Apache-2.0)

A command line tool for interacting with Fairwinds Insights.

**Want to learn more?** Reach out on [the Slack channel](https://fairwindscommunity.slack.com/messages/fairwinds-insights) ([request invite](https://join.slack.com/t/fairwindscommunity/shared_invite/zt-e3c6vj4l-3lIH6dvKqzWII5fSSFDi1g)), send an email to `opensource@fairwinds.com`, or join us for [office hours on Zoom](https://fairwindscommunity.slack.com/messages/office-hours)


## Contributing

PRs welcome! Check out the [Contributing Guidelines](CONTRIBUTING.md) and
[Code of Conduct](CODE_OF_CONDUCT.md) for more information.

## Installation

### Binary

Install the binary from our [releases](https://github.com/FairwindsOps/insights-cli/releases) page.

### Go Get

```
go get -u github.com/fairwindsops/insights-cli/cmd/insights
```

## Usage

The Insights CLI requires a configuration file which by default is named `fairwinds-insights.yaml` in your current directory. You can overwrite this with the `--config` flag.

### Policy

These are functions related to the Rules Engine or OPA polices.

#### Sync

Synchronize OPA policies to Fairwinds Insights.

The sync functionality expects a directory structure like the following.

```
.
+-- policy1
|   +-- policy.yaml
|   +-- instance1.yaml
+-- policy2
|   +-- policy.rego
|   +-- instance1.yaml
```

This will create two policies, one called `policy1` the other called `policy2` which each have an instance named `instance1`. The rego for these policies will be pulled from the `policy.x` file inside that folder. See the [examples](https://github.com/FairwindsOps/insights-plugins/tree/master/plugins/opa/examples) folder for more examples.


