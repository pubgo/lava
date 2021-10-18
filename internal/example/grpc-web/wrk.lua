wrk.method = "POST"
wrk.body   = [[{"input":"hello"}]]
wrk.headers["Content-Type"] = "grpc+json"