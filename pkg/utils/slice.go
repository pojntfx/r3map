package utils

type SliceWriter struct {
	slice []byte
}

func NewSliceWriter(slice []byte) *SliceWriter {
	return &SliceWriter{slice}
}

func (w *SliceWriter) Write(p []byte) (n int, err error) {
	n = copy(w.slice, p)

	w.slice = w.slice[n:]

	return n, nil
}
