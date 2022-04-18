package bodyBuffer

import (
	"bytes"
	"log"
	"strconv"
	"sync"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestNewBodLogWriter(t *testing.T) {
	e := echo.New()
	e.Use(EchoBodyLogMiddleware)
	e.Use(debug)
	e.Any("*", func(ctx echo.Context) error {
		l, _ := strconv.Atoi(ctx.FormValue("len"))
		_, err := ctx.Response().Write(bytes.Repeat([]byte("*"), l))
		if err != nil {
			return err
		}
		// ctx.Response().Write(bytes.Repeat([]byte("#"), 10))
		return nil
	})
	e.Start(":45000")
}

var (
	requestBodyPool  = sync.Pool{}
	responseBodyPool = sync.Pool{}
)

const BODY_LIMIT = 1024

func init() {
	requestBodyPool.New = func() interface{} {
		return &bytes.Buffer{}
	}
	responseBodyPool.New = func() interface{} {
		return NewBodWriter(10)
	}
}

func EchoBodyLogMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		// request body
		/*reqLen := ctx.Request().ContentLength
		if reqLen > 0 && reqLen < BODY_LIMIT {
			reqBuf := requestBodyPool.Get().(*bytes.Buffer)
			_, _ = reqBuf.ReadFrom(ctx.Request().Body)
			ctx.Request().Body = ioutil.NopCloser(reqBuf)
			ctx.Set(confer.REQUEST_BODY_KEY, reqBuf.Bytes())

			defer func() {
				reqBuf.Reset()
				requestBodyPool.Put(reqBuf)
			}()
		}*/
		// response body
		buf := responseBodyPool.Get().(*BodyWriter)
		buf.ResponseWriter = ctx.Response().Writer
		ctx.Response().Writer = buf
		defer func() {
			buf.Reset()
			responseBodyPool.Put(buf)
		}()
		return next(ctx)
	}
}

func debug(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		res := ctx.Response().Writer.(*BodyWriter)
		err := next(ctx)
		log.Println(res.Body.Len(), res.Body.String())
		return err
	}
}
