.DEFAULT_GOAL := help

# Determine this makefile's path.
# Be sure to place this BEFORE `include` directives, if any.
PKG := github.com/natemarks/puppers
DEFAULT_BRANCH := main
VERSION := 0.0.10
COMMIT := $(shell git rev-parse HEAD)
SHELL := $(shell which bash)
CDIR = $(shell pwd)
CURRENT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
DEFAULT_BRANCH := main
EXECUTABLES := puppers pupperswebserver
GOOS := linux darwin
GOARCH := amd64


help: ## Show this help.
	@fgrep -h "##" $(MAKEFILE_LIST) | fgrep -v fgrep | sed -e 's/\\$$//' | sed -e 's/##//'

clean-venv: ## re-create virtual env
	rm -rf .venv
	python3 -m venv .venv
	( \
       source .venv/bin/activate; \
       pip install --upgrade pip setuptools; \
	   pip install -r requirements.txt; \
    )

unittest: ## run test that don't require deployed resources
	go test -v ./... -tags unit

${EXECUTABLES}:
	@for o in $(GOOS); do \
	  for a in $(GOARCH); do \
        echo "$(COMMIT)/$${o}/$${a}" ; \
        mkdir -p build/$(COMMIT)/$${o}/$${a} ; \
        echo "VERSION: $(VERSION)" > build/$(COMMIT)/$${o}/$${a}/version.txt ; \
        echo "COMMIT: $(COMMIT)" >> build/$(COMMIT)/$${o}/$${a}/version.txt ; \
        env GOOS=$${o} GOARCH=$${a} \
        go build  -v -o build/$(COMMIT)/$${o}/$${a}/$@ ${PKG}/cmd/$@; \
	  done \
    done ; \

write_commit: git-status ## create version.txt for go:embed
	echo $(COMMIT) > version.txt

write_version: git-status ## create version.txt for go:embed
ifneq ($(CURRENT_BRANCH), $(DEFAULT_BRANCH))
	$(error Not on branch $(DEFAULT_BRANCH))
else
	echo $(VERSION) > version.txt
endif

build: write_commit ${EXECUTABLES}
	-rm -rf build/current
	mkdir -p build
	ln -s $(CDIR)/build/$(COMMIT) $(CDIR)/build/current

s3_upload: ## upload commit build to S3 bucket for testing
		zip -r puppers_$(COMMIT).zip build/$(COMMIT)
		aws s3 cp puppers_$(COMMIT).zip "s3://$(S3_BUCKET)/puppers/"

release: git-status write_version ${EXECUTABLES}
ifneq ($(CURRENT_BRANCH), $(DEFAULT_BRANCH))
	$(error Not on branch $(DEFAULT_BRANCH))
else
	-rm -rf build/current
	mkdir -p build
	ln -s $(CDIR)/build/$(COMMIT) $(CDIR)/build/current
	mkdir -p release/$(VERSION)
	@for o in $(GOOS); do \
		for a in $(GOARCH); do \
				tar -C ./build/$(COMMIT)/$${o}/$${a} -czvf release/$(VERSION)/puppers_$(VERSION)_$${o}_$${a}.tar.gz . ; \
		done \
		done ; \
		
endif

deploymenttest: ##  run all tests
	go test -v ./...

static: write_commit lint ## run fmt, vet, goimports, gocyclo
	( \
			 gofmt -w  -s .; \
			 test -z "$$(go vet ./...)"; \
			 go install golang.org/x/tools/cmd/goimports@latest; \
			 goimports -w .; \
			 go install github.com/fzipp/gocyclo/cmd/gocyclo@latest; \
			 test -z "$$(gocyclo -over 25 .)"; \
			 go install honnef.co/go/tools/cmd/staticcheck@latest ; \
			 staticcheck ./... ; \
    )

lint:  ##  run golint
	( \
			 go install golang.org/x/lint/golint@latest; \
			 golint ./...; \
			 test -z "$$(golint ./...)"; \
    )

bump: git-status clean-venv  ## bump version in main branch
ifeq ($(CURRENT_BRANCH), $(DEFAULT_BRANCH))
	( \
	   source .venv/bin/activate; \
	   pip install bump2version; \
	   bump2version $(part); \
	)
else
	@echo "UNABLE TO BUMP - not on Main branch"
	$(info Current Branch: $(CURRENT_BRANCH), main: $(DEFAULT_BRANCH))
endif


git-status: ## require status is clean so we can use undo_edits to put things back
	@status=$$(git status --porcelain); \
	if [ ! -z "$${status}" ]; \
	then \
		echo "Error - working directory is dirty. Commit those changes!"; \
		exit 1; \
	fi

docker-build: git-status ## create docker image with commit tag
	( \
		aws ecr get-login-password --region us-east-1 | docker login \
		--username AWS \
		--password-stdin 709310380790.dkr.ecr.us-east-1.amazonaws.com; \
	   docker build \
       	-t puppers:$(COMMIT) \
       	-t 709310380790.dkr.ecr.us-east-1.amazonaws.com/puppers:$(COMMIT) \
       	-f docker/Dockerfile .; \
       	docker push 709310380790.dkr.ecr.us-east-1.amazonaws.com/puppers:$(COMMIT); \
	)


docker-release: git-status ## create docker image with release version tag
	( \
		aws ecr get-login-password --region us-east-1 | docker login \
		--username AWS \
		--password-stdin 709310380790.dkr.ecr.us-east-1.amazonaws.com; \
	   docker build \
       	-t puppers:$(VERSION) \
       	-t 709310380790.dkr.ecr.us-east-1.amazonaws.com/puppers:$(VERSION) \
       	-f docker/Dockerfile .; \
       	docker push 709310380790.dkr.ecr.us-east-1.amazonaws.com/puppers:$(VERSION); \
	)

deploy:  ## deploy puppers
	( \
    	   source .venv/bin/activate; \
		   cd deployments/cdk; \
    	   cdk deploy --all; \
    	)

destroy:  ## deploy puppers
	( \
    	   source .venv/bin/activate; \
		   cd deployments/cdk; \
    	   cdk destroy --all; \
    	)

print-%  : ; @echo $($*)

pylint: ## run pylint on python files
	( \
       . .venv/bin/activate; \
       git ls-files '*.py' | xargs pylint --max-line-length=90; \
    )

black: ## use black to format python files
	( \
       . .venv/bin/activate; \
       git ls-files '*.py' | xargs black; \
    )

.PHONY: build static test	
