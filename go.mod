module fnd

go 1.12

replace fnd.localhost/dwire => ./vendor/fnd.localhost/dwire

replace fnd.localhost/handshake => ./vendor/fnd.localhost/handshake

require (
	fnd.localhost/dwire v1.0.1
	fnd.localhost/handshake v0.0.0-20200428084808-2c986090302e
	github.com/btcsuite/btcd v0.20.1-beta
	github.com/golang/protobuf v1.4.2
	github.com/golang/snappy v0.0.1 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/kr/pretty v0.1.0 // indirect
	github.com/mattn/go-isatty v0.0.11
	github.com/mattn/go-runewidth v0.0.4 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/olekukonko/tablewriter v0.0.1
	github.com/onsi/ginkgo v1.8.0 // indirect
	github.com/onsi/gomega v1.5.0 // indirect
	github.com/pelletier/go-toml v1.4.0
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.5
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/stretchr/testify v1.5.1
	github.com/syndtr/goleveldb v1.0.0
	golang.org/x/crypto v0.0.0-20200422194213-44a606286825
	golang.org/x/sync v0.0.0-20190423024810-112230192c58
	golang.org/x/sys v0.0.0-20191228213918-04cbcbbfeed8 // indirect
	golang.org/x/text v0.3.2 // indirect
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0
	google.golang.org/grpc v1.27.0
	google.golang.org/protobuf v1.25.0
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/yaml.v2 v2.2.8 // indirect
)
