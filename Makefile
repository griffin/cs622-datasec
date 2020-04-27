run:
	go run cmd/api-server/main.go --sql_username=postgres --sql_database=postgres --sql_host=localhost --sql_password=qwerty123

tunnel:
	ssh -f ubuntu@54.210.93.153 -L 5432:cs.ccvuktxhltfq.us-east-1.rds.amazonaws.com:5432 -N
