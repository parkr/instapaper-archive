FROM golang:1.26rc2

WORKDIR /srv/instapaper-archive
COPY . .
RUN go install github.com/parkr/instapaper-archive

ENTRYPOINT [ "instapaper-archive" ]
