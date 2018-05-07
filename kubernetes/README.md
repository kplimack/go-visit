Kubernetes Example
==================

This is a simple visit counting app that tracks a incrementing counter in redis
as it is visited. It serves as a good example for getting started with
Kubernetes since it deals with deploying an application with dependencies.

# Terms

- *Master*: the name of the controlling class of nodes in a kubernetes cluster.
- *Minion*: the name of the worker class of nodes in a kubernetes cluster.
- *Namespace*: a logical grouping of kubernetes objects. For these examples only
  the `default` namespace will be used.
- *Pod*: the base unit of work in kubernetes. A *pod* is composed of one or more
  containers that run on the same kubernetes minion and share storage and
  network.
- *Deployment*: an object type in kubernetes for representing a deployment of a
  pod. This is the best option for deploying applications.
- *Config Map*: an object type in kubernetes for storing data to be used by other
  objects.

# Setting up a development environment

- [Docker edge for mac](https://docs.docker.com/docker-for-mac/edge-release-notes/) with the kubernetes option enabled.
- [Minikube](https://kubernetes.io/docs/getting-started-guides/minikube/)

# Creating a [Pod](https://kubernetes.io/docs/concepts/workloads/pods/pod/)

Example YAML config for a simple pod running redis:

```
---
apiVersion: v1
kind: Pod
metadata:
  name: redis
spec:
  containers:
  - name: redis
    image: redis
    ports:
    - containerPort: 6379
```

Create the pod using this command:

```
$ kubectl apply -f kubernetes/redis-pod.yaml
pod "redis" created
```

While the pod is creating you might see this:

```
$ kubectl get pods
NAME      READY     STATUS              RESTARTS   AGE
redis     0/1       ContainerCreating   0          <invalid>
```

Eventually you should see something like this:

```
$ kubectl get pods
NAME      READY     STATUS    RESTARTS   AGE
redis     1/1       Running   0          1m
```

To see the logs from a running pod, use the `kubectl logs` command:

```
$ kubectl logs redis
1:C 03 May 00:13:05.848 # oO0OoO0OoO0Oo Redis is starting oO0OoO0OoO0Oo
1:C 03 May 00:13:05.849 # Redis version=4.0.9, bits=64, commit=00000000, modified=0, pid=1, just started
1:C 03 May 00:13:05.849 # Warning: no config file specified, using the default config. In order to specify a config file use redis-server /path/to/redis.conf
1:M 03 May 00:13:05.852 * Running mode=standalone, port=6379.
1:M 03 May 00:13:05.852 # WARNING: The TCP backlog setting of 511 cannot be enforced because /proc/sys/net/core/somaxconn is set to the lower value of 128.
1:M 03 May 00:13:05.852 # Server initialized
1:M 03 May 00:13:05.852 # WARNING you have Transparent Huge Pages (THP) support enabled in your kernel. This will create latency and memory usage issues with Redis. To fix this issue run the command 'echo never > /sys/kernel/mm/transparent_hugepage/enabled' as root, and add it to your /etc/rc.local in order to retain the setting after a reboot. Redis must be restarted after THP is disabled.
1:M 03 May 00:13:05.852 * Ready to accept connections
```

To access a port on the pod (either using minikube or docker for mac native) you use the `port-forward`
subcommand.

```
$ kubectl port-forward redis 56379:6379
Forwarding from 127.0.0.1:56379 -> 6379
```

This example shows mapping a port on the host to a port on the container
(the format is HOST:CONTAINER). The mapping is optional, but there was already
a redis running on my machine so I needed to remap it.

Now redis will be available on that port.

```
$ redis-cli -h localhost -p 56379
localhost:56379> info
# Server
redis_version:4.0.9
```

We don't actually want to create a pod like this, since we are almost always
going to be managing a deployed application. Let's destroy this pod.

```
$ kubectl delete -f kubernetes/redis-pod.yaml
pod "redis" deleted
$ kubectl get pods
No resources found.
```

# Creating a [deployment](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/) for redis

We would rather manage a deployment than a vanilla pod. A deployment opens us
up to better options for managing the lifecycle of an application. With a
deployment we can run multiple instances of the same pod without managing them
separately. Also, it makes it easier to upgrade versions of an application
with a deployment.

Example deployment yaml for a redis server:

```
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: redis
spec:
  replicas: 1
  selector:
    matchLabels:
      app: redis
  template:
    metadata:
      labels:
        app: redis
    spec:
      containers:
      - name: redis
        image: redis
        ports:
        - containerPort: 6379
```

The labels here are used internally to locate the pods and associate them with
this deployment. These need to be unique across a namespace.

We can create this deployment just like we created pods earlier:

```
$ kubectl apply -f kubernetes/redis-deployment.yaml
deployment "redis" created
```

This will create a few resources. To see them all run the following:

```
$ kubectl get all
NAME           DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
deploy/redis   1         1         1            1           6s

NAME                  DESIRED   CURRENT   READY     AGE
rs/redis-6c6df5bbc6   1         1         1         6s

NAME           DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
deploy/redis   1         1         1            1           6s

NAME                  DESIRED   CURRENT   READY     AGE
rs/redis-6c6df5bbc6   1         1         1         6s

NAME                        READY     STATUS    RESTARTS   AGE
po/redis-6c6df5bbc6-cbrkh   1/1       Running   0          6s

NAME             TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)   AGE
svc/kubernetes   ClusterIP   10.96.0.1    <none>        443/TCP   26m
```

Note that deleting a deployment is the same as deleting a pod, and all
children objects will be removed as well.

```
$ kubectl delete -f kubernetes/redis-deployment.yaml
deployment "redis" deleted
$ kubectl get all
NAME             TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)   AGE
svc/kubernetes   ClusterIP   10.96.0.1    <none>        443/TCP   24m
```

*Remember to bring the deployment back up if you deleted it.*

Connecting to the redis is similar to above, but the pod names are made unique
by kubernetes. Use the full pod name to port-forward.

```
$ kubectl port-forward redis-6c6df5bbc6-pnq4q 56379:6379
Forwarding from 127.0.0.1:56379 -> 6379
```

# Creating a [service](https://kubernetes.io/docs/concepts/services-networking/service/) for redis


We will need a service definintion so that other kubernetes pods can
communicate with this redis service. This will set up a DNS name for
the redis deployment.

```
---
apiVersion: v1
kind: Service
metadata:
  name: redis
  labels:
    app: redis
spec:
  ports:
  - port: 6379
    targetPort: 6379
  selector:
    app: redis
```

To test that this service exists, we can spin up a one-off container
and query dns.

```
$ kubectl run -it --rm testdummycontainername --image alpine -- sh
If you don't see a command prompt, try pressing enter.
/ # nslookup redis
nslookup: can't resolve '(null)': Name does not resolve

Name:      redis
Address 1: 10.104.66.48 redis.default.svc.cluster.local
```

Or attach to the redis instance using redis-cli.

```
/ # apk --update add redis
fetch http://dl-cdn.alpinelinux.org/alpine/v3.7/main/x86_64/APKINDEX.tar.gz
fetch http://dl-cdn.alpinelinux.org/alpine/v3.7/community/x86_64/APKINDEX.tar.gz
(1/1) Installing redis (4.0.6-r0)
Executing redis-4.0.6-r0.pre-install
Executing busybox-1.27.2-r7.trigger
OK: 7 MiB in 12 packages
/ # redis-cli -h redis
redis:6379> get visit.count
(nil)
```

# Creating a [deployment](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/) for the visit application

This is exactly the same as the redis example, with the addition
of some environment config.

```
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: visit
spec:
  selector:
    matchLabels:
      app: visit
  replicas: 1
  template:
    metadata:
      labels:
        app: visit
    spec:
      containers:
      - name: visit
        image: partkyle/go-visit:0.5.2
        env:
          - name: VISIT_HOST
            value: "0.0.0.0"
          - name: VISIT_PORT
            value: "80"
          - name: VISIT_REDISADDR
            value: "redis:6379"
          - name: VISIT_REDISKEY
            value: "visit.count"
        ports:
        - containerPort: 80
```

Use apply to create the deployment & service.

```
$ kubectl apply -f kubernetes/visit-deployment.yaml -f kubernetes/visit-service.yaml
deployment "visit" created
service "visit" created
```

Now the service should be available inside the kubernetes cluster
under the name "visit".

```
/ # curl visit
The current visit count is 1 on visit-77c696dc89-xw279 running version 0.5.2.
```

Let's say we want to scale this app. We can change the amount of
`replicas`.

```
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: visit
spec:
  selector:
    matchLabels:
      app: visit
  replicas: 3
  template:
    metadata:
      labels:
        app: visit
    spec:
      containers:
      - name: visit
        image: partkyle/go-visit:0.5.2
        env:
          - name: VISIT_HOST
            value: "0.0.0.0"
          - name: VISIT_PORT
            value: "80"
          - name: VISIT_REDISADDR
            value: "redis:6379"
          - name: VISIT_REDISKEY
            value: "visit.count"
        ports:
        - containerPort: 80
```

Apply the new config, and kubernetes will start running multiple versions
of the app.

```
$ kubectl apply -f kubernetes/visit-deployment.yaml
deployment "visit" configured
$ kubectl get pods
NAME                                      READY     STATUS    RESTARTS   AGE
visit-77c696dc89-djb48                    1/1       Running   0          1s
visit-77c696dc89-xw279                    1/1       Running   0          1h
visit-77c696dc89-zd52c                    1/1       Running   0          1s
```

The `visit` service should now balance between the 3 pods.

```
/ # curl visit
The current visit count is 7 on visit-77c696dc89-xw279 running version 0.5.2.
/ # curl visit
The current visit count is 8 on visit-77c696dc89-zd52c running version 0.5.2.
/ # curl visit
The current visit count is 9 on visit-77c696dc89-zd52c running version 0.5.2.
/ # curl visit
The current visit count is 10 on visit-77c696dc89-zd52c running version 0.5.2.
/ # curl visit
The current visit count is 11 on visit-77c696dc89-djb48 running version 0.5.2.
```

# Using a [configmap](https://kubernetes.io/docs/tasks/configure-pod-container/configure-pod-configmap/) to make a reusable deployment definition

This deployment definition is nice, but we can't reuse it across environments
since the `VISIT_REDISADDR` attribute might change. We can use a configmap to
store the config outside of the deployment definition, allowing us to
run different configurations in different environments without having to
save multiple copies of the same deployment file.

```
---
apiVersion: v1
kind: ConfigMap
data:
  VISIT_HOST: "0.0.0.0"
  VISIT_PORT: "80"
  VISIT_REDISADDR: "redis:6379"
  VISIT_REDISKEY: "visit.count"
metadata:
  name: visit
```

Applying the config to kubernetes.

```
$ kubectl apply -f kubernetes/visit-configmap.yaml
configmap "visit" created
$ kubectl describe configmap visit
Name:         visit
Namespace:    default
Labels:       <none>
Annotations:  kubectl.kubernetes.io/last-applied-configuration={"apiVersion":"v1","data":{"VISIT_HOST":"0.0.0.0","VISIT_PORT":"80","VISIT_REDISADDR":"redis-master:6379","VISIT_REDISKEY":"visit.count"},"kind":"Confi...

Data
====
VISIT_HOST:
----
0.0.0.0
VISIT_PORT:
----
80
VISIT_REDISADDR:
----
redis-master:6379
VISIT_REDISKEY:
----
visit.count
Events:  <none>
```

And change the deployment definition to

```
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: visit
spec:
  selector:
    matchLabels:
      app: visit
  replicas: 3
  template:
    metadata:
      labels:
        app: visit
    spec:
      containers:
      - name: visit
        image: partkyle/go-visit:0.5.2
        envFrom:
        - configMapRef:
            name: visit
        ports:
        - containerPort: 80
```

# Liveness And Ready checks

- Liveness checks tell kube whether or not it needs to restart the application
- Readiness checks tell kube whether or not a service is ready to accept connections

```
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: visit
spec:
  selector:
    matchLabels:
      app: visit
  replicas: 3
  template:
    metadata:
      labels:
        app: visit
    spec:
      containers:
      - name: visit
        image: partkyle/go-visit:0.5.1
        envFrom:
        - configMapRef:
            name: visit
        livenessProbe:
          httpGet:
            path: /health
            port: 80
          initialDelaySeconds: 5
          timeoutSeconds: 1
          periodSeconds: 10
          failureThreshold: 3
        ports:
        - containerPort: 80

```
