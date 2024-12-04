package fairwinds

import data.utils.array_contains

labelrequired[actionItem] {
    requiredLabelValue := "development"
    provided := [input.metadata.labels[_]]
    not array_contains(provided, requiredLabelValue)
    description := sprintf("Label value %v is missing", [requiredLabelValue])
    actionItem := {
        "title": "Label is missing",
        "description": description,
        "severity": .2,
        "remediation": "Add the label",
        "category": "Reliability"
    }
}