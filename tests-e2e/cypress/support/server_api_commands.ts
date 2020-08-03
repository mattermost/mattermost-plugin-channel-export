// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import {Team} from 'mattermost-redux/types/teams';
import {Channel, ChannelType} from 'mattermost-redux/types/channels';
import {UserProfile} from 'mattermost-redux/types/users';
import {Post} from 'mattermost-redux/types/posts';

import users from '../fixtures/users';

// *****************************************************************************
// Authentication
// https://api.mattermost.com/#tag/authentication
// *****************************************************************************

const httpStatusOk = 200;
const httpStatusCreated = 201;

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

function apiCreateMultiplePosts(channelId: string, numMessages: number) : void {
    for (let i = 0; i < numMessages; i++) {
        cy.apiCreatePost(channelId, 'lorem ipsum ' + i);
    }
}
Cypress.Commands.add('apiCreateMultiplePosts', apiCreateMultiplePosts);

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

function apiExportChannel(channelId: string) : Cypress.Chainable<string> {
    const endpoint = '/plugins/com.mattermost.plugin-channel-export/api/v1/export';
    const queryString = `?format=csv&channel_id=${channelId}`;

    return cy.request({
        headers: {'X-Requested-With': 'XMLHttpRequest'},
        url: endpoint + queryString,
        method: 'GET',
    }).then((response: Cypress.Response) => {
        expect(response.status).to.equal(httpStatusOk);

        const file = response.body as string;
        return cy.wrap(file);
    });
}
Cypress.Commands.add('apiExportChannel', apiExportChannel);

// // Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// // See LICENSE.txt for license information.

// import {getRandomInt} from '../utils';
// import users from '../fixtures/users.json';
// import timeouts from '../fixtures/timeouts';

// // *****************************************************************************
// // Authentication
// // https://api.mattermost.com/#tag/authentication
// // *****************************************************************************

// /**
//  * User login directly via API
//  * @param {String} username - username
//  * @param {String} password - password
//  */
// Cypress.Commands.add('apiLogin', (username = 'user-1', password = null) => {
//     cy.request({
//         headers: {'X-Requested-With': 'XMLHttpRequest'},
//         url: '/api/v4/users/login',
//         method: 'POST',
//         body: {login_id: username, password: password || users[username].password},
//     }).then((response) => {
//         expect(response.status).to.equal(200);
//         return cy.wrap(response);
//     });
// });

// /**
//  * Logout a user directly via API
//  */
// Cypress.Commands.add('apiLogout', () => {
//     cy.request({
//         headers: {'X-Requested-With': 'XMLHttpRequest'},
//         url: '/api/v4/users/logout',
//         method: 'POST',
//         log: false,
//         timeout: timeouts.HUGE,
//     });
// });

// // *****************************************************************************
// // Teams
// // https://api.mattermost.com/#tag/teams
// // *****************************************************************************

// /**
//  * Creates a team directly via API
//  * This API assume that the user is logged in and has permission to access
//  * @param {String} name - Unique handler for a team, will be present in the team URL
//  * @param {String} displayName - Non-unique UI name for the team
//  * @param {String} type - 'O' for open (default), 'I' for invite only
//  * All parameters required
//  */
// Cypress.Commands.add('apiCreateTeam', (name, displayName, type = 'O') => {
//     const uniqueName = `${name}-${getRandomInt(9999).toString()}`;

//     return cy.request({
//         headers: {'X-Requested-With': 'XMLHttpRequest'},
//         url: '/api/v4/teams',
//         method: 'POST',
//         body: {
//             name: uniqueName,
//             display_name: displayName,
//             type,
//         },
//     }).then((response) => {
//         expect(response.status).to.equal(201);
//         cy.wrap(response);
//     });
// });

// /**
//  * Add user into a team directly via API
//  * This API assume that the user is logged in and has permission to access
//  * @param {String} teamId - The team ID
//  * @param {String} userId - ID of user to be added into a team
//  * All parameter required
//  */
// Cypress.Commands.add('apiAddUserToTeam', (teamId, userId) => {
//     cy.request({
//         method: 'POST',
//         url: `/api/v4/teams/${teamId}/members`,
//         headers: {'X-Requested-With': 'XMLHttpRequest'},
//         body: {team_id: teamId, user_id: userId},
//         qs: {team_id: teamId},
//     }).then((response) => {
//         expect(response.status).to.equal(201);
//         return cy.wrap(response);
//     });
// });

// // *****************************************************************************
// // Users
// // https://api.mattermost.com/#tag/users
// // *****************************************************************************

// /**
//  * Get user by email directly via API
//  * This API assume that the user is logged in and has permission to access
//  * @param {String} email
//  * All parameter required
//  */
// Cypress.Commands.add('apiGetUserByEmail', (email) => {
//     return cy.request({
//         headers: {'X-Requested-With': 'XMLHttpRequest'},
//         url: '/api/v4/users/email/' + email,
//     }).then((response) => {
//         expect(response.status).to.equal(200);
//         cy.wrap(response);
//     });
// });

// // *****************************************************************************
// // Plugins
// // https://api.mattermost.com/#tag/plugins
// // *****************************************************************************

// /**
//  * Get webapp plugins directly via API
//  * This API assume that the user is logged in and has permission to access
//  */
// Cypress.Commands.add('apiGetWebappPlugins', () => {
//     return cy.request({
//         headers: {'X-Requested-With': 'XMLHttpRequest'},
//         url: '/api/v4/plugins/webapp',
//         method: 'GET',
//     }).then((response) => {
//         expect(response.status).to.equal(200);
//         cy.wrap(response);
//     });
// });

/**
 * Creates a group channel directly via API
 * This API assume that the user is logged in and has cookie to access
 * @param {String} userIds - IDs of users as member of the group
 * All parameters required except purpose and header
 */
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

function apiCreateDirectMessage(userIds : string[]) : Cypress.Chainable<Channel> {
    return cy.request({
        headers: {'X-Requested-With': 'XMLHttpRequest'},
        url: '/api/v4/channels/direct',
        method: 'POST',
        body: userIds,
    }).then((response: Cypress.Response) => {
        expect(response.status).to.equal(httpStatusCreated);

        const channel = response.body as Channel;
        return cy.wrap(channel);
    });
}
Cypress.Commands.add('apiCreateDirectMessage', apiCreateDirectMessage);

// /**
//  * Gets current user's teams
//  */

// Cypress.Commands.add('apiGetTeams', () => {
//     return cy.request({
//         headers: {'X-Requested-With': 'XMLHttpRequest'},
//         url: 'api/v4/users/me/teams',
//         method: 'GET',
//     }).then((response) => {
//         expect(response.status).to.equal(200);
//         return cy.wrap(response);
//     });
// });

/**
* Gets users
*/
Cypress.Commands.add('apiGetUsers', (usernames : string[] = []) => {
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
});

// Cypress.Commands.add('apiCreateGroupChannel', (userList = [], teamName) => {
//     cy.apiGetUsers(userList).then((res) => {
//         const userIds = res.body.map((user) => user.id);
//         cy.apiCreateGroup(userIds).then((resp) => {
//             cy.apiGetTeams().then((response) => {
//                 const teamNameUrl = teamName ? teamName : response.body[0].name;
//                 cy.visit(`/${teamNameUrl}/messages/${resp.body.name}`);
//             });
//         });
//     });
// });
