#!/bin/bash
# Copyright (c) 2021 Gitpod GmbH. All rights reserved.
# Licensed under the GNU Affero General Public License (AGPL).
# See License-AGPL.txt in the project root for license information.

export CGO_ENABLED=0

mkdir -p bin

# shellcheck disable=SC2044
for i in pkg/agent/*; do
    echo building agent "$i"
    base=$(basename "$i")
    go build -trimpath -ldflags="-buildid= -w -s" -o bin/gitpod-integration-test-"${base%_agent}"-agent ./pkg/agent/"$i"
done

go test -trimpath -ldflags="-buildid= -w -s" -c -o bin/integration.test .
