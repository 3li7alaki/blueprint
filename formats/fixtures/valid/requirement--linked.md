scope: cli/internal/billing/**, cli/internal/checkout/**
confidence: derived
ears: WHILE [[money:amount]] is pending, THE system SHALL retain [[billing:invoice]].
fit: property: Every [[billing:invoice]] preserves [[money:amount]].
evidence: cli/internal/billing/amount.go:12
needs: billing/defines-amount, billing/loads-invoice
after: billing/opens-checkout
tier: consequential
superseded-by: billing/retains-complete-invoice

Rationale keeps its spacing.  

[[body text is not validated
unknown: body content
