package bodyBuffer

import (
	"bytes"
	"net/http"
)

type BodyWriter struct {
	http.ResponseWriter
	Body *bytes.Buffer
	// len        int // 记录总长度
	cloneLimit int // 响应镜像复制大小限制
}

func NewBodWriter(cloneLimit int) *BodyWriter {
	return &BodyWriter{
		ResponseWriter: nil,
		Body:           bytes.NewBuffer(nil),
		cloneLimit:     cloneLimit,
	}
}

func (w *BodyWriter) Write(b []byte) (int, error) {
	if len(b) == 0 {
		return 0, nil
	}
	// 验证响应体大小,
	// w.len += len(b)
	// clone body
	l := w.Body.Len()
	if l <= w.cloneLimit {
		if len(b) > w.cloneLimit-l {
			w.Body.Write(b[0 : w.cloneLimit-l])
		} else {
			w.Body.Write(b)
		}
	}
	/*if w.len <= w.cloneLimit {
		w.Body.Write(b)
	}*/
	return w.ResponseWriter.Write(b)
}

func (w *BodyWriter) Reset() {
	w.Body.Reset()
	// w.len = 0
}
