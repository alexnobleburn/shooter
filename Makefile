.PHONY: build_get_content_access
build_get_content_access:
	CGO_ENABLED=0 go build -ldflags="-extldflags '-O2'" -o bin/shooter_get_content_access ./cmd/get_content_access.go
build_get_content_access_linux:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-extldflags '-O2'" -o bin/shooter_get_content_access_linux ./cmd/get_content_access.go
copy_get_content_access_linux:
	scp /Users/e.ishutkin/GolandProjects/vkgo/projects/donut/shooter/bin/shooter_get_content_access_linux eishutkin@adm512:/home/eishutkin/sh

.PHONY: build_get_accessible_owners
build_get_accessible_owners:
	CGO_ENABLED=0 go build -ldflags="-extldflags '-O2'" -o bin/shooter_get_accessible_owners ./cmd/get_accessible_owners.go
build_get_accessible_owners_linux:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-extldflags '-O2'" -o bin/shooter_get_accessible_owners_linux ./cmd/get_accessible_owners.go
copy_get_accessible_owners_linux:
	scp /Users/aleksey.trofimov/GolandProjects/vkgo/projects/donut/shooter/bin/shooter_get_accessible_owners_linux alekseytrofimov@adm512:/home/alekseytrofimov/sh

### --target-rps=1300 --shooting-duration=30m --avg-timing=1s --stats-period=15s --actor-id 17142
