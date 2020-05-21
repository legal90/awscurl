# This Dockerfile is supposed to be used by the `goreleaser` tool
# We don't build any go files in the docker build phase
# and just merely copy the binary to a scratch image
FROM scratch
COPY awscurl /bin/awscurl
ENTRYPOINT ["/bin/awscurl"]
