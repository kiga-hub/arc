.PHONY: image
image:
	docker build -t common:dev .

.PHONY: run
run:
	docker run --rm common:dev
	