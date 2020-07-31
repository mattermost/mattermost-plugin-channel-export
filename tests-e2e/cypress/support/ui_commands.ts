// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.


import {Team} from 'mattermost-redux/types/teams';
import {Channel} from 'mattermost-redux/types/channels';

Cypress.Commands.add('postMessage', (message: string) => {
    cy.findByTestId('post_textbox').clear().type(message).type('{enter}');
    cy.findByTestId('post_textbox').should('have.text', '');
});

Cypress.Commands.add('exportSlashCommand', () => {
    cy.postMessage('/export');
});

Cypress.Commands.add('visitNewPublicChannel', () => {
    const id = Date.now().toString();
    const name = `channelexport_${id}`;
    const displayName = `Channel Export - ${id}`;

    cy.apiGetTeamByName('ad-1').then((team: Team) => {
        return cy.apiCreatePublicChannel(team.id, name, displayName);
    }).then((response: Channel) => {
        cy.visit(`/ad-1/channels/${name}`);
        return cy.wrap(response);
    });
});

function waitUntilPermanentPost() {
    cy.get('#postListContent').should('be.visible');
    cy.waitUntil(() => cy.findAllByTestId('postView').last().then((el) => !(el[0].id.includes(':'))));
}

Cypress.Commands.add('getLastPostId', () => {
    waitUntilPermanentPost();

    cy.findAllByTestId('postView').last().should('have.attr', 'id').and('not.include', ':').
        invoke('replace', 'post_', '');
});

Cypress.Commands.add('verifyExportSystemMessage', (channelName : string) => {
    cy.getLastPostId().then((lastPostId: string) => {
        cy.get(`#post_${lastPostId}`).should('contain.text', `Exporting ~${channelName}. @channelexport will send you a direct message when the export is ready.`);
    });
});
