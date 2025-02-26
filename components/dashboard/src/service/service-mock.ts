/**
 * Copyright (c) 2021 Gitpod GmbH. All rights reserved.
 * Licensed under the GNU Affero General Public License (AGPL).
 * See License-AGPL.txt in the project root for license information.
 */

import { createServiceMock, Event, Project, Team, User } from "@gitpod/gitpod-protocol";



const u1: User = {
    "id": "1234",
    "creationDate": "2018-05-01T07:00:00.000Z",
    "avatarUrl": "https://avatars.githubusercontent.com/u/10137?v=4",
    "name": "gp-test",
    "fullName": "Alex",
    "identities": [
        {
            "authProviderId": "Public-GitHub",
            "authId": "1234",
            "authName": "GitpodTester",
            "primaryEmail": "tester@gitpod.io",
        }
    ],
    rolesOrPermissions: ["teams-and-projects"],
    additionalData: {
        whatsNewSeen: {
            "April-2021": "true",
            "June-2021": "true",
        },
        emailNotificationSettings: {
            allowsChangelogMail: true,
            allowsDevXMail: true
        }
    }
}
const t1 = new Date(Date.now() - 123533).toISOString();
const team1: Team = {
    id: "team1",
    name: "ACME",
    slug: "ACME",
    creationTime: t1
}
const pr1: Project = {
    appInstallationId: "app1",
    cloneUrl: "https://github.com/AlexTugarev/txt.git",
    creationTime: t1,
    id: "pr1",
    name: "TXT",
    teamId: "team1",
}
const gitpodServiceMock = createServiceMock({
    getLoggedInUser: async () => {
        return u1;
    },
    updateLoggedInUser: async (user: User) => {
        for (const attribute in user) {
            // @ts-ignore
            u1[attribute] = user[attribute];
        }
        return u1;
    },
    getTeams: async () => {
        return [team1]
    },
    getTeamMembers: async (teamId) => {
        return [{
            memberSince: t1,
            role: "owner",
            userId: u1.id,
            avatarUrl: u1.avatarUrl,
            fullName: u1.fullName,
            primaryEmail: "alex@gitpod.io",
        }]
    },
    getGenericInvite: async () => {
        return {
            id: "000",
            creationTime: t1,
            invalidationTime: t1,
            role: "member",
            teamId: "team1",
        }
    },
    getTeamProjects: async () => {
        return [pr1]
    },
    getUserProjects: async () => {
        return [pr1]
    },
    getProjectOverview: async () => {
        return {
            branches: [{
                name: "main",
                url: "branchUrl",
                changeDate: t1,
                changeAuthor: u1.fullName!,
                changeAuthorAvatar: u1.avatarUrl,
                changeHash: "2C0FFE",
                changeTitle: "[Comp] Add new functionality for",
                isDefault: true,
                status: "available",
            }]
        }
    },
    findPrebuilds: async (p) => {
        const { projectId } = p;
        return [{
            info: {
                id: "pb1",
                branch: "main",
                buildWorkspaceId: "123",
                teamId: "t1",
                projectId,
                projectName: "pb1",
                cloneUrl: pr1.cloneUrl,
                startedAt: t1,
                startedBy: u1.id,
                startedByAvatar: u1.avatarUrl,
                changeTitle: "[Comp] Add new functionality for",
                changeDate: t1,
                changeAuthor: u1.fullName!,
                changeAuthorAvatar: u1.avatarUrl,
                changePR: "4647",
                changeUrl: "https://github.com/gitpod-io/gitpod/pull/4738",
                changeHash: "2C0FFE"
            }, status: "available"
        }, {
            info: {
                id: "pb1",
                branch: "foo/bar",
                buildWorkspaceId: "1234",
                teamId: "t1",
                projectId,
                projectName: "pb1",
                cloneUrl: pr1.cloneUrl,
                startedAt: t1,
                startedBy: u1.id,
                startedByAvatar: u1.avatarUrl,
                changeTitle: "Fix Bug Nr 1",
                changeDate: t1,
                changeAuthor: u1.fullName!,
                changeAuthorAvatar: u1.avatarUrl,
                changePR: "4245",
                changeUrl: "https://github.com/gitpod-io/gitpod/pull/4738",
                changeHash: "1C0FFE"
            }, status: "available"
        }
        ]
    },
    getProviderRepositoriesForUser: async () => {
        return []
    },
    getWorkspaces: async () => {
        return []
    },
    getFeaturedRepositories: async () => {
        return []
    },
    getAuthProviders: async () => {
        return [{
            "authProviderId": "Public-GitHub",
            "authProviderType": "GitHub",
            "verified": true,
            "host": "github.com",
            "icon": "",
            "description": "",
            "isReadonly": false
        },
        {
            "authProviderId": "Public-GitLab",
            "authProviderType": "GitLab",
            "verified": true,
            "host": "gitlab.com",
            "icon": "",
            "description": "",
            "isReadonly": false
        }]
    },
    getOwnAuthProviders: async () => {
        return [{
            "id": "foobar123",
            "ownerId": "1234",
            "status": "verified",
            "host": "testing.doptig.com/gitlab",
            "type": "GitLab",
            "oauth": {
                "authorizationUrl": "https://testing.doptig.com/gitlab/oauth/authorize",
                "tokenUrl": "https://testing.doptig.com/gitlab/oauth/token",
                "settingsUrl": "https://testing.doptig.com/gitlab/-/profile/applications",
                "callBackUrl": "https://gitpod-staging.com/auth/testing.doptig.com/gitlab/callback",
                "clientId": "clientid-123",
                "clientSecret": "redacted"
            },
            "deleted": false
        }]
    },
    onDidOpenConnection: Event.None,
    onDidCloseConnection: Event.None,

});

export { gitpodServiceMock };

