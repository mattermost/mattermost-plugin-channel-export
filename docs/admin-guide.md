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
- [Webhooks](#webhooks)
- [Slash Commands](#slash-commands)
- [Onboard Users](#onboard-users)
- [FAQ](#faq)


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