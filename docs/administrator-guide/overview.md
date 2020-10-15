# Overview

This plugin allows channel export into a human readable format. The Mattermost Incident Response plugin supports channel export. The following artifacts are included in the `.CSV` output file:

- Messages
- Stages
- Tasks

## Getting started

Clone the repository:

```
git clone https://github.com/mattermost/mattermost-plugin-channel-export.git
```

Build the plugin:

```
make
```

This will produce a single plugin file (with support for multiple architectures) for upload to your Mattermost server:

```
dist/com.mattermost.plugin-channel-export.tar.gz
```
