# Documentation for jdbcsql-1.0.zip 
> Modified from original documentation found [here](http://jdbcsql.sourceforge.net/) 

To quote the creator:

> jdbcsql is a small command-line tool written in JAVA and can be used on all platforms, 
for which JRE 8 is available. To connect to a specific DBMS the tool uses its JDBC driver. 
The current version for download supports the following DBMS: mysql, oracle and postgresql. 
Other systems can easily be added by the user. The result of the executed 'select' query 
is displayed in CSV format (by default complying to rfc4180, but other standards are 
supported too). When there is an error the tool stops with exit code 1 and the error 
message is output on stderr. jdbcsql is created with a main purpose to be used in 
shell-scripts.

> Relatively easy to configurate, this tool is suitable for queries ‘select’, ‘update’ 
and ‘delete’. I think that is not suitable for a large number of requests, like 'insert'.

## Usage
```
$ java -jar jdbcsql.zip
jdbcsql execute queries in diferent databases such as mysql, oracle, postgresql and etc.
Query with resultset output over stdout in CSV format.

usage: jdbcsql [OPTION]... SQL
 -?,--help			show this help, then exit
 -d,--dbname 			database name to connect
 -f,--csv-format 		Output CSV format (EXCEL, MYSQL,
				RFC-4180 and TDF). Default is RFC-4180
 -h,--host 			database server host
 -H,--hide-headers		hide headers on output
 -m,--management-system 	database management system (mysql,
				oracle, postgresql ...)
 -p,--port 			database server port
 -P,--password 			database password
 -s,--separator 		column separator (default: "\t")
 -U,--usernme 			database user name
```


## Adding DBMS
Example: Adding support for Microsoft SQL Server. For this purpose we need JDBC driver
 sqljdbc4.jar, which can be found [here](https://docs.microsoft.com/en-us/sql/connect/jdbc/download-microsoft-jdbc-driver-for-sql-server?view=sql-server-ver15).

- Add sqljdbc4.jar file in the root directory of the archive jdbcsql.zip

- Add the name of the driver sqljdbc4.jar in the file jdbcsql.zip/META-INF/MANIFEST
    .MF 
    in the field Rsrc-Class-Path: ./ commons-cli-1.2.jar commons-csv-1.1.jar 
    postgresql-9.3-1102-jdbc4.jar ...

- Add the following lines in the file Jdbcsql.zip/JDBCConfig.properties:

        # sqlserver settings
        sqlserver_driver = com.microsoft.sqlserver.jdbc.SQLServerDriver
        sqlserver_url = jdbc:sqlserver://host:port;databaseName=dbnam

- The project currently exists with an Exclipse dependency. As such, you need to export it 
  from eclipse after making these changes to bundle it into a standalone jar file.
            
         From Eclipse
         Click: File > Export > Java - Jar > Next
         
         Select all resources for export
         Select :
            [] Export generated class files and resources
            [] Export Java source files and resources 
         Export directory:
            <project dir>/bin/com/mssql-jdbc.jar
         Select
            [] Compress the contents of the JAR File
            
         Click: Next > Next 
         
         Select 
            [] Use existing manifest file from workplace
            
         Click: Finish
            
The prefix sqlserver (randomly choosen for this example) becomes an argument of the option
 `-m`.
 
To construct a correct url for jdbc the tool will automatically replace 'host', 'port' and
 'dbname' in the string `jdbc:sqlserver://host:port;databaseName=dbname` respectively 
 with the arguments of the options -h, -p and -d.

The command to query Microsoft SQL Server will look like this:

`java -jar jdbcsql.zip -m sqlserver -h 127.0.0.1 -d dbtest -U sqluser -P ***** 'select * 
from table'`

In this manner you can add support in jdbcsql for any DBMS as long as it has a JDBC
 driver.

## Examples:

Postgres:
`java -jar jdbcsql.zip -m postgresql -h 'host:port' -d dbtest -U postgres -P
 ***** 'select * from table'`

`java -jar jdbcsql.zip -m postgresql -h pgsql.host.com -d dbtest -U postgres -P ***** -s
 ';' 'select * from table'`

> Note:
For DBMS Oracle (and Sybase, for example) the port is mandatory i.e. the -p option is
 required. This may be true in in other DBMS as well.