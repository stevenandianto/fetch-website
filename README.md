# fetch-website
## How to use
### Using Go Command
```
  go run main.go https://www.google.com https://autify.com
```

### Using Docker
1. Build docker
```
docker build -t go-fetch:1.23 .
```
2. Run
```
 docker run --rm go-fetch:1.23 https://www.google.com https://autify.com
```

