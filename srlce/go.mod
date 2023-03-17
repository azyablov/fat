module github.com/azyablov/fat/srlce

go 1.18

require (
	github.com/azyablov/fat/lib v0.0.0-00010101000000-000000000000
	github.com/azyablov/fat/lib/gnoi/file v0.0.0-00010101000000-000000000000
	github.com/azyablov/fat/lib/jrpc v0.0.0-00010101000000-000000000000
	github.com/scrapli/scrapligo v1.1.6
	github.com/sirupsen/logrus v1.9.0
)

require (
	github.com/azyablov/gnmi-pg/gnmilib v0.0.0-20230307170529-c2aceddccaa0 // indirect
	github.com/creack/pty v1.1.18 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/openconfig/gnoi v0.0.0-20230221223856-1727ed932554 // indirect
	github.com/sirikothe/gotextfsm v1.0.1-0.20200816110946-6aa2cfd355e4 // indirect
	golang.org/x/crypto v0.5.0 // indirect
	golang.org/x/net v0.5.0 // indirect
	golang.org/x/sys v0.4.0 // indirect
	golang.org/x/text v0.6.0 // indirect
	google.golang.org/genproto v0.0.0-20230110181048-76db0878b65f // indirect
	google.golang.org/grpc v1.53.0 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/azyablov/fat/lib/jrpc => ../lib/jrpc

replace github.com/azyablov/fat/lib => ../lib/

replace github.com/azyablov/fat/lib/gnoi/file => ../lib/gnoi/file
