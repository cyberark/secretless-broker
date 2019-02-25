# C shared library

**Status**: Pre-alpha

This is an experiment in exposing the internals of Secretless as a C shared library.
Static would be best but for now this needs dynamic linking since some Secretless components use CGO.

## BUILD

Build the shared library *by* changing directory to the Secretless project root and running `go build` with the `-buildmode` flag set to `c-shared`:

```
cd ../..; // go to secretless project root 
go build -buildmode c-shared -o ./cmd/shared/lib.a ./cmd/shared/main.go
```

## USAGE

An example C++ program's source file (`main.cpp`) is provided in the same directory as this README.

This program 
+ imports the headers from the shared library
+ uses the shared struct `StoredSecret` to specify the ID, Provider and Name of the secret to be retrieved
+ uses the shared function `GetSecret` to retrieve the `StoredSecret` using the specified provider.

```cgo
StoredSecret password = {
    .ID=to_c_string("db_password"),
    .Provider=to_c_string("env"),
    .Name=to_c_string("db-password")
  };

std::string passwordValue = GetSecret(password);
```

To build the example program using the shared library built from [BUILD](#BUILD) section above, run the following: 
```bash
g++ -o example ./cmd/shared/main.cpp ./cmd/shared/lib.a
```

Run the resulting binary making sure to specify 
1. the dynamic library load path otherwise the program might not be able to find the shared library
2. an environment variable named `db_password`, the example program uses the `env` secret provider

```
db_password=completelysecret \
  LD_LIBRARY_PATH=$PWD/cmd/shared \
  DYLD_LIBRARY_PATH=$LD_LIBRARY_PATH \
  ./example
```

We expect the following output from running the example binary.
```
2019/02/22 12:02:42 Instantiating provider 'env'
Secret:completelysecret
```
