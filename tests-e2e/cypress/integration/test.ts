// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
/// <reference path="../support/index.d.ts" />

import '@testing-library/cypress/add-commands';

describe('Testing playground', () => {
    beforeEach(() => {
        // # Login as non-admin user
        cy.apiLogin('user-1');
        cy.visit('/');
    });

    it('who tests the testers?', () => {
        // WORKING
        //
        // cy.visitNewPublicChannel();
        // cy.verifyExportSystemMessage();

        // STILL TO BE TESTED
        //
        // apiExportChannel();
        // archiveChannel();
        // unarchiveChannel();
        // leaveChannel();
        // inviteUser();
        // kickUser();
        // postMessages();
        // verifyChannelDoesNotExist();
        // verifyExportBotMessage();
        // verifyExportCommandIsAvailable();
        // verifyFileCanBeDownloaded();
        // verifyFileExtension();
        // verifyFileName();
        // verifyNoExport();
        // verifyNoPosts();
        // verifySuccessfulExport();
        // visitNewDirectMessage();
        // visitNewGroupMessage();
        // visitNewPrivateChannel();
        // visitSelfDM();
    });
});
