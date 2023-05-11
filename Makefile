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

setupui:
	npm --prefix ./ui install

ui:
	npm --prefix ./ui run build

uidev:
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

.PHONY: wab ui full uidev test api setupui lint
