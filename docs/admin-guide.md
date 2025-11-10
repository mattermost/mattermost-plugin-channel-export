# Admin Guide

The Admin Guide provides instructions for installing, configuring, and managing the Mattermost Channel Export Plugin in different environments.

---

## Table of Contents
- [Prerequisites](#prerequisites)
- [Installation](#installation)
  - [Marketplace](#marketplace)
  - [Manual](#manual)
  - [Cloud](#cloud)
  - [Upgrade](#upgrade)
- [Configuration](#configuration)
  - [Deploy with local mode](#deploy-with-local-mode)
  - [Deploy with credentials](#deploy-with-credentials)
  - [Personal Access Token](#personal-access-token)
- [Webhooks](#webhooks)
- [Slash Commands](#slash-commands)
- [Onboard Users](#onboard-users)
- [FAQ](#faq)
- [Get Help](#get-help)

---

## Prerequisites
- A running Mattermost server (Team Edition or Enterprise Edition).
- Admin access to the System Console.
- For production use: a valid Enterprise Edition E20 license.
- Optional: Marketplace access if installing via Marketplace.

---

## Installation

### Marketplace
1. Go to **System Console → Plugins → Marketplace**.
2. Search for **Channel Export Plugin**.
3. Click **Install**, then **Enable**.

### Manual
1. Build or download the plugin package:
   ```bash
   make

This produces :
 ```bash
  dist/com.mattermost.plugin-channel-export.tar.gz
```
2. Go to System Console → Plugins → Management → Upload Plugin.
3. Upload the .tar.gz file and click Enable.

### Cloud
- For Mattermost Cloud, use the Marketplace flow.
- If restricted, contact Mattermost support to enable the plugin.

### Upgrade
1. (Optional) Disable the plugin temporarily.
2. Upload the new version via Marketplace or Manual upload.
3. Re‑enable and verify settings.
4. Test export on a staging environment before production.

---

## Configuration
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

You may also customize the Unix socket path if needed:

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

### or with a [personal access token](https://developers.mattermost.com/integrate/reference/personal-access-token/):

```bash
export MM_SERVICESETTINGS_SITEURL=http://localhost:8065
export MM_ADMIN_TOKEN=j44acwd8obn78cdcx7koid4jkr
make deploy
```

---

## Webhooks
If the plugin uses webhooks (optional), configure them in System Console → Integrations.

- Provide endpoint URLs and secrets.
- Test with sample events to confirm delivery.

---

## Slash Commands
- Ensure slash commands are enabled in System Console → Integrations.
- Admins can restrict or allow commands per team.
- Example: `/export <options>`.

---

## Onboard Users
- Announce plugin availability to your teams.
- Share export policies and compliance expectations.
- Provide examples of when and how to use exports.

---

## FAQ
- Why is export not available in some channels? 
  Check permissions and plugin settings.

- Where are exports stored? 
    Typically as downloadable CSV files; retention depends on your policy.

- How do I test locally? 
    Use Local Mode or Credentials with make deploy.

---

## Get Help

- **Developer Workflow:** [Mattermost Plugin Developer Workflow](https://developers.mattermost.com/extend/plugins/developer-workflow/)  
  Learn how to build, extend, and maintain Mattermost plugins.

- **Developer Setup:** [Plugin Developer Setup Guide](https://developers.mattermost.com/extend/plugins/developer-setup/)  
  Step‑by‑step instructions for setting up your development environment.

- **Product Documentation:** [Export Channel Data](https://docs.mattermost.com/administration-guide/comply/export-mattermost-channel-data.html)  
  Official Mattermost documentation on exporting channel data.

- **Report Issues:**  
  To report a bug or request a feature, please open a GitHub issue in this repository.

- **Community & Support:**  
  Join the Mattermost community forums or contact Mattermost support if you need additional help.
