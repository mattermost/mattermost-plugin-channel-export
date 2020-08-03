// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import {Team} from 'mattermost-redux/types/teams';
import {Channel} from 'mattermost-redux/types/channels';
import {UserProfile} from 'mattermost-redux/types/users';

function waitUntilPermanentPost() {
    cy.get('#postListContent').should('be.visible');
    cy.waitUntil(() =>
        cy.findAllByTestId('postView').
            last().
            then((el) => !(el[0].id.includes(':')))
    );
}

function postMessage(message: string): void {
    cy.findByTestId('post_textbox').clear().type(message).type('{enter}');
    cy.findByTestId('post_textbox').should('have.text', '');
}
Cypress.Commands.add('postMessage', postMessage);

function exportSlashCommand() : void {
    cy.postMessage('/export');
}
Cypress.Commands.add('exportSlashCommand', exportSlashCommand);

function visitNewPublicChannel() : void {
    const id = Date.now().toString();
    const name = `channelexport_${id}`;
    const displayName = `Channel Export - ${id}`;

    cy.apiGetTeamByName('ad-1').then((team: Team) => {
        return cy.apiCreatePublicChannel(team.id, name, displayName);
    }).then((response: Channel) => {
        cy.visit(`/ad-1/channels/${name}`);
        return cy.wrap(response);
    });
}
Cypress.Commands.add('visitNewPublicChannel', visitNewPublicChannel);

function getLastPostId() : Cypress.Chainable<string> {
    waitUntilPermanentPost();

    return cy.findAllByTestId('postView').last().should('have.attr', 'id').and('not.include', ':').
        invoke('replace', 'post_', '');
}
Cypress.Commands.add('getLastPostId', getLastPostId);

function verifyExportSystemMessage(channelName : string) : void {
    cy.getLastPostId().then((lastPostId: string) => {
        cy.get(`#post_${lastPostId}`).should('contain.text', `Exporting ~${channelName}. @channelexport will send you a direct message when the export is ready.`);
    });
}
Cypress.Commands.add('verifyExportSystemMessage', verifyExportSystemMessage);

function verifyExportBotMessage(channelName : string, userName = 'user-1', botName = 'channelexport') : void {
    interface DM {
        user: UserProfile;
        bot: UserProfile;
    }

    cy.apiGetUserByUsername(userName).then((user: UserProfile) => {
        return cy.apiGetUserByUsername(botName).then((bot: UserProfile) => {
            return cy.wrap({user, bot});
        });
    }).then((dm: DM) => {
        cy.get(`#sidebarItem_${dm.user.id}__${dm.bot.id}`).click();
        cy.getLastPostId().then((lastPostId: string) => {
            cy.get(`#post_${lastPostId}`).
                should('contain.text', `Channel ~${channelName} exported:`);
        });
    });
}
Cypress.Commands.add('verifyExportBotMessage', verifyExportBotMessage);
