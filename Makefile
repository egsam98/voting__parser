IMAGE = egsam98/voting-parser:latest

build:
	docker build -t $(IMAGE) .
