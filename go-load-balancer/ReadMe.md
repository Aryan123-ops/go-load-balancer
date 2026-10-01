# This is a sample directory for load balancer using Round-Robin Algorithm.
# Steps to run this Project.

# Run three server on command prompt 
# first run backend\main.go
# commands- set PORT=9001
#           set SERVER_NAME=server-1
#           go run .\backend\main.go

# second run backend\main.go
# commands- set PORT=9002
#           set SERVER_NAME=server-2
#           go run .\backend\main.go

# three run backend\main.go
# commands- set PORT=9003
#           set SERVER_NAME=server-3
#           go run .\backend\main.go

# Now run loadbalancer\main.go
#       go run .\loadbalancer\main.go

# postman collection to run for backend/main.go:
# for backend/main.go = http://127.0.0.1:9001/health = to check server health.
#                       http://127.0.0.1:9002/health = to check server health.
#                       http://127.0.0.1:9003/health = to check server health.

# sample response after running backend/main.go from http://127.0.0.1:9001/health api  
# {
#    "server": "server-1",
#    "status": "UP"
# }

# postman collection to run for loadbalancer/main.go:
#                       http://localhost:8080/api/users = when this api will be hit it will fetch response from multiple server.

# sample response after running loadbalancer/main.go from http://localhost:8080/api/users api
# {
#    "server": "server-3",
#    "message": "Request handled successfully",
#   "time": "2026-10-01T13:08:04+05:30"
# }
