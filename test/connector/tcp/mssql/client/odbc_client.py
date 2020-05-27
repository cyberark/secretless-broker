#!/usr/bin/env python3
import argparse
import pyodbc
from sys import exit

parser = argparse.ArgumentParser(description='Run an ODBC query.')
parser.add_argument('--server', default="127.0.0.1",
                    help='server name (default: 127.0.0.1)')
parser.add_argument('--database', default="")
parser.add_argument('--username', default="")
parser.add_argument('--password', default="")
parser.add_argument('--application-intent', default="readwrite")
parser.add_argument('--query', default="",
                    help='query to execute (default: "")')

args = parser.parse_args()

CONN_INFO = {
    "server": args.server,
    "database": args.database,
    "username": args.username,
    "password": args.password,
    "application_intent": args.application_intent,
}
CONN_TEMPLATE_STR = ";".join(
    [
        "DRIVER={{ODBC Driver 17 for SQL Server}}",
        "SERVER={server}",
        "DATABASE={database}",
        "UID={username}",
        "PWD={password}",
        "applicationintent={application_intent}",
    ],
)
conn_string = CONN_TEMPLATE_STR.format(**CONN_INFO)
SQL_ATTR_CONNECTION_TIMEOUT = 113
login_timeout = 2
connection_timeout = 2
cnx = pyodbc.connect(conn_string,
                     timeout=login_timeout,
                     attrs_before={
                         SQL_ATTR_CONNECTION_TIMEOUT: connection_timeout,
                     })
cursor = cnx.cursor()

if args.query.strip() == "":
    exit()

cursor.execute(args.query)
for row in cursor:
    print(row)
