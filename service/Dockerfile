FROM golang:1.8

RUN go get github.com/go-ozzo/ozzo-dbx 
RUN go get github.com/go-ozzo/ozzo-routing 
RUN go get github.com/golang/gddo/httputil/header
RUN go get github.com/lib/pq 

ADD . /go/src/app
WORKDIR /go/src/app 

RUN go build -o stservice .

RUN go build -o main . 
CMD ["/go/src/app/stservice"]
