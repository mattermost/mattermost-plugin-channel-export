// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

declare namespace Cypress {

    // We cannot use normal imports; otherwise, we would convert this to a
    // normal module, and we want it to be an ambient module in order to merge
    // our declaration of Cypress with the global one.
    // See https://stackoverflow.com/a/51114250/3248221 for more information
    type Channel = import('mattermost-redux/types/channels').Channel;
    type Team = import('mattermost-redux/types/teams').Team;

    interface Chainable<Subject> {

        /**
         * User login directly via API
         * @param {String} username - username
         * @param {String} password - password
         * @return A Chainable-wrapped Response
        */
        apiLogin(username?: string, password?: string | null): Chainable<Response>;
        apiCreatePublicChannel(teamId: string, name: string, displayName: string): Chainable<Channel>;
        apiGetTeamByName(name: string): Chainable<Team>;

        /**
         * Post a message in the current channel
         * @param {String} message - the string to post as a message in the current
         * channel
        */
        postMessage(message: string): Chainable<void>;

        /**
         * Post a message in the current channel
         * @param {String} message - the string to post as a message in the current
         * channel
        */
        exportSlashCommand(): Chainable<Element>;

        getLastPostId(): Chainable<string>;

        // apiCreateReadOnlyChannel(): Chainable<Element>;
        // apiExportChannel(): Chainable<Element>;

        // archiveChannel(): Chainable<Element>;
        // unarchiveChannel(): Chainable<Element>;
        // leaveChannel(): Chainable<Element>;

        // inviteUser(): Chainable<Element>;
        // kickUser(): Chainable<Element>;

        // postMessages(): Chainable<Element>;

        // verifyChannelDoesNotExist(): Chainable<Element>;
        // verifyExportBotMessage(): Chainable<Element>;
        // verifyExportCommandIsAvailable(): Chainable<Element>;
        verifyExportSystemMessage(channelName: string): Chainable<Element>;

        // verifyFileCanBeDownloaded(): Chainable<Element>;
        // verifyFileExtension(): Chainable<Element>;
        // verifyFileName(): Chainable<Element>;
        // verifyNoExport(): Chainable<Element>;
        // verifyNoPosts(): Chainable<Element>;
        // verifySuccessfulExport(): Chainable<Element>;

        // visitNewDirectMessage(): Chainable<Element>;
        // visitNewGroupMessage(): Chainable<Element>;
        // visitNewPrivateChannel(): Chainable<Element>;
        visitNewPublicChannel(): Chainable<Channel>;

        // visitSelfDM(): Chainable<Element>;
    }
}