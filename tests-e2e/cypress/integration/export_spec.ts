// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
/// <reference path="../support/index.d.ts" />

import '@testing-library/cypress/add-commands';
import {Channel} from 'mattermost-redux/types/channels';

import {FileFormat} from '../support/ui_commands';

describe('Test Area - Export', () => {
    beforeEach(() => {
        // # Login as non-admin user
        cy.apiLogin('user-1');

        // # Visit the default channel
        cy.visit('/');
    });

    it('ID 19 - A system message notifies of successful export command execution on the channel where export is initiated', () => {
        cy.visitNewPublicChannel().then((channel: Channel) => {
            cy.exportSlashCommand();
            cy.verifyExportSystemMessage(channel);
        });
    });

    it('ID 20 - A bot message notifies of a successful export', () => {
        cy.visitNewPublicChannel().then((channel: Channel) => {
            cy.exportSlashCommand();
            cy.visitDMWithBot('user-1');
            cy.verifyExportBotMessage(channel);
        });
    });

    it('ID 21 - The exported file can be downloaded locally', () => {
        cy.visitNewPublicChannel().then((channel: Channel) => {
            cy.exportSlashCommand();
            cy.visitDMWithBot('user-1');
            cy.verifyFileCanBeDownloaded(channel);
        });
    });

    // // it('ID 22 - Channel is exported in CSV file format', () => {
    // //     cy.visitNewPublicChannel();
    // //     cy.exportSlashCommand();
    // //     cy.verifyFileExtension('csv');
    // // });

    it('ID 23 - Exported CSV filename has [channel-name].csv format', () => {
        cy.visitNewPublicChannel().then((channel) => {
            cy.exportSlashCommand();
            cy.visitDMWithBot('user-1');
            cy.verifyFileName(FileFormat.CSV, channel);
        });
    });

    // // it('ID 24 - A bot message notifies of an unsuccessful export', () => {
    // //     cy.visitNewPublicChannel();
    // // });

    // // it('ID 25 - Exported CSV has messages of a channel in chronological order', () => {
    // //     cy.visitNewPublicChannel();
    // // });

    // // it('ID 26 - Exported CSV has date', () => {
    // //     cy.visitNewPublicChannel();
    // // });

    // // it('ID 27 - Exported CSV has timestamp', () => {
    // //     cy.visitNewPublicChannel();
    // // });

    // // it('ID 28 - Exported CSV has message senders username', () => {
    // //     cy.visitNewPublicChannel();
    // // });

    it('ID 29 - A channel with no messages can be exported successfully', () => {
        cy.visitNewPublicChannel().then((channel) => {
            cy.verifyNoPosts();
            cy.exportSlashCommand();
            cy.visitDMWithBot('user-1');
            cy.verifyFileCanBeDownloaded(channel);
        });
    });

    // it('ID 30 - A channel with more than 100 messages can be exported successfully', () => {
    //     cy.visitNewPublicChannel();

    //     const numMessages = 150;
    //     cy.postMessages(numMessages);
    //     cy.exportSlashCommand();
    //     cy.verifySuccessfulExport();
    // });

    // // it('ID 31 - A channel with media files can be exported successfully', () => {
    // //     cy.visitNewPublicChannel();
    // // });
});
