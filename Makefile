jaegerImage := jaegertracing/all-in-one:1.21
prometheusImage := prom/prometheus:v2.23.0


compile_proto:
	protoc --proto_path=. \
		   --go_out=./internal/pb \
		   --go_opt=paths=source_relative \
		   --go-grpc_out=./internal/pb \
		   --go-grpc_opt=paths=source_relative \
		   ./proto/key_value.proto
run:
	APP_PORT=:8009 \
	LOGGER_TYPE=file \
	TRANSACTION_FILENAME=transaction.log \
	POSTGRES_DBNAME=postgres \
	POSTGRES_HOST=postgres \
	POSTGRES_USER=postgres \
	POSTGRES_PASSWORD=postgres \
	USE_NEW_STORAGE=true \
	go run .
start_jaeger:
	docker kill jaeger || true
	docker rm jaeger || true
	docker run -d --name jaeger \
		-p 16686:16686 \
		-p 14268:14268 \
		${jaegerImage}
start_prometheus:
	docker kill prometheus || true
	docker rm prometheus || true
	docker run -d --name prometheus \
		-p 9090:9090 \
		-v "${PWD}/prometheus.yml:/etc/prometheus/prometheus.yml" \
		--add-host=host.docker.internal:host-gateway \
	 	${prometheusImage}
