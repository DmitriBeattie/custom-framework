package config

type File struct {
	Path string `json:"path"`
}

func (f *File) InstanceKind() string {
	return "file"
}
