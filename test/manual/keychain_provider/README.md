# keychain_provider tests

Note: This directory does not contain `start`, so it will not be run by the wrapper script `./bin/test`.

The reason is that (keychain_provider_test.go)[keychain_provider_test.go] will open a confirmation dialog, which does not play well with automation.

The OS X Keychain Provider is meant for interactive use by a user; this dialog is a security feature of the OS.

To run the test locally if you are working on a Mac, just run
```
./start
./test
```
from this directory. The `start` script prepares your local
environment by adding a secret to your OSX Keyring. Note that you may be
prompted to grant access to your OSX Keyring in order for this test to run
properly. To clean up after running the test, you can run `./stop`.
