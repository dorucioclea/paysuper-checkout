package common

import (
	"bytes"
	"fmt"
	"github.com/labstack/echo/v4"
	billing "github.com/paysuper/paysuper-proto/go/billingpb"
	"io/ioutil"
)

var (
	BinderDefault     = &Binder{}
	EchoBinderDefault = &echo.DefaultBinder{}
)

type Binder struct{}

func (b *Binder) Bind(i interface{}, ctx echo.Context) (err error) {
	if err := EchoBinderDefault.Bind(i, ctx); err != nil {
		return err
	}

	if binder := ExtractBinderContext(ctx); binder != nil {
		return binder.Bind(i, ctx)
	}

	return nil
}

type OrderJsonBinder struct{}
type PaymentCreateProcessBinder struct{}

// Bind
func (cb *OrderJsonBinder) Bind(i interface{}, ctx echo.Context) (err error) {
	var buf []byte

	if ctx.Request().Body != nil {
		buf, err = ioutil.ReadAll(ctx.Request().Body)
		rdr := ioutil.NopCloser(bytes.NewBuffer(buf))

		if err != nil {
			return err
		}

		ctx.Request().Body = rdr
	}

	if err = BinderDefault.Bind(i, ctx); err != nil {
		return err
	}

	structure := i.(*billing.OrderCreateRequest)
	structure.RawBody = string(buf)

	return
}

// Bind
func (cb *PaymentCreateProcessBinder) Bind(i interface{}, ctx echo.Context) (err error) {
	untypedData := make(map[string]interface{})

	if err = BinderDefault.Bind(&untypedData, ctx); err != nil {
		return
	}

	data := i.(map[string]string)

	for k, v := range untypedData {
		switch sv := v.(type) {
		case bool:
			data[k] = "0"

			if sv == true {
				data[k] = "1"
			}
			break
		default:
			data[k] = fmt.Sprintf("%v", sv)
		}
	}

	return
}
