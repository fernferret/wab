all: full

# Standard build command used in a few places NOTE: right now ther's no way to
# not re-generrate the swagger files, i might change this later but it's so fast
# to compile them now and not annoying
stdbuild = CGO_ENABLED=0 go build -o ./bin/wab ./cmd/wab

noui:
	CGO_ENABLED=0 go build -tags noui -o ./bin/wab ./cmd/wab

debug:
	go build  -gcflags="all=-N -l"  -tags noui -o ./bin/wab ./cmd/wab

api:
	buf generate apis
	buf generate --template buf.gen.tags.yaml apis

# Overall generation function, does everything.
generate: api
	$(swagger)
	$(swagger_fmt)

fast:
	$(stdbuild)

full: ui
	$(stdbuild)

proto:
	bash -c "\
	protoc \
		-I proto \
		--go_out=./gen/greeterpb --go_opt=paths=source_relative \
    --go-grpc_out=./gen/greeterpb --go-grpc_opt=paths=source_relative \
		--grpchan_out=paths=source_relative:./gen/greeterpb \
    proto/greeter.proto && \
	protoc -I proto --plugin=./ui/node_modules/.bin/protoc-gen-ts_proto \
		--ts_proto_opt=outputClientImpl=grpc-web \
		--ts_proto_opt=esModuleInterop=true --ts_proto_out=./ui/src/gen proto/greeter.proto"

# NOTE: I added  "importsNotUsedAsValues": "remove", to tsconfig.app.json to fix:
#   https://github.com/stephenh/ts-proto/issues/594
#
# The below sed was how I had originally hacked it together.
# 	sed -I "" -r 's/import \{ Observable \} from "rxjs";/import type \{ Observable \} from "rxjs";/g' ui/src/gen/greeter.ts"

setupui:
	npm --prefix ./ui install

ui:
	npm --prefix ./ui run build

dev-ui:
	npm --prefix ./ui run dev

test:
	go test -tags noui ./...

ui-test:
	# NOTE: you should run npm install first, this stage does not do this so you get faster tests
	npm --prefix ./ui run test:unit

lint:
	# Lint will exit with code 0, don't use this target in CI/CD.
	golangci-lint run --issues-exit-code=0

clean:
	rm -f ./bin/*
	rm -rf ./ui/dist ./ui/src/gen ./gen

.PHONY: wab ui full uidev test api setupui lint proto
