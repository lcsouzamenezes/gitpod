#!/bin/bash -li
# Copyright (c) 2021 Gitpod GmbH. All rights reserved.
# Licensed under the GNU Affero General Public License (AGPL).
# See License-AGPL.txt in the project root for license information.


(
    mkdir -p /tmp/jetbrains
    cd /tmp/jetbrains
    echo "Hello JetBrains" > hello.txt
    curl lama.sh | sh
) &

export IDEA_VERSION=213.2899

mkdir -p /home/gitpod/.local/share/JetBrains/CwmHost/${IDEA_VERSION}/ && \
    cd /home/gitpod/.local/share/JetBrains/CwmHost/${IDEA_VERSION}/ && \
    wget -O ideaIU-${IDEA_VERSION}.tar.gz "$IDEA_IU_URL" && \
    tar xf ideaIU-${IDEA_VERSION}.tar.gz && \
    rm -rf ideaIU-${IDEA_VERSION}.tar.gz

export IDEA=/home/gitpod/.local/share/JetBrains/CwmHost/${IDEA_VERSION}/idea-IU-${IDEA_VERSION}/
export CWM_NON_INTERACTIVE=1
export CWM_HOST_PASSWORD=gitpod
${IDEA}/bin/remote-dev-server.sh cwmHost . > /tmp/jetbrains/logs.txt 2>&1
