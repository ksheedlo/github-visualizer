.PHONY: all clean test
all: js go

dashboard/bundle.js: dashboard/index.js
	browserify dashboard/index.js -o dashboard/bundle.js

dashboard/bundle.min.js: dashboard/bundle.js
	java -jar compiler.jar --js dashboard/bundle.js --js_output_file dashboard/bundle.min.js -O SIMPLE -W QUIET

js: dashboard/bundle.min.js

ghviz: main.go errors/*.go github/*.go
	go build

go: ghviz

clean:
	rm ghviz dashboard/bundle.js dashboard/bundle.min.js

test:
	go test github.com/ksheedlo/ghviz/github
