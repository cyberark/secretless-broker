# keychain_provider tests

The OS X Keychain Provider is meant for interactive use by a user. In production,
it will open a confirmation dialog at least once to request access to the
keychain for Secretless.

The tests are captured in [keychain_provider_test.go](keychain_provider_test.go).

To run the test locally if you are working on a Mac, just run

```
./test
```

from this directory. The `test` script both prepares your local environment by
adding secrets to your OSX Keyring and runs the tests against the Keyring. It's
necessary for the Secrets to be added in the same process where the tests are
run because the keychain automatically trusts (to read a secret) the process
that writes a secret. Without this a user would need to confirm a keychain
prompt at least once before a secret read is permitted.

To independently clean up any test fixtures you can run `./stop`.
