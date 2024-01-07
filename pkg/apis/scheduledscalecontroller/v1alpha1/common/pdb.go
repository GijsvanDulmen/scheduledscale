package common

type PodDisruptionBudgetEnable struct {
	MinAvailable   *int32 `json:"minAvailable"`
	MaxUnavailable *int32 `json:"maxUnavailable"`
	Enabled        bool   `json:"enabled"`
}

type PdbOnDelete struct {
	RemovePodDisruptionBudget *bool `json:"removePodDisruptionBudget"`
}
