package bootstrap

const runtimeName = "go"

type Status struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Runtime string `json:"runtime"`
	Ready   bool   `json:"ready"`
	Message string `json:"message"`
}

type Service struct {
	name    string
	version string
}

func NewService(name string, version string) *Service {
	return &Service{name: name, version: version}
}

func (s *Service) Status() Status {
	return Status{
		Name:    s.name,
		Version: s.version,
		Runtime: runtimeName,
		Ready:   true,
		Message: "LoomiDBX bootstrap skeleton is ready.",
	}
}
