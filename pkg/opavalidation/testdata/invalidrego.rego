package fairwinds

# This policy is non-standard and only serves to incorrectly call our insightsinfo function.
callInsightsInfoIncorrectly[clusterName] {
clusterName := insightsinfo("TheCluster")
}

