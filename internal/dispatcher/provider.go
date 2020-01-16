package dispatcher

import (
	"context"
	"github.com/ProtocolONE/go-core/v2/pkg/config"
	"github.com/ProtocolONE/go-core/v2/pkg/invoker"
	"github.com/ProtocolONE/go-core/v2/pkg/provider"
	"github.com/google/wire"
	"github.com/paysuper/paysuper-checkout/internal/dispatcher/common"
	"github.com/paysuper/paysuper-checkout/internal/validators"
	"github.com/paysuper/paysuper-checkout/pkg/micro"
	billing "github.com/paysuper/paysuper-proto/go/billingpb"
	"gopkg.in/go-playground/validator.v9"
)

// ProviderCfg
func ProviderCfg(cfg config.Configurator) (*Config, func(), error) {
	c := &Config{
		WorkDir: cfg.WorkDir(),
		invoker: invoker.NewInvoker(),
	}
	e := cfg.UnmarshalKeyOnReload(common.UnmarshalKey, c)
	return c, func() {}, e
}

// ProviderGlobalCfg
func ProviderGlobalCfg(cfg config.Configurator) (*common.Config, func(), error) {
	c := &common.Config{}
	e := cfg.UnmarshalKey(common.UnmarshalGlobalConfigKey, c)
	return c, func() {}, e
}

// ProviderServices
func ProviderServices(srv *micro.Micro) common.Services {
	return common.Services{
		Billing: billing.NewBillingService(billing.ServiceName, srv.Client()),
	}
}

// ProviderValidators
func ProviderValidators(v *validators.ValidatorSet) (validate *validator.Validate, _ func(), err error) {
	validate = validator.New()
	if err = validate.RegisterValidation("phone", v.PhoneValidator); err != nil {
		return
	}
	if err = validate.RegisterValidation("uuid", v.UuidValidator); err != nil {
		return
	}
	if err = validate.RegisterValidation("zip_usa", v.ZipUsaValidator); err != nil {
		return
	}
	if err = validate.RegisterValidation("name", v.NameValidator); err != nil {
		return
	}
	if err = validate.RegisterValidation("city", v.CityValidator); err != nil {
		return
	}
	if err = validate.RegisterValidation("locale", v.UserLocaleValidator); err != nil {
		return
	}
	return validate, func() {}, nil
}

// ProviderDispatcher
func ProviderDispatcher(ctx context.Context, set provider.AwareSet, appSet AppSet, cfg *Config, globalCfg *common.Config, ms *micro.Micro) (*Dispatcher, func(), error) {
	d := New(ctx, set, appSet, cfg, globalCfg, ms)
	return d, func() {}, nil
}

var (
	// Dependencies: go-shared/provider.AwareSet, internal/*validators.ValidatorSet, pkg/micro.Micro, go-shared/config.Configurator, ProviderHandlers
	WireSet = wire.NewSet(
		ProviderDispatcher,
		ProviderServices,
		ProviderValidators,
		ProviderCfg,
		ProviderGlobalCfg,
		wire.Struct(new(AppSet), "*"),
	)
	// Dependencies: go-shared/provider.AwareSet, internal/*validators.ValidatorSet, common.Services, common.Handlers, go-shared/config.Configurator
	WireTestSet = wire.NewSet(
		ProviderDispatcher,
		ProviderValidators,
		ProviderCfg,
		ProviderGlobalCfg,
		wire.Struct(new(AppSet), "*"),
	)
)
