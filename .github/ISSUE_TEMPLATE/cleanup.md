name: Cleanup
description: Pay down technical debt, reduce friction, etc.
labels: kind/cleanup
body:
- type: textarea
  id: feature
  attributes:
  label: What should be cleaned up or changed?
  validations:
  required: true

- type: textarea
  id: rationale
  attributes:
  label: Provide any links for context
  validations:
  required: true
