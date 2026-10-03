package utils

import (
	pb "github.com/Albert-Ti/go-keeper/pkg/proto"
)

type StreamReader struct {
	Stream pb.GoKeeperService_CreateArbitraryDataServer
	buf    []byte
	Total  int64
}

func (r *StreamReader) Read(p []byte) (int, error) {

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
