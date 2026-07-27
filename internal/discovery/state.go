package discovery

import "errors"

// NodeState define el tipo estricto para los estados del ciclo de vida del nodo.
// Al crear un tipo personalizado, blindamos el código en tiempo de compilación.
type NodeState string

// Constantes inmutables de los estados de negocio
const (
	StateDiscovered   NodeState = "DISCOVERED"
	StateInspecting   NodeState = "INSPECTING"
	StateReady        NodeState = "READY"
	StateProvisioning NodeState = "PROVISIONING"
	StateProvisioned  NodeState = "PROVISIONED"
	StateMaintenance  NodeState = "MAINTENANCE"
	StateReleasing    NodeState = "RELEASING"
	StateFailed       NodeState = "FAILED"
)

// ErrInvalidTransition se devuelve cuando se intenta un cambio de estado lógico prohibido.
var ErrInvalidTransition = errors.New("transición de estado no válida para las reglas de negocio")

// validTransitions define el motor lógico de la FSM (Finite State Machine).
// Mapea desde un estado actual hacia qué otros estados está permitido viajar.
var validTransitions = map[NodeState][]NodeState{
	StateDiscovered:   {StateInspecting, StateFailed},
	StateInspecting:   {StateReady, StateFailed},
	StateReady:        {StateProvisioning, StateMaintenance, StateReleasing},
	StateProvisioning: {StateProvisioned, StateFailed},
	StateProvisioned:  {StateMaintenance, StateReleasing},
	StateMaintenance:  {StateReady, StateReleasing},
	StateReleasing:    {StateReady, StateFailed},
	// Si un nodo falla, permitimos que un operador lo recupere pasándolo a READY,
	// lo ponga en MANTENIMIENTO o lo libere directamente.
	StateFailed: {StateReady, StateMaintenance, StateReleasing},
}

// CanTransition verifica si el movimiento del estado 'current' al 'next' está permitido.
func CanTransition(current, next NodeState) bool {
	allowedStates, exists := validTransitions[current]
	if !exists {
		return false
	}

	for _, state := range allowedStates {
		if state == next {
			return true
		}
	}
	return false
}

// Transition evalúa y ejecuta el cambio de estado, devolviendo un error si es ilegal.
func Transition(current, next NodeState) (NodeState, error) {
	if CanTransition(current, next) {
		return next, nil
	}
	return current, ErrInvalidTransition
}
