package fairwinds

labelrequired[actionItem] {
    requiredLabelValue := "development"
    provided := [input.metadata.labels[_]]
    description := sprintf("Label value %v is missing", [requiredLabelValue])
    actionItem := {
        "title": "Label is missing",
        "description": description,
        "severity": .2,
        "remediation": "Add the label",
        "category": "Reliability"
    }
}