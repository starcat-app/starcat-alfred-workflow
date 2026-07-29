// Package alfredjson 定义 Alfred Script Filter 的最小稳定 JSON 结构。
package alfredjson

type Output struct {
	Items     []Item            `json:"items"`
	Variables map[string]string `json:"variables,omitempty"`
	Rerun     *float64          `json:"rerun,omitempty"`
}

type Item struct {
	UID          string `json:"uid,omitempty"`
	Title        string `json:"title"`
	Subtitle     string `json:"subtitle,omitempty"`
	Arg          string `json:"arg,omitempty"`
	Autocomplete string `json:"autocomplete,omitempty"`
	Match        string `json:"match,omitempty"`
	Valid        bool   `json:"valid"`
	Icon         Icon   `json:"icon"`
}

type Icon struct {
	Path string `json:"path"`
}
