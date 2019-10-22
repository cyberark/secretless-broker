run_secretless:
	./reflex.sh
backend:
	docker run --rm -e 'ACCEPT_EULA=Y' -e 'SA_PASSWORD=yourStrong()Password' -p 1433:1433 -d mcr.microsoft.com/mssql/server:2017-latest
client_with_secretless:
	sqlcmd -S "127.0.0.1,2223" -U x -P "x" -Q "select 1, 'abc'"
client:
	sqlcmd -S "127.0.0.1,1433" -U sa -P "yourStrong()Password" -Q "select 1, 'abc'"
