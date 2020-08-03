// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

declare namespace Cypress {

    // We cannot use normal imports; otherwise, we would convert this to a
    // normal module, and we want it to be an ambient module in order to merge
    // our declaration of Cypress with the global one.
    // See https://stackoverflow.com/a/51114250/3248221 for more information
    type Channel = import('mattermost-redux/types/channels').Channel;
    type Team = import('mattermost-redux/types/teams').Team;
    type UserProfile = import('mattermost-redux/types/users').UserProfile;
    type FileFormat = import('./ui_commands').FileFormat;

    interface Chainable<Subject> {

        /**
         * User login directly via API
         * @param {String} username - username
         * @param {String} password - password
         * @return A Chainable-wrapped Response
        */
        apiLogin(username?: string, password?: string | null): Chainable<Response>;
        apiCreatePublicChannel(teamId: string, name: string, displayName: string): Chainable<Channel>;
        apiCreatePrivateChannel(teamId: string, name: string, displayName: string): Chainable<Channel>;
        apiCreateGroupMessage(userIds: string[]): Chainable<Channel>;
        apiCreateDirectMessage(userIds: string[]): Chainable<Channel>;
        apiGetTeamByName(name: string): Chainable<Team>;
        apiGetUserByUsername(username: string): Chainable<UserProfile>;
        apiGetUsers(usernames: string[]): Chainable<UserProfile[]>;

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

        apiMakeChannelReadOnly(channelId: string): Chainable<Response>;
        apiExportChannel(channelId: string, expectedStatus?: number): Chainable<string>;

        archiveCurrentChannel(): Chainable<void>;
        unarchiveCurrentChannel(): Chainable<void>;

        leaveCurrentChannel(): Chainable<void>;

        inviteUser(userName: string): Chainable<void>;
        kickUser(userName: string): Chainable<void>;

        apiCreatePost(channelId: string, message: string, fileIds?: string[]): Chainable<void>;
        apiCreateMultiplePosts(channelId: string, numMessages: number): Chainable<void>;
        apiGetChannelByName(teamName: string, channelName: string) : Chainable<Channel>;

        verifyExportBotMessage(channelDisplayName: string): Chainable<void>;

        verifyExportCommandIsAvailable(): Chainable<void>;
        verifyExportSystemMessage(channelDisplayName: string): Chainable<void>;

        verifyFileCanBeDownloaded(channelDisplayName: string): Chainable<void>;

        verifyFileName(fileFormat: FileFormat, channelDisplayName: string, channelName: string): Chainable<void>;

        verifyNoPosts(channelDisplayName: string): Chainable<Channel>;
        verifyAtLeastPosts(channelDisplayName: string, numPosts: number): Chainable<Channel>;

        visitNewDirectMessage(creatorName: string, otherName: string): Chainable<Channel>;
        visitNewGroupMessage(userNames: string[]): Chainable<Channel>;

        visitNewPublicChannel(): Chainable<Channel>;
        visitNewPrivateChannel(): Chainable<Channel>;
        visitDMWithBot(userName: string, botName?: string): Chainable<Channel>;
    }
}
