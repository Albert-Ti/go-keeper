package utils

type ChunkedReader interface {
	GetChunk() []byte
}

type StreamReader[T ChunkedReader] interface {
	Recv() (T, error)
}

type GoKeeperStream[T ChunkedReader] struct {
	Stream StreamReader[T]
	buf    []byte
	Total  int64
}

func (r *GoKeeperStream[T]) Read(p []byte) (int, error) {

	for len(r.buf) == 0 {
		req, err := r.Stream.Recv()
		if err != nil {
			return 0, err
		}

		if chunk := req.GetChunk(); len(chunk) > 0 {
			r.buf = chunk
		}
	}

	n := copy(p, r.buf)
	r.buf = r.buf[n:]
	r.Total += int64(n)

	return n, nil
}
