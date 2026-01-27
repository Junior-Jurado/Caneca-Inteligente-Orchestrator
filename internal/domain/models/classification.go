// Package models defines the domain models for the Smart Bin Orchestrator.
package models

// ClassificationLabel represents the type of waste identified.
type ClassificationLabel string

// Classification represents the result of the ML classification service.
// It contains the predicted label, confidence score, model version, and alternative predictions.
type Classification struct {
	Label          string        `json:"label" dynamodbav:"label"`
	Confidence     float64       `json:"confidence" dynamodbav:"confidence"`
	ModelVersion   string        `json:"model_version" dynamodbav:"model_version"`
	Alternatives   []Alternative `json:"alternatives,omitempty" dynamodbav:"alternatives,omitempty"`
	ProcessingTime int64         `json:"processing_time_ms" dynamodbav:"processing_time_ms"`
}

// Alternative represents an alternative classification prediction.
// It contains the next most probable labels with their confidence scores.
type Alternative struct {
	Label      string  `json:"label" dynamodbav:"label"`
	Confidence float64 `json:"confidence" dynamodbav:"confidence"`
}

const (
	// LabelPlasticBottle represents a plastic bottle.
	LabelPlasticBottle ClassificationLabel = "plastic_bottle"

	// LabelPlasticContainer represents a plastic container.
	LabelPlasticContainer ClassificationLabel = "plastic_container"

	// LabelGlassBottle represents a glass bottle.
	LabelGlassBottle ClassificationLabel = "glass_bottle"

	// LabelAluminumCan represents an aluminum can.
	LabelAluminumCan ClassificationLabel = "aluminum_can"

	// LabelPaper represents paper waste.
	LabelPaper ClassificationLabel = "paper"

	// LabelCardboard represents cardboard waste.
	LabelCardboard ClassificationLabel = "cardboard"

	// LabelOrganicWaste represents organic waste.
	LabelOrganicWaste ClassificationLabel = "organic_waste"

	// LabelGeneralWaste represents general non-recyclable waste.
	LabelGeneralWaste ClassificationLabel = "general_waste"
)

// IsRecyclable returns true if the classified item is recyclable.
func (c *Classification) IsRecyclable() bool {
	recyclables := map[string]bool{
		"plastic_bottle":    true,
		"plastic_container": true,
		"glass_bottle":      true,
		"aluminum_can":      true,
		"paper":             true,
		"cardboard":         true,
	}
	return recyclables[c.Label]
}

// IsHighConfidence returns true if the confidence is above 80%.
func (c *Classification) IsHighConfidence() bool {
	return c.Confidence >= 0.8
}

// ShouldReview returns true if the classification should be manually reviewed due to low confidence.
func (c *Classification) ShouldReview() bool {
	return c.Confidence < 0.7
}

// GetTopAlternatives returns the top N alternative predictions.
func (c *Classification) GetTopAlternatives(n int) []Alternative {
	if len(c.Alternatives) <= n {
		return c.Alternatives
	}
	return c.Alternatives[:n]
}
