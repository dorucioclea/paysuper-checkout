package common

type Config struct {
	CookieDomain string `envconfig:"COOKIE_DOMAIN" required:"true"`
	CookieName   string `envconfig:"COOKIE_NAME" default:"_ps_ctkn"`
	AllowOrigin  string `envconfig:"ALLOW_ORIGIN" default:"*"`

	// OrderInlineFormUrlMask url like a https://checkout.tst.pay.super.com/pay/order/
	OrderInlineFormUrlMask string `envconfig:"ORDER_INLINE_FORM_URL_MASK" required:"true"`

	CustomerTokenCookiesLifetimeHours int64 `envconfig:"CUSTOMER_TOKEN_COOKIES_LIFETIME" default:"720"`
}
