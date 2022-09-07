FROM golang:1.19.1

WORKDIR /srv/instapaper-archive
COPY . .
RUN go install github.com/parkr/instapaper-archive

ENTRYPOINT [ "instapaper-archive" ]
