package fairwinds

import data.utils.array_contains

labelrequired[actionItem] {
    requiredLabel := "department2"
    provided := {input.metadata.labels[label]}
    
    array_contains(provided, requiredLabel) == false
    
    input.name == "abc"

    description := sprintf("Label %v is missing", [requiredLabel])
    actionItem := {
        "title": "Label is missing",
        "description": description,
        "severity": .2,
        "remediation": "Add the label",
        "category": "Reliability"
    }
}