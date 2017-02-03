FROM docker:1.13-dind

ADD drone-gcr /bin/
ENTRYPOINT ["/usr/local/bin/dockerd-entrypoint.sh", "/bin/drone-gcr"]
