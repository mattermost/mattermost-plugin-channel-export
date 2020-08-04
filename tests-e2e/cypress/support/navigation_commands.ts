// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import {Team} from 'mattermost-redux/types/teams';
import {Channel, ChannelType} from 'mattermost-redux/types/channels';
import {UserProfile} from 'mattermost-redux/types/users';

// Create a public or private channel depending on the channelType parameter
// and visit it.
function visitNewChannel(channelType: ChannelType) : (() => Cypress.Chainable<Channel>) {
    // Select the correct function to create a channel (private or public).
    let apiCreateChannel = cy.apiCreatePrivateChannel;
    if (channelType === 'O') {
        apiCreateChannel = cy.apiCreatePublicChannel;
    }

    return () => {
        // Generate a unique name for the channel
        const id = Date.now().toString();
        const name = `channelexport_${id}`;
        const displayName = `Channel Export - ${id}`;

        // # Get the team's information to retrieve its ID.
        return cy.apiGetTeamByName('ad-1').then((team: Team) => {
            // # Create the channel.
            return apiCreateChannel(team.id, name, displayName);
        }).then((response: Channel) => {
            // # Visit the new channel.
            cy.visit(`/ad-1/channels/${name}`);

            return cy.wrap(response);
        });
    };
}
Cypress.Commands.add('visitNewPublicChannel', visitNewChannel('O'));
Cypress.Commands.add('visitNewPrivateChannel', visitNewChannel('P'));

function visitNewGroupMessage(userNames: string[]) : Cypress.Chainable<Channel> {
    // # Get the users in the group to retrieve their IDs.
    return cy.apiGetUsers(userNames).then((users : UserProfile[]) => {
        const userIds = users.map((u) => u.id);

        // # Cerate a group message via the API.
        return cy.apiCreateGroupMessage(userIds).then((channel: Channel) => {
            // # Visit the new channel.
            cy.visit(`/ad-1/messages/${channel.name}`);

            return cy.wrap(channel);
        });
    });
}
Cypress.Commands.add('visitNewGroupMessage', visitNewGroupMessage);

function visitNewDirectMessage(creatorName: string, otherName: string) : Cypress.Chainable<Channel> {
    // # Get the users in the DM to retrieve their IDs.
    return cy.apiGetUsers([creatorName, otherName]).then((users : UserProfile[]) => {
        const [selfId, otherId] = users.map((u) => u.id);

        // # Cerate a direct message via the API.
        return cy.apiCreateDirectMessage(selfId, otherId).then((channel: Channel) => {
            // # Visit the new channel.
            cy.visit(`/ad-1/messages/@${otherName}`);
            return cy.wrap(channel);
        });
    });
}
Cypress.Commands.add('visitNewDirectMessage', visitNewDirectMessage);

function visitDMWithBot(userName: string, botName = 'channelexport') : void {
    interface DM {
        user: UserProfile;
        bot: UserProfile;
    }

    // # Get the user and bot to retrieve their IDs.
    cy.apiGetUserByUsername(userName).then((user: UserProfile) => {
        return cy.apiGetUserByUsername(botName).then((bot: UserProfile) => {
            return cy.wrap({user, bot});
        });
    }).then((dm: DM) => {
        // # Click on the sidebar item corresponding to the DM with the bot.
        cy.get(`#sidebarItem_${dm.user.id}__${dm.bot.id}`).click();
    });
}
Cypress.Commands.add('visitDMWithBot', visitDMWithBot);

