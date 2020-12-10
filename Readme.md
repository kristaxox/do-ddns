# do-ddns
A small container to keep DNS records up to date with your external IP.

## Build
```
docker build -t <tag>/do-ddns:latest .
```

## Docker image
https://hub.docker.com/repository/docker/kristaxox/do-ddns

## Example deployment in k8s
I personally run a k3s node/cluster VM in my rack and run a few of these deployments for each of my domains.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: do-ddns
  namespace: ddns
  labels:
    app: do-ddns
spec:
  replicas: 1
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 1
      maxSurge: 0
  selector:
    matchLabels:
      app: do-ddns
  template:
    metadata:
      labels:
        app: do-ddns
    spec:
      containers:
        - name: do-ddns
          image: kristaxox/do-ddns:latest
          imagePullPolicy: Always
          env:
            - name: DO_TOKEN
              value: "<digital ocean token"
            - name: RECORDS
              value: "example.com,dev.example.com" # example.com = @, dev.example.com =dev
            - name: DOMAIN
              value: "example.com"
            - name: FREQUENCY
              value: "5m"
```