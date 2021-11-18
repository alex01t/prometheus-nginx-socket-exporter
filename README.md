
    docker build . -t r
    docker run -it -p8080:8080 -p9080:9080 r
    
nginx -- http://localhost:8080/debian/doc/

metrics -- http://localhost:9080/metrics

