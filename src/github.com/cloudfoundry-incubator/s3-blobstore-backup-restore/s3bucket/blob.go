package s3bucket

type Blob struct {
	path string
}

func NewBlob(path string) Blob {
	return Blob{
		path: path,
	}
}

func (b Blob) Path() string {
	return b.path
}
