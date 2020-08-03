// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
/// <reference path="../support/index.d.ts" />

import '@testing-library/cypress/add-commands';
import {Channel} from 'mattermost-redux/types/channels';

describe('Test Area - Permissions', () => {
    beforeEach(() => {
        // # Login as non-admin user
        cy.apiLogin('user-1');

        // # Visit the default channel
        cy.visit('/');
    });

    it('ID 9 - User can export a public channel', () => {
        cy.visitNewPublicChannel().then((channel: Channel) => {
            cy.exportSlashCommand();
            cy.visitDMWithBot('user-1');
            cy.verifyFileCanBeDownloaded(channel);
        });
    });

    it('ID 10 - User can export a private channel', () => {
        cy.visitNewPrivateChannel().then((channel: Channel) => {
            cy.exportSlashCommand();
            cy.visitDMWithBot('user-1');
            cy.verifyFileCanBeDownloaded(channel);
        });
    });

    // it('ID 11 - User can export a group message channel', () => {
    //     cy.visitNewGroupMessage();
    //     cy.exportSlashCommand();
    //     cy.verifySuccessfulExport();
    // });

    // it('ID 12 - User can export a direct message channel', () => {
    //     cy.visitNewDirectMessage();
    //     cy.exportSlashCommand();
    //     cy.verifySuccessfulExport();
    // });

    // it('ID 13 - User can export a direct message channel with sefl', () => {
    //     cy.visitSelfDM();
    //     cy.exportSlashCommand();
    //     cy.verifySuccessfulExport();
    // });

    // it('ID 14 - User can export a bot message channel', () => {
    //     cy.visitNewGroupMessage();
    //     cy.exportSlashCommand();
    //     cy.verifySuccessfulExport();
    // });

    // it('ID 15 - User can export archived channel', () => {
    //     cy.visitNewPublicChannel().then((channel) => {
    //         cy.archiveChannel(channel);
    //         cy.apiExportChannel(channel);
    //         cy.verifySuccessfulExport();
    //     });
    // });

    // it('ID 16 - User can export an unarchived channel', () => {
    //     cy.visitNewPublicChannel().then((channel) => {
    //         cy.archiveChannel(channel);
    //         cy.unarchiveChannel(channel);
    //     });
    //     cy.exportSlashCommand();
    //     cy.verifySuccessfulExport();
    // });

    // it('ID 15 (2nd one) - User cannot export a channel they are not added to', () => {
    //     cy.visitNewPublicChannel().then((channel) => {
    //         cy.leaveChannel();
    //         cy.apiExportChannel(channel);
    //         cy.verifyNoExport();
    //     });
    // });

    // it('ID 17 - User cannot export a channel once they are ‘kicked’ from the channel', () => {
    //     cy.apiLogin('user-2');
    //     cy.visitNewPublicChannel().then((channel) => {
    //         cy.inviteUser('user-1');

    //         cy.apiLogin('user-1');
    //         cy.visit(channel);
    //         cy.verifyExportCommandIsAvailable();

    //         cy.apiLogin('user-2');
    //         cy.kickUser('user-1');

    //         cy.apiLogin('user-1');
    //         cy.visit(channel);
    //         cy.verifyChannelDoesNotExist(channel);

    //         cy.apiExportChannel(channel);
    //         cy.verifyNoExport();
    //     });
    // });

    // it('ID 18 - User can export a read-only channel', () => {
    //     cy.apiCreateReadOnlyChannel().then((channel) => {
    //         ay.apiExportChannel(channel);
    //         cy.verifySuccessfulExport();
    //     });

    // });
});
