package models

/*
╔════════════════════════════════════════════════════════════════╗
║                                                                ║
║  DECISION.GO - MODELO DE DECISIÓN                              ║
║                                                                ║
║  Representa la decisión tomada por el servicio de reglas       ║
║  sobre qué hacer con el residuo clasificado.                   ║
║                                                                ║
╚════════════════════════════════════════════════════════════════╝
*/

// DecisionAction representa la acción a tomar.
type DecisionAction string

const (
	// DecisionActionAccept - Aceptar el residuo en el compartimento indicado.
	DecisionActionAccept DecisionAction = "accept"

	// DecisionActionReject - Rechazar el residuo (no va en este contenedor).
	DecisionActionReject DecisionAction = "reject"

	// DecisionActionManualReview - Requiere revisión manual.
	DecisionActionManualReview DecisionAction = "manual_review"
)

// }.
type Decision struct {
	// ═══════════════════════════════════════════════════════════════
	// DECISIÓN PRINCIPAL
	// ═══════════════════════════════════════════════════════════════

	// Action - Acción a tomar (accept, reject, manual_review)
	Action string `json:"action" dynamodbav:"action"`

	// BinCompartment - Compartimento del contenedor (si action=accept)
	// Valores posibles: "recyclable", "organic", "general"
	BinCompartment string `json:"bin_compartment,omitempty" dynamodbav:"bin_compartment,omitempty"`

	// Message - Mensaje descriptivo para el usuario
	Message string `json:"message" dynamodbav:"message"`

	// ═══════════════════════════════════════════════════════════════
	// VALIDACIÓN
	// ═══════════════════════════════════════════════════════════════

	// ConfidenceThresholdMet - Si la confianza de clasificación cumple el umbral
	ConfidenceThresholdMet bool `json:"confidence_threshold_met" dynamodbav:"confidence_threshold_met"`

	// ConfidenceThreshold - Umbral de confianza usado (ej: 0.7)
	ConfidenceThreshold float64 `json:"confidence_threshold,omitempty" dynamodbav:"confidence_threshold,omitempty"`

	// ═══════════════════════════════════════════════════════════════
	// REGLAS APLICADAS
	// ═══════════════════════════════════════════════════════════════

	// RuleApplied - ID de la regla que se aplicó
	RuleApplied string `json:"rule_applied,omitempty" dynamodbav:"rule_applied,omitempty"`

	// RuleVersion - Versión del motor de reglas
	RuleVersion string `json:"rule_version,omitempty" dynamodbav:"rule_version,omitempty"`

	// ═══════════════════════════════════════════════════════════════
	// RAZONES Y METADATA
	// ═══════════════════════════════════════════════════════════════

	// Reasons - Razones detalladas de la decisión
	Reasons []string `json:"reasons,omitempty" dynamodbav:"reasons,omitempty"`

	// Metadata - Información adicional de la decisión
	Metadata map[string]interface{} `json:"metadata,omitempty" dynamodbav:"metadata,omitempty"`
}

// ═══════════════════════════════════════════════════════════════════
//                     MÉTODOS DEL MODELO
// ═══════════════════════════════════════════════════════════════════

// IsAccepted retorna true si la decisión es aceptar el residuo.
func (d *Decision) IsAccepted() bool {
	return d.Action == string(DecisionActionAccept)
}

// IsRejected retorna true si la decisión es rechazar el residuo.
func (d *Decision) IsRejected() bool {
	return d.Action == string(DecisionActionReject)
}

// RequiresManualReview retorna true si requiere revisión manual.
func (d *Decision) RequiresManualReview() bool {
	return d.Action == string(DecisionActionManualReview)
}

// AddReason agrega una razón a la decisión.
func (d *Decision) AddReason(reason string) {
	if d.Reasons == nil {
		d.Reasons = []string{}
	}
	d.Reasons = append(d.Reasons, reason)
}

// Validate valida que la decisión tenga los campos obligatorios.
func (d *Decision) Validate() error {
	if d.Action == "" {
		return ErrInvalidAction
	}

	// Si la acción es accept, debe tener bin_compartment
	if d.IsAccepted() && d.BinCompartment == "" {
		return ErrInvalidBinCompartment
	}

	if d.Message == "" {
		return ErrInvalidMessage
	}

	return nil
}

/*
═══════════════════════════════════════════════════════════════════
                    EJEMPLO DE USO
═══════════════════════════════════════════════════════════════════

// Decisión de aceptar
decision := &Decision{
    Action:                 string(DecisionActionAccept),
    BinCompartment:         "recyclable",
    Message:                "Plastic bottle accepted in recyclable bin",
    ConfidenceThresholdMet: true,
    ConfidenceThreshold:    0.7,
    RuleApplied:            "recyclable_plastics_high_confidence",
}

// Agregar razones
decision.AddReason("High confidence classification (94%)")
decision.AddReason("Material type matches bin type")

// Verificar tipo de decisión
if decision.IsAccepted() {
    fmt.Printf("Item accepted in %s compartment\n", decision.BinCompartment)
}

// Decisión de revisión manual
manualDecision := &Decision{
    Action:                 string(DecisionActionManualReview),
    Message:                "Low confidence - manual review required",
    ConfidenceThresholdMet: false,
    ConfidenceThreshold:    0.7,
}

if manualDecision.RequiresManualReview() {
    fmt.Println("Requires manual intervention")
}

═══════════════════════════════════════════════════════════════════
*/
