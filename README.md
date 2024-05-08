# Mattermost Channel Export Plugin

This plugin exports playbook channels into a CSV format.

![image](https://github.com/mattermost/mattermost-plugin-channel-export/assets/74422101/2b3fd0bd-75e3-4ae4-8a3c-251a215348a4)

See the [Mattermost Product Documentation](https://docs.mattermost.com/comply/export-mattermost-channel-data.html#export-channel-data) for details on installing, configuring, enabling, and using this Mattermost integration.

## Development

### Get started

1. Clone the repository: `git clone https://github.com/mattermost/mattermost-plugin-channel-export.git`.
2. Build the plugin using `make`.
3. This will produce a single plugin file (with support for multiple architectures) for upload to your Mattermost server: `dist/com.mattermost.plugin-channel-export.tar.gz`.
4. To avoid having to manually install your plugin, build and deploy your plugin using one of the following options.

### Deploy with local mode

If your Mattermost server is running locally, you can enable [local mode](https://docs.mattermost.com/manage/mmctl-command-line-tool.html#local-mode) to streamline deploying your plugin. Edit your server configuration as follows:

```json
{
    "ServiceSettings": {
        ...
        "EnableLocalMode": true,
        "LocalModeSocketLocation": "/var/tmp/mattermost_local.socket"
    }
}
```

Deploy your plugin with ``make deploy``.

You may also customize the Unix socket path:

```bash
export MM_LOCALSOCKETPATH=/var/tmp/alternate_local.socket
make deploy
```

If developing a plugin with a webapp, watch for changes and deploy those automatically using ``make watch``.

### Deploy with credentials

Alternatively, you can authenticate with the server's API with credentials:

```bash
export MM_SERVICESETTINGS_SITEURL=http://localhost:8065
export MM_ADMIN_USERNAME=admin
export MM_ADMIN_PASSWORD=password
make deploy
```

or with a [personal access token](https://developers.mattermost.com/integrate/reference/personal-access-token/):

```bash
export MM_SERVICESETTINGS_SITEURL=http://localhost:8065
export MM_ADMIN_TOKEN=j44acwd8obn78cdcx7koid4jkr
make deploy
```

## License

This repository is licensed under the [Mattermost Source Available License](LICENSE) and requires a valid Enterprise Edition E20 license when used for production. See [frequently asked questions](https://docs.mattermost.com/overview/faq.html#mattermost-source-available-license) to learn more.

Although Mattermost Enterprise is required if using this plugin in production, the [Mattermost Source Available License](LICENSE) allows you to compile and test this plugin in development and testing environments without Mattermost Enterprise. As such, we welcome community contributions to this plugin.

On startup, the plugin checks for a valid Mattermost Enterprise license. If you're running an Enterprise Edition of Mattermost and don't already have a valid license, you can obtain a trial license from **System Console > Edition and License**. If you're running the Team Edition of Mattermost, including when you run the server directly from source, you may instead configure your server to enable both testing (`ServiceSettings.EnableTesting`) and developer mode (`ServiceSettings.EnableDeveloper`). These settings are not recommended in production environments.

## Help and support

This plugin contains both a server and web app portion. Read our documentation about the [Developer Workflow](https://developers.mattermost.com/extend/plugins/developer-workflow/) and [Developer Setup](https://developers.mattermost.com/extend/plugins/developer-setup/) for more information about developing and extending plugins.

To report a bug, please open a GitHub issue.
