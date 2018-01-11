# keychain_provider tests

Note: This directory does not contain `start.sh`, so it will not be run by the wrapper script `./build/test.sh`. 

The reason is that (keychain_provider_test.go)[keychain_provider_test.go] will open a confirmation dialog, which does not play well with automation. 

The OS X Keychain Provider is meant for interactive use by a user; this dialog is a security feature of the OS.
