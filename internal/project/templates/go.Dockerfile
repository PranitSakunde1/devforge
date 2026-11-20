FROM golang:1.23.8
RUN go install github.com/air-verse/air@v1.61.7 \
 && go install gotest.tools/gotestsum@v1.12.2
WORKDIR /workspace
COPY . .
EXPOSE 8080
CMD ["air", "-c", ".air.toml"]
