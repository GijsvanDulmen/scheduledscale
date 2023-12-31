package annotations

type Annotations struct {
	Add    map[string]string `json:"add"`
	Remove []string          `json:"remove"`
}
