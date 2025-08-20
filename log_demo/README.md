```bash
docker build -f Dockerfile.zookeeper -t my-zookeeper .
docker build -f es -t my-es .
docker build -f kafka -t my-kafka .
```

```bash
docker run -d --name zoo1 my-zookeeper
docker run -d --name kafka1 my-kafka
docker run -d --name es1 my-es
```