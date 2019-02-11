#include "lib.h"
#include <stdio.h>
#include <iostream>
#include <string>

// TODO: look into setting rpath
// http://gridengine.eu/index.php/other-stories/232-avoiding-the-ldlibrarypath-with-shared-libs-in-go-cgo-applications-2015-12-21


// BUILD
// go build -buildmode c-shared -o ./cmd/envoy/lib.a ./cmd/envoy/main.go
// g++ -o ./cmd/envoy/main ./cmd/envoy/main.cpp ./cmd/envoy/lib.a
// cd ./cmd/envoy; ./main; cd ../..;
char* to_c_string(std::string str) {
  return &str[0u];
}

int main() {
  StoredSecret password = {
    .ID=to_c_string("db-password"),
    .Provider=to_c_string("literal"),
    .Name=to_c_string("db-password")
  };

  std::string passwordValue = GetSecret(password);
  std::cout << "Secret:" << passwordValue;
}
