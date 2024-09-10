package utils

hasMatchingHPA(hpas, elem) {
  hpa := hpas[_]
  hpa.spec.scaleTargetRef.kind == elem.kind
  hpa.spec.scaleTargetRef.name == elem.metadata.name
  hpa.metadata.namespace == elem.metadata.namespace
  hpa.spec.scaleTargetRef.apiVersion == elem.apiVersion
}
