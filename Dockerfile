#Stage 1 build and test
#docker.io prefix required by podman
FROM docker.io/golang:alpine as builder
RUN mkdir /build
WORKDIR /build
COPY go.mod .
COPY go.sum .

# TODO this doesn't work with podman 3.x but does with 4.x
#RUN --mount=type=cache,target=/root/.cache go mod download
RUN go mod download
COPY . .
#RUN --mount=type=cache,target=/root/.cache make build
RUN apk --no-cache add gcc build-base git
RUN make build

# test that that the build is good and app launches
RUN /build/bin/previewd version
# values required by unit / integration tests

#ARG ARG_TEST_JEKPREV_REPO_NOAUTH
#ARG ARG_TEST_JEKPREV_LOCALDIR
#ARG ARG_TEST_JEKPREV_BRANCHSWITCH
#ARG ARG_TEST_JEKPREV_SECURE_REPO_NOAUTH
#ARG ARG_TEST_JEKPREV_SECURECLONEPW
#
#ENV TEST_JEKPREV_REPO_NOAUTH $ARG_TEST_JEKPREV_REPO_NOAUTH
#ENV TEST_JEKPREV_LOCALDIR $ARG_TEST_JEKPREV_LOCALDIR
#ENV TEST_JEKPREV_BRANCHSWITCH $ARG_TEST_JEKPREV_BRANCHSWITCH
#ENV TEST_JEKPREV_SECURE_REPO_NOAUTH $ARG_TEST_JEKPREV_SECURE_REPO_NOAUTH
#ENV TEST_JEKPREV_SECURECLONEPW $ARG_TEST_JEKPREV_SECURECLONEPW

#RUN go test -v

# generate clean, final image for end users
FROM alpine:3.11.3
RUN apk update
RUN apk add git
COPY --from=builder /build/bin/previewd .

# executable
ENTRYPOINT [ "./previewd" ]
CMD ["testserver"]
# arguments that can be override
#CMD [ "3", "300" ]
