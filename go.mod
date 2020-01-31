module github.com/paysuper/paysuper-checkout

require (
	github.com/ProtocolONE/go-core/v2 v2.1.0
	github.com/PuerkitoBio/purell v1.1.1
	github.com/alexeyco/simpletable v0.0.0-20190222165044-2eb48bcee7cf
	github.com/fatih/color v1.7.0
	github.com/globalsign/mgo v0.0.0-20181015135952-eeefdecb41b8
	github.com/go-log/log v0.1.0
	github.com/golang/groupcache v0.0.0-20191002201903-404acd9df4cc // indirect
	github.com/google/uuid v1.1.1
	github.com/google/wire v0.3.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.1.0 // indirect
	github.com/gurukami/typ/v2 v2.0.1
	github.com/labstack/echo/v4 v4.1.11
	github.com/micro/go-micro v1.18.0
	github.com/micro/go-plugins v1.2.0
	github.com/paysuper/paysuper-proto/go/billingpb v0.0.0-20200122152034-6d6bb19abebf
	github.com/pkg/errors v0.8.1
	github.com/spf13/cobra v0.0.5
	github.com/stretchr/testify v1.4.0
	github.com/ttacon/builder v0.0.0-20170518171403-c099f663e1c2 // indirect
	github.com/ttacon/libphonenumber v1.0.1
	go.uber.org/automaxprocs v1.2.0
	gopkg.in/go-playground/validator.v9 v9.30.0
	gopkg.in/paysuper/paysuper-database-mongo.v1 v1.0.0-20191120092306-dc35c6f924f1 // indirect
)

replace (
	github.com/gogo/protobuf v0.0.0-20190410021324-65acae22fc9 => github.com/gogo/protobuf v1.2.2-0.20190723190241-65acae22fc9d
	github.com/hashicorp/consul => github.com/hashicorp/consul v1.5.1
	github.com/hashicorp/consul/api => github.com/hashicorp/consul/api v1.1.0
	github.com/micro/go-micro => github.com/micro/go-micro v1.8.0
)

go 1.12
