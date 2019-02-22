#include "lib.h"
#include <stdio.h>
#include <iostream>
#include <string>

char* to_c_string(std::string str) {
  return &str[0u];
}

int main() {
  StoredSecret password = {
    .ID=to_c_string("db_password"),
    .Provider=to_c_string("env"),
    .Name=to_c_string("db-password")
  };

  std::string passwordValue = GetSecret(password);
  std::cout << "Secret:" << passwordValue;
}
