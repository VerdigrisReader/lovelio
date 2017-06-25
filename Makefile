build:
	docker build --rm -t lovelio .

run-dev:
	docker run -d -p 3000:3000 -v `pwd`:/go/src/app --name dev --rm lovelio

stop-dev:
	docker stop dev
