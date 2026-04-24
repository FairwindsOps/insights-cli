<div align="center" class="no-border">
  <img src="logo.png" alt="Insights Logo" width="650">
  <br>
  <h3>A CLI for Fairwinds Insights</h3>
  <a href="https://opensource.org/licenses/Apache-2.0">
    <img src="https://img.shields.io/badge/license-Apache2-brightgreen.svg">
  </a>
  <a href="https://circleci.com/gh/FairwindsOps/insights-cli">
    <img src="https://circleci.com/gh/FairwindsOps/insights-cli.svg?style=svg">
  </a>
</div>

# Insights CLI
A command line tool for interacting with Fairwinds Insights.

**Want to learn more?** Reach out on [the Slack channel](https://fairwindscommunity.slack.com/messages/fairwinds-insights) ([request invite](https://join.slack.com/t/fairwindscommunity/shared_invite/zt-e3c6vj4l-3lIH6dvKqzWII5fSSFDi1g)), send an email to `opensource@fairwinds.com`, or join us for [office hours on Zoom](https://fairwindscommunity.slack.com/messages/office-hours)

## Documentation
Check out the docs at [docs.fairwinds.com](https://insights.docs.fairwinds.com/configure/policy/cli/).

## Contributing

PRs welcome! Check out the [Contributing Guidelines](CONTRIBUTING.md) and
[Code of Conduct](CODE_OF_CONDUCT.md) for more information.

## Notice: Registry Migration and Immutable Images (v3.1.10 → v3.2.0)

Starting with **v3.2.0**:

- Images moved to `us-docker.pkg.dev/fairwinds-ops/oss/insights-cli`
- `quay.io/fairwinds/insights-cli` is deprecated

### Required action

```diff
- quay.io/fairwinds/insights-cli:<tag>
+ us-docker.pkg.dev/fairwinds-ops/oss/insights-cli:<tag>
```

---

## Immutable and signed images

* Images are now **signed**
* Tags are **immutable**
* No more floating tags:

  * `v3`
  * `v3.1`
  * `latest`

Use full version tags:

```
us-docker.pkg.dev/fairwinds-ops/oss/insights-cli:v<major>.<minor>.<patch>
```

Or pin by digest:

```
us-docker.pkg.dev/fairwinds-ops/oss/insights-cli@sha256:<digest>
```