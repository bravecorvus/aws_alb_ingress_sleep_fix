# AWS EKS ALB + Ingress Downtime Fix

While my organization appreciates the serverless properties of running in EKS Fargate, we needed to change the service discovery model to use Application Load Balancers and Kubernetes Ingress objects.

In using this over our old Classic Load Balancer + Service set up, we could not get our Green/Blue rollouts to work properly on low pod instance deployments (we were seeing Gateway 504 HTTP Status codes from 20 seconds to around 2 minutes). After talking with AWS Support, they helped us understand that this was not an EKS specific issue, but a problem with Kubernetes in general when Ingress decides to re-route traffic from the old pod to the new pod (think of polling). The Ingress controller will continue to forward traffic to the old pod despite being destroyed leading to some downtime.

While Talking to AWS Support, they offered the following patch in the pod specification until this issue get's resolved in the upstream Kubernetes project.

```
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-name
  labels:
    app: app-name
spec:
  replicas: 1
  selector:
    matchLabels:
      app: pod-name
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 0
      maxSurge: 1
  template:
    metadata:
      labels:
        app: app-name
    spec:
      containers:
      - name: container-name
        image: container-name:container-version
        imagePullPolicy: IfNotPresent
        lifecycle:
          preStop:
            exec:
              command: ["/bin/sh", "-c", "sleep 60"]
      terminationGracePeriodSeconds: 70
```

> The fix is the last 5 lines.

The idea is to keep traffic flowing the old container for an extra 1 minute before terminating it completely.

However, many of the containers I have created for my organization has been built using Docker Scratch (due to the small image sizes and added security through reduced attack surface). For containers that have a shell and the `sleep` command, the below container is unecessary.

I wrote a very simple Go executable to do the equivalent of `/bin/sh -c "sleep 60"` without needing a shell, nor the `sleep` command.

The way to use is as follows:

## Dockerfile (must use multi-stage-build compatible version of Docker)
```
# Used for sleep docker container
FROM golang:1.15-alpine as sleep-build-env
RUN apk --no-cache add git
WORKDIR /
RUN git clone https://github.com/gilgameshskytrooper/aws_alb_ingress_sleep_fix.git sleep
WORKDIR /sleep
RUN GO111MODULE=on CGO_ENABLED=0 GOOS=linux go build


....
FROM scratch
...
COPY --from=sleep-build-env /sleep/sleep /sleep
```

## Podspec

```
...
        ports:
        - containerPort: 8080
        lifecycle:
          preStop:
            exec:
              command: ["/sleep", "60"]
      terminationGracePeriodSeconds: 70
```
