local:
	CGO_ENABLED=0 GOOS=linux go build -o crawler -a -installsuffix cgo .
run:
	docker build -t crawler . && docker run -p 8080:8080 crawler
