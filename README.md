# Insights-cli
A command line tool for Fairwinds Insights

## Policy Functions

These are functions related to the Rules Engine or OPA polices.

### Sync

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

