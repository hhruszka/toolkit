module linuxtester

go 1.23.5

toolchain go1.24.1

require (
	github.com/fatih/color v1.18.0
	github.com/hhruszka/secretscanner v1.0.6
	github.com/jedib0t/go-pretty/v6 v6.6.7
	github.com/spf13/cobra v1.9.1
	golang.org/x/exp v0.0.0-20250305212735-054e65f0b394
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/go-enry/go-enry/v2 v2.9.2 // indirect
	github.com/go-enry/go-oniguruma v1.2.1 // indirect
	github.com/golang-jwt/jwt/v5 v5.2.1 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.16 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/spf13/pflag v1.0.6 // indirect
	golang.org/x/net v0.0.0-20181017193950-04a2e542c03f // indirect
	golang.org/x/sys v0.31.0 // indirect
	golang.org/x/text v0.23.0 // indirect
)

replace github.com/hhruszka/secretscanner v1.0.6 => ../secretscanner-main
