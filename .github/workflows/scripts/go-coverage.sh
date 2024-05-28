#!/bin/sh
# Copyright (C) 2024 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

cat > $cov-ignore << EndOfMessage
zz_generated.deepcopy.go
openapi_generated.go
testing
tests
test
EndOfMessage

PASSPERCENT=10

while read p || [ -n "$p" ]
do
    sed -i "/${p}/d" ./coverage.out
done < $cov-ignore

# get the total coverage percentage number
COVPERCENT=$(go tool cover -func ./coverage.out | grep total | awk '{print $3}')
# remove the % sign
COVPERCENT=${COVPERCENT%\%}
echo "Coverage: $COVPERCENT%"

# if coverage is less than $PASSPERCENT then exit with error
if [ $(echo "$COVPERCENT < $PASSPERCENT" | bc) -eq 1 ]; then
    echo "Coverage is less than $PASSPERCENT%. Failed!"
    exit 1
fi

echo "Coverage is bigger than $PASSPERCENT%. Pass!"
