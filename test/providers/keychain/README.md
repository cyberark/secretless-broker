# keychain_provider tests

The OS X Keychain Provider is meant for interactive use by a user, which
is why this test is "manual".

The (keychain_provider_test.go)[keychain_provider_test.go] will open a
confirmation dialog during this test, and you'll need to click "Allow"

To run the test locally if you are working on a Mac, just run

```
./start
./test
```

from this directory. The `start` script prepares your local
environment by adding a secret to your OSX Keyring.  To clean up after running
the test, you can run `./stop`.
