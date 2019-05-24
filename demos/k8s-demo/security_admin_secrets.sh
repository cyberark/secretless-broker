# database url accessible to kubernetes cluster and local machine
# NOTE: Defined in pg.yml as nodePort
# admin-user credentials
export DB_ADMIN_USER=postgres
export DB_ADMIN_PASSWORD=admin_password

# application-user credentials
export DB_USER=quick_start
export DB_INITIAL_PASSWORD=quick_start
