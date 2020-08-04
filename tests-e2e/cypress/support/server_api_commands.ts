// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import {Team} from 'mattermost-redux/types/teams';
import {Channel, ChannelType} from 'mattermost-redux/types/channels';
import {UserProfile} from 'mattermost-redux/types/users';
import {Post} from 'mattermost-redux/types/posts';

import users from '../fixtures/users';
import {httpStatusOk, httpStatusCreated} from '../support/constants';

function apiLogin(username = 'user-1', password : string | null = null) : Cypress.Chainable<Cypress.Response> {
    return cy.request({
        headers: {'X-Requested-With': 'XMLHttpRequest'},
        url: '/api/v4/users/login',
        method: 'POST',
        body: {
            login_id: username,
            password: password || users[username].password,
        },
    }).then((response: Cypress.Response) => {
        expect(response.status).to.equal(httpStatusOk);
        return cy.wrap(response);
    });
}
Cypress.Commands.add('apiLogin', apiLogin);

// Return a function to create either a public or private channel depending on
// the channelType parameter.
function apiCreateChannel(channelType: ChannelType) : ((teamId: string, name: string, displayName: string) => Cypress.Chainable<Channel>) {
    return (teamId: string, name: string, displayName: string): Cypress.Chainable<Channel> => {
        return cy.request({
            headers: {'X-Requested-With': 'XMLHttpRequest'},
            url: '/api/v4/channels',
            method: 'POST',
            body: {
                team_id: teamId,
                name,
                display_name: displayName,
                type: channelType,
            },
        }).then((response: Cypress.Response) => {
            expect(response.status).to.equal(httpStatusCreated);

            const channel = response.body as Channel;
            return cy.wrap(channel);
        });
    };
}
Cypress.Commands.add('apiCreatePublicChannel', apiCreateChannel('O'));
Cypress.Commands.add('apiCreatePrivateChannel', apiCreateChannel('P'));

function apiCreateGroupMessage(userIds : string[]) : Cypress.Chainable<Channel> {
    return cy.request({
        headers: {'X-Requested-With': 'XMLHttpRequest'},
        url: '/api/v4/channels/group',
        method: 'POST',
        body: userIds,
    }).then((response: Cypress.Response) => {
        expect(response.status).to.equal(httpStatusCreated);

        const channel = response.body as Channel;
        return cy.wrap(channel);
    });
}
Cypress.Commands.add('apiCreateGroupMessage', apiCreateGroupMessage);

function apiCreateDirectMessage(selfId : string, otherId : string) : Cypress.Chainable<Channel> {
    return cy.request({
        headers: {'X-Requested-With': 'XMLHttpRequest'},
        url: '/api/v4/channels/direct',
        method: 'POST',
        body: [selfId, otherId],
    }).then((response: Cypress.Response) => {
        expect(response.status).to.equal(httpStatusCreated);

        const channel = response.body as Channel;
        return cy.wrap(channel);
    });
}
Cypress.Commands.add('apiCreateDirectMessage', apiCreateDirectMessage);

function apiGetTeamByName(name: string) : Cypress.Chainable<Team> {
    return cy.request({
        headers: {'X-Requested-With': 'XMLHttpRequest'},
        url: `/api/v4/teams/name/${name}`,
        method: 'GET',
        body: {},
    }).then((response: Cypress.Response) => {
        expect(response.status).to.equal(httpStatusOk);

        const team = response.body as Team;
        return cy.wrap(team);
    });
}
Cypress.Commands.add('apiGetTeamByName', apiGetTeamByName);

function apiGetUserByUsername(name: string) : Cypress.Chainable<UserProfile> {
    return cy.request({
        headers: {'X-Requested-With': 'XMLHttpRequest'},
        url: '/api/v4/users/username/' + name,
        method: 'GET',
    }).then((response: Cypress.Response) => {
        expect(response.status).to.equal(httpStatusOk);

        const user = response.body as UserProfile;
        return cy.wrap(user);
    });
}
Cypress.Commands.add('apiGetUserByUsername', apiGetUserByUsername);

function apiGetUsers(usernames : string[] = []) : Cypress.Chainable<UserProfile[]> {
    return cy.request({
        headers: {'X-Requested-With': 'XMLHttpRequest'},
        url: '/api/v4/users/usernames',
        method: 'POST',
        body: usernames,
    }).then((response: Cypress.Response) => {
        expect(response.status).to.equal(httpStatusOk);

        const userList = response.body as UserProfile[];
        return cy.wrap(userList);
    });
}
Cypress.Commands.add('apiGetUsers', apiGetUsers);

function apiCreatePost(channelId: string, message: string, fileIds: string[] = []) : Cypress.Chainable<Post> {
    return cy.request({
        headers: {'X-Requested-With': 'XMLHttpRequest'},
        url: '/api/v4/posts',
        method: 'POST',
        body: {
            channel_id: channelId,
            message,
            file_ids: fileIds,
        },
    }).then((response: Cypress.Response) => {
        expect(response.status).to.equal(httpStatusCreated);

        const post = response.body as Post;
        return cy.wrap(post);
    });
}
Cypress.Commands.add('apiCreatePost', apiCreatePost);

function apiGetChannelByName(teamName: string, channelName: string) : Cypress.Chainable<Channel> {
    return cy.request({
        headers: {'X-Requested-With': 'XMLHttpRequest'},
        url: `/api/v4/teams/name/ad-1/channels/name/${channelName}`,
        method: 'GET',
        body: {},
    }).then((response: Cypress.Response) => {
        expect(response.status).to.equal(httpStatusOk);

        const channel = response.body as Channel;
        return cy.wrap(channel);
    });
}
Cypress.Commands.add('apiGetChannelByName', apiGetChannelByName);

function apiMakeChannelReadOnly(channelId: string) : Cypress.Chainable<Cypress.Response> {
    return cy.request({
        headers: {'X-Requested-With': 'XMLHttpRequest'},
        url: `/api/v4/channels/${channelId}/moderations/patch`,
        method: 'PUT',
        body: [{name: 'create_post', roles: {members: false, guests: false}}],
    }).then((response: Cypress.Response) => {
        expect(response.status).to.equal(httpStatusOk);
        return cy.wrap(response);
    });
}
Cypress.Commands.add('apiMakeChannelReadOnly', apiMakeChannelReadOnly);

function apiExportChannel(channelId: string, expectedStatus: number = httpStatusOk) : Cypress.Chainable<string> {
    const endpoint = '/plugins/com.mattermost.plugin-channel-export/api/v1/export';
    const queryString = `?format=csv&channel_id=${channelId}`;

    return cy.request({
        headers: {'X-Requested-With': 'XMLHttpRequest'},
        url: endpoint + queryString,
        method: 'GET',
        failOnStatusCode: false,
    }).then((response: Cypress.Response) => {
        expect(response.status).to.equal(expectedStatus);

        const file = response.body as string;
        return cy.wrap(file);
    });
}
Cypress.Commands.add('apiExportChannel', apiExportChannel);
