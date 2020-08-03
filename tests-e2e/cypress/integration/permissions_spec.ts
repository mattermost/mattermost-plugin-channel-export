// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
/// <reference path="../support/index.d.ts" />

import '@testing-library/cypress/add-commands';
import {Channel} from 'mattermost-redux/types/channels';
import {UserProfile} from 'mattermost-redux/types/users';

import {httpStatusNotFound} from '../support/constants';

describe('Test Area - Permissions', () => {
    const fileHeader =
    'Post Creation Time,User Id,User Email,User Type,User Name,Post Id,Parent Post Id,Post Message,Post Type';

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
            cy.verifyFileCanBeDownloaded(channel.display_name);
        });
    });

    it('ID 10 - User can export a private channel', () => {
        cy.visitNewPrivateChannel().then((channel: Channel) => {
            cy.exportSlashCommand();
            cy.visitDMWithBot('user-1');
            cy.verifyFileCanBeDownloaded(channel.display_name);
        });
    });

    it('ID 11 - User can export a group message channel', () => {
        const userNames = ['user-1', 'aaron.medina', 'aaron.peterson'];
        cy.visitNewGroupMessage(userNames).then((channel: Channel) => {
            cy.exportSlashCommand();
            cy.visitDMWithBot('user-1');
            cy.verifyFileCanBeDownloaded(channel.name);
        });
    });

    it('ID 12 - User can export a direct message channel', () => {
        cy.visitNewDirectMessage('user-1', 'anne.stone').then((channel: Channel) => {
            cy.exportSlashCommand();
            cy.visitDMWithBot('user-1');
            cy.verifyFileCanBeDownloaded(channel.name);
        });
    });

    it('ID 13 - User can export a direct message channel with self', () => {
        cy.apiGetUserByUsername('user-1').then((user: UserProfile) => {
            cy.visit('/ad-1/messages/@user-1');
            cy.exportSlashCommand();
            cy.visitDMWithBot('user-1');
            cy.verifyFileCanBeDownloaded(`${user.id}__${user.id}`);
        });
    });

    it('ID 14 - User can export a bot message channel', () => {
        cy.visitDMWithBot('user-1');
        cy.exportSlashCommand();

        cy.apiGetUserByUsername('user-1').then((user: UserProfile) => {
            cy.apiGetUserByUsername('channelexport').then((bot: UserProfile) => {
                cy.verifyFileCanBeDownloaded(`${user.id}__${bot.id}`);
            });
        });
    });

    it('ID 15 - User can export archived channel', () => {
        cy.visitNewPublicChannel().then((channel: Channel) => {
            cy.archiveCurrentChannel();
            cy.apiExportChannel(channel.id).then((fileContents: string) => {
                expect(fileContents).to.contain(fileHeader);
            });
        });
    });

    it('ID 16 - User can export an unarchived channel', () => {
        cy.visitNewPublicChannel().then((channel: Channel) => {
            cy.archiveCurrentChannel();

            cy.apiLogin('sysadmin').then(() => {
                cy.visit(`/ad-1/channels/${channel.name}`);
                cy.unarchiveCurrentChannel();
            });

            cy.apiLogin('user-1').then(() => {
                cy.visit(`/ad-1/channels/${channel.name}`);
                cy.exportSlashCommand();
                cy.visitDMWithBot('user-1');
                cy.verifyFileCanBeDownloaded(channel.display_name);
            });
        });
    });

    it('ID 15 (2nd one) - User cannot export a channel they are not added to', () => {
        cy.visitNewPublicChannel().then((channel) => {
            cy.leaveCurrentChannel();
            cy.apiExportChannel(channel.id, 404);
        });
    });

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
