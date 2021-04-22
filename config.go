package prototype

type Config struct {
	Inputs  []Input  `json:"inputs,omitempty"`
	Outputs []Output `json:"outputs,omitempty"`
	Caches  []Cache  `json:"caches,omitempty"`
}

type Input struct {
	Name string `json:"name"`
	Path string `json:"path,omitempty"`
}

type Output struct {
	Name string `json:"name"`
	Path string `json:"path,omitempty"`
}

type Cache struct {
	Path string `json:"path,omitempty"`
}

func EmptyConfig(_ interface{}) Config {
	return Config{}
}
