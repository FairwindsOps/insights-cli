package fairwinds
makeValidActionItem[actionItem] {
    actionItem := {
        "title": "valid",
        "description": "this action item has the correct fields",
        "severity": 0.9,
        "remediation": "nothing",
        "category": "Security"
    }
}

makeInvalidActionItem[actionItem] {
    actionItem := {
        "title": "invalid",
        "description": "this action item has an incorrect category field",
        "severity": 0.9,
        "remediation": "nothing",
        "category": "security"
    }
}
