# Mattermost Channel Export Plugin

This plugin allows channel export into a human readable format.

## Getting Started

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

## Development

To avoid having to manually install your plugin, build and deploy your plugin with login credentials:
```
export MM_SERVICESETTINGS_SITEURL=http://localhost:8065
export MM_ADMIN_USERNAME=admin
export MM_ADMIN_PASSWORD=password
make deploy
```

or with a [personal access token](https://docs.mattermost.com/developer/personal-access-tokens.html):
```
export MM_SERVICESETTINGS_SITEURL=http://localhost:8065
export MM_ADMIN_TOKEN=j44acwd8obn78cdcx7koid4jkr
make deploy
```
