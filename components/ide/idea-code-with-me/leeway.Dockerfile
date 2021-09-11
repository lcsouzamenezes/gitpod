# Copyright (c) 2021 Gitpod GmbH. All rights reserved.
# Licensed under the GNU Affero General Public License (AGPL).
# See License-AGPL.txt in the project root for license information.

FROM scratch

# copy static web resources in first layer to serve from blobserve
COPY --chown=33333:33333 index.html startup.sh supervisor-ide-config.json /ide/
