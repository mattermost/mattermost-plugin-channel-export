// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import manifest from './manifest';

export default class Plugin {
    // eslint-disable-next-line no-unused-vars
    initialize(registry, store) {
        // @see https://developers.mattermost.com/extend/plugins/webapp/reference/

        // The webapp fails to emit an ActionTypes.RECEIVED_WEBAPP_PLUGIN when the plugin is
        // activated after the page has already loaded. Emit one ourselves to simplify integration
        // while we wait for the webapp to be fixed.
        store.dispatch({type: 'RECEIVED_WEBAPP_PLUGIN', data: manifest});
    }
}

window.registerPlugin(manifest.id, new Plugin());
