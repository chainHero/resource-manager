# Copyright 2018 Antoine CHABERT, toHero.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

.PHONY: all dev clean build env-up env-down run

all: clean build env-up init

dev: build run

##### BUILD
build:
	@echo "Build ..."
	@cd chaincode && dep ensure
	@cd app && dep ensure
	@cd app && go build
	@echo "Build done"

##### ENV
env-up:
	@echo "Start environment ..."
	@cd fixtures && docker-compose up -d
	@echo "Environment up"

env-down:
	@echo "Stop environment ..."
	@cd fixtures && docker-compose down
	@echo "Environment down"

##### RUN
init:
	@echo "Start app and init ..."
	@cd app && ./app -install -register

run:
	@echo "Start app ..."
	@cd app && ./app

##### TEST & LINT
test:
	@echo "Test..."
	@GOCACHE=off go test -cover -v ./...
	@echo "Test done"

lint:
	@echo "Lint..."
	@gometalinter --vendor --deadline=8m --exclude=app/vendor --exclude=chaincode/vendor --cyclo-over=15 ./...
	@echo "Lint done"

##### CLEAN
clean: env-down
	@echo "Clean up ..."
	@rm -rf /tmp/chainhero-* app/app chaincode/chaincode
	@docker rm -f -v `docker ps -a --no-trunc | grep "chainhero-resource-manager" | cut -d ' ' -f 1` 2>/dev/null || true
	@docker rmi `docker images --no-trunc | grep "chainhero-resource-manager" | cut -d ' ' -f 1` 2>/dev/null || true
	@echo "Clean up done"
