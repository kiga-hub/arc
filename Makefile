.PHONY: image
image:
	docker build -t arc:dev .

.PHONY: run
run:
	docker run --rm arc:dev
	