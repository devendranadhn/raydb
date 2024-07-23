# syntax=docker/dockerfile:1

FROM golang:1.20

RUN apt-get update && apt-get install -y redis-tools
# RUN apt-get update && apt-get install net-tools
# RUN apt-get update && apt-get install make

WORKDIR /raydb

# RUN wget https://download.redis.io/redis-stable.tar.gz
# RUN tar -xzvf redis-stable.tar.gz
# RUN cd redis-stable
# RUN make BUILD_TLS=yes 
COPY go.mod  ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /raydb/ray-server
EXPOSE 7379
CMD ["/ray/ray-server"]
