package main

import (
	"github.com/paysuper/paysuper-checkout/cmd/http"
	"github.com/paysuper/paysuper-checkout/cmd/root"
)

// @title PaySuper payment solution service.
// @desc Swagger Specification for PaySuper Payment Form API.
//
// @ver 1.0.0
// @server https://checkout.pay.super.com Production API
func main() {
	args := []string{
		"http", "-c", "configs/local.yaml", "-d",
	}
	root.ExecuteDefault(args, http.Cmd)
}
