package fairwinds
foo := {"s": "foo"}
envMaxReplicasDeployments[actionItem] {
    print(foo.s)
    input.kind == "Deployment"
    env_suffix := array.reverse(split(input.metadata.namespace, "-"))[0]
    replicas := input.spec.replicas
    actionItem := {
      "title": "Non-production environment replica count exceeds maximum",
      "description": sprintf("The Deployment %v in the %v environment replicas exceed the maximum replica count for this environment.", [input.metadata.name, env_suffix]),
      "severity": 0.5,
      "remediation": "Reduce the number of replicas",
      "category": "Reliability"
    }
}
