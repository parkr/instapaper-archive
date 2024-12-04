FROM golang:1.23.4

WORKDIR /srv/instapaper-archive
COPY . .
RUN go install github.com/parkr/instapaper-archive

ENTRYPOINT [ "instapaper-archive" ]
