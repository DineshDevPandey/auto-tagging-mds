
.PHONY: all

# Original make file
all: build

build:
	GOOS=linux GOARCH=amd64 $(MAKE) service_index
	GOOS=linux GOARCH=amd64 $(MAKE) service_show
	GOOS=linux GOARCH=amd64 $(MAKE) service_update
	GOOS=linux GOARCH=amd64 $(MAKE) service_delete
	GOOS=linux GOARCH=amd64 $(MAKE) service_create

	GOOS=linux GOARCH=amd64 $(MAKE) company_index
	GOOS=linux GOARCH=amd64 $(MAKE) company_show
	GOOS=linux GOARCH=amd64 $(MAKE) company_update
	GOOS=linux GOARCH=amd64 $(MAKE) company_delete
	GOOS=linux GOARCH=amd64 $(MAKE) company_create

	GOOS=linux GOARCH=amd64 $(MAKE) tag_index
	GOOS=linux GOARCH=amd64 $(MAKE) tag_show
#	GOOS=linux GOARCH=amd64 $(MAKE) tag_update
	GOOS=linux GOARCH=amd64 $(MAKE) tag_delete
	GOOS=linux GOARCH=amd64 $(MAKE) tag_create

#	GOOS=linux GOARCH=amd64 $(MAKE) rule_index
#	GOOS=linux GOARCH=amd64 $(MAKE) rule_show
#	GOOS=linux GOARCH=amd64 $(MAKE) rule_update
#	GOOS=linux GOARCH=amd64 $(MAKE) rule_delete
#	GOOS=linux GOARCH=amd64 $(MAKE) rule_create

# service
service_index: ./api/service/index/main.go
	go build -o ./api/service/index/index ./api/service/index

service_show: ./api/service/show/main.go
	go build -o ./api/service/show/show ./api/service/show

service_create: ./api/service/create/main.go
	go build -o ./api/service/create/create ./api/service/create

service_update: ./api/service/update/main.go
	go build -o ./api/service/update/update ./api/service/update

service_delete: ./api/service/delete/main.go
	go build -o ./api/service/delete/delete ./api/service/delete

# company
company_index: ./api/company/index/main.go
	go build -o ./api/company/index/index ./api/company/index

company_show: ./api/company/show/main.go
	go build -o ./api/company/show/show ./api/company/show

company_create: ./api/company/create/main.go
	go build -o ./api/company/create/create ./api/company/create

company_update: ./api/company/update/main.go
	go build -o ./api/company/update/update ./api/company/update

company_delete: ./api/company/delete/main.go
	go build -o ./api/company/delete/delete ./api/company/delete

# tag
tag_index: ./api/tag/index/main.go
	go build -o ./api/tag/index/index ./api/tag/index

tag_show: ./api/tag/show/main.go
	go build -o ./api/tag/show/show ./api/tag/show

tag_create: ./api/tag/create/main.go
	go build -o ./api/tag/create/create ./api/tag/create

tag_update: ./api/tag/update/main.go
	go build -o ./api/tag/update/update ./api/tag/update

tag_delete: ./api/tag/delete/main.go
	go build -o ./api/tag/delete/delete ./api/tag/delete

# rule
rule_index: ./api/rule/index/main.go
	go build -o ./api/rule/index/index ./api/rule/index

rule_show: ./api/rule/show/main.go
	go build -o ./api/rule/show/show ./api/rule/show

rule_create: ./api/rule/create/main.go
	go build -o ./api/rule/create/create ./api/rule/create

rule_update: ./api/rule/update/main.go
	go build -o ./api/rule/update/update ./api/rule/update

rule_delete: ./api/rule/delete/main.go
	go build -o ./api/rule/delete/delete ./api/rule/delete
