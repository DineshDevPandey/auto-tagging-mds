# auto tagging with mds

This repository provides REST API with the data in Modern Data Stack - Everything that you need to know ! | Modern Data Stack so that clients manage modern data stack (MDS) services with user defined tag and rules.

There are four entities - Service, Tag, Rule, and Company.

To manage them we have following APIs -

Service	POST    : http://127.0.0.1:3000/api/v1/services
Service	GET ALL : http://127.0.0.1:3000/api/v1/services
Service	GET     : http://127.0.0.1:3000/api/v1/services/{service_name}
Service	PUT     : http://127.0.0.1:3000/api/v1/services/{service_uuid}
Service	DELETE  : http://127.0.0.1:3000/api/v1/services/{service_name}

Company	POST    : http://127.0.0.1:3000/api/v1/companies
Company	GET ALL : http://127.0.0.1:3000/api/v1/companies
Company	GET     : http://127.0.0.1:3000/api/v1/companies/{company_name}
Company	PUT     : http://127.0.0.1:3000/api/v1/companies/{company_uuid}
Company	DELETE  : http://127.0.0.1:3000/api/v1/companies/{company_name}

Tag	POST        : http://127.0.0.1:3000/api/v1/tags
Tag	GET ALL     : http://127.0.0.1:3000/api/v1/tags
Tag	GET         : http://127.0.0.1:3000/api/v1/tags/{key}/{value}
Tag	PUT         : http://127.0.0.1:3000/api/v1/tags/{key}/{value}
Tag	DELETE      : http://127.0.0.1:3000/api/v1/tags/{key}/{value}
	
Rule POST       : http://127.0.0.1:3000/api/v1/rules
Rule GET ALL    : http://127.0.0.1:3000/api/v1/rules
Rule GET        : http://127.0.0.1:3000/api/v1/rules/{rule_uuid}
Rule PUT        : http://127.0.0.1:3000/api/v1/rules/{rule_uuid}
Rule DELETE     : http://127.0.0.1:3000/api/v1/rules/{rule_uuid}

This is a sample template for hello-world-sam - Below is a brief explanation of what we have generated for you:

```bash
.
├── Makefile                <----------- make to automate build
├── README.md
├── api
│   ├── company             <----------- CRUD API for company 
│   ├── rule                <----------- CRUD API for rule 
│   ├── service             <----------- CRUD API for service 
│   └── tag                 <----------- CRUD API for tag 
├── buildspec.yml
├── database
│   ├── database.go         <----------- DB operation interface
│   ├── dynamodb
│   │   └── dynamodb.go     <----------- DB operation
│   └── models
│       └── models.go       <----------- models for entities 
├── go.mod
├── go.sum
├── streams
│   └── main.go             <----------- DynamoDB stream handler
├── template.yaml           <----------- SAM to handle resources
└── utils
    └── helper.go           <----------- Helper functions
```

## Requirements for local development

* AWS CLI already configured with Administrator permission
* [Docker installed](https://www.docker.com/community-edition)
* [Golang](https://golang.org)
* SAM CLI - [Install the SAM CLI](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/serverless-sam-cli-install.html)

## Steps to execute

Build project
 
```shell
make
```

**Invoking function locally through local API Gateway**

```bash
sam local start-api
```
