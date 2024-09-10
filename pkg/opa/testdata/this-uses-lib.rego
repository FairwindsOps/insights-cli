package fairwinds

import data.utils.hasMatchingHPA

hpaRequired[actionItem] {
  not hasMatchingHPA(kubernetes("autoscaling", "HorizontalPodAutoscaler"), input)
  actionItem := {
    "title": "No horizontal pod autoscaler found"
  }
}
