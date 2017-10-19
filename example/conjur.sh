#!/bin/bash -ex

conjur policy load root example/conjur.yml
conjur variable values add pg/username conjur
conjur variable values add pg/password conjur
conjur variable values add pg/url pg:5432
