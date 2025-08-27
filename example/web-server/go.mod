module web-server-example

go 1.23.4

replace github.com/kydenul/log => ../..

require github.com/kydenul/log v0.0.0-00010101000000-000000000000

require (
	go.uber.org/multierr v1.10.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
