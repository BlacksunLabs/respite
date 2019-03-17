# respite
Read-only Slack RTM API client for spying on teams.

Respite is a terminal based, read-only client for Slack’s RTM API which authenticates using auth tokens to bypass 2FA and SSO.

Respite was developed to provide a useful tool for leveraging tokens acquired from the `toke_em` tool by @n0ncetonic https://github.com/n0ncetonic/toke_em

TODO : Write a better readme ...

![](https://user-images.githubusercontent.com/29786827/54484700-dae76400-4828-11e9-9d53-37111a95ebfe.png)
## Roadmap
**Current stable version:** _1.0_

**Current dev version:** _1.0_

### v1.0
- Utilize legacy auth tokens to authenticate with Slack's RTM API
- Receive messages posted into any channel/dm which your user has permissions to
- Allow filtering of messages by channel/dm

### v1.1
- Supports downloading files uploaded to team
- Adds functionality to export team user profile directory
- Adds function to export a channel's history
