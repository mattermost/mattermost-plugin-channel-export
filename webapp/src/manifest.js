// This file is automatically generated. Do not modify it manually.

const manifest = JSON.parse(`
{
    "id": "com.mattermost.plugin-channel-export",
    "name": "Channel Export",
    "description": "This plugin allows channel export into a human readable format.",
    "homepage_url": "https://github.com/mattermost/mattermost-plugin-channel-export/",
    "support_url": "https://github.com/mattermost/mattermost-plugin-channel-export/issues",
    "release_notes_url": "https://github.com/mattermost/mattermost-plugin-channel-export/releases/tag/v1.0.0",
    "version": "1.0.0",
    "min_server_version": "5.37.0",
    "server": {
        "executables": {
            "darwin-amd64": "server/dist/plugin-darwin-amd64",
            "darwin-arm64": "server/dist/plugin-darwin-arm64",
            "linux-amd64": "server/dist/plugin-linux-amd64",
            "linux-arm64": "server/dist/plugin-linux-arm64",
            "windows-amd64": "server/dist/plugin-windows-amd64.exe"
        },
        "executable": ""
    },
    "webapp": {
        "bundle_path": "webapp/dist/main.js"
    },
    "settings_schema": {
        "header": "",
        "footer": "",
        "settings": [
            {
                "key": "EnableAdminRestrictions",
                "display_name": "Enable Admin Restrictions",
                "type": "bool",
                "help_text": "Restricts the exporting of channels to system administrators or channel administrators",
                "placeholder": "",
                "default": false
            },
            {
                "key": "MaxFileSize",
                "display_name": "Maximum size of channel export file in bytes",
                "type": "number",
                "help_text": "Determines the maximum size of the channel export file when using the slash command. A value of 0 will use the [FileSettings.MaxFileSize](https://docs.mattermost.com/configure/environment-configuration-settings.html#maximum-file-size) from Mattermost server.",
                "placeholder": "",
                "default": 0
            }
        ]
    }
}
`);

export default manifest;
export const id = manifest.id;
export const version = manifest.version;
