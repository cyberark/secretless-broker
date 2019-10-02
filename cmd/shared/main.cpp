#include "lib.h"
#include <stdio.h>
#include <iostream>
#include <string>

char* to_c_string(std::string str) {
  return &str[0u];
}

int main() {
  CredentialSpec passwordSpec = {
    .Name=to_c_string("db-password")
    .Get=to_c_string("db_password"),
    .From=to_c_string("env"),
  };

  std::string passwordValue = GetCredential(passwordSpec);
  std::cout << "Credential:" << passwordValue;
}
