### how to run

    docker build . -t r
    docker run -it -p8080:8080 -p10254:10254 r
    
nginx -- http://localhost:8080/debian/doc/

metrics -- http://localhost:10254/metrics

### get metric list
curl -s 127.0.0.1:10254/metrics |grep -v ^# | cut -d'{' -f1|cut -d' ' -f1 | sort | uniq

### expected metrics
echo
_bytes_sent_bucket
_bytes_sent_count
_bytes_sent_sum
_ingress_upstream_latency_seconds_count
_ingress_upstream_latency_seconds_sum
    _nginx_process_connections
    _nginx_process_connections_total
    _nginx_process_cpu_seconds_total
    _nginx_process_num_procs
    _nginx_process_oldest_start_time_seconds
    _nginx_process_read_bytes_total
    _nginx_process_requests_total
    _nginx_process_resident_memory_bytes
    _nginx_process_virtual_memory_bytes
    _nginx_process_write_bytes_total
_request_duration_seconds_bucket
_request_duration_seconds_count
_request_duration_seconds_sum
_request_size_bucket
_request_size_count
_request_size_sum
_requests
_response_duration_seconds_bucket
_response_duration_seconds_count
_response_duration_seconds_sum
_response_size_bucket
_response_size_count
_response_size_sum
    _success
