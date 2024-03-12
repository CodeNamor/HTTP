module github.com/CodeNamor/http

go 1.20

require (
	github.com/codenamor/custom-logging/custom_logging v0.0.0
	github.com/go-kit/kit v0.12.0
	github.com/golang/mock v1.6.0
	github.com/google/uuid v1.3.1
	github.com/gorilla/mux v1.8.1
	github.com/kr/pretty v0.3.1
	github.com/pkg/errors v0.9.1
	github.com/sethgrid/pester v1.2.0
	github.com/sirupsen/logrus v1.9.3
	github.com/stretchr/testify v1.9.0
)

require (
	github.com/VividCortex/gohistogram v1.0.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rogpeppe/go-internal v1.9.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	github.com/codenamor/custom-logging/custom_logging => ../custom_logging
	github.com/codenamor/utilities => ../utilities
)
