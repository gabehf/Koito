<div align="center">
  
![Koito logo](https://github.com/user-attachments/assets/bd69a050-b40f-4da7-8ff1-4607554bfd6d)

*Koito (小糸) is a Japanese surname. It is also homophonous with the words 恋と (koi to), meaning "and/with love".*

</div>

<div align="center">
  
  [![Ko-Fi](https://img.shields.io/badge/Ko--fi-F16061?style=for-the-badge&logo=ko-fi&logoColor=white)](https://ko-fi.com/gabehf)
  
</div>

Koito is a modern, themeable ListenBrainz-compatible scrobbler for self-hosters who want control over their data and insights into their listening habits. 
It supports relaying to other compatible scrobblers, so you can try it safely without replacing your current setup.

> This project is under active development and still considered "unstable", and therefore you can expect some bugs. If you don't want to replace your current scrobbler
with Koito quite yet, you can [set up a relay](https://koito.io/guides/scrobbler/#set-up-a-relay) from Koito to another ListenBrainz-compatible
scrobbler. This is what I've been doing for the entire development of this app and it hasn't failed me once. Or, you can always use something
like [multi-scrobbler](https://github.com/FoxxMD/multi-scrobbler).

## Features

- ⚡ More performant than similar software
- 🖌️ Sleek UI
- 🔁 Compatible with anything that scrobbles to ListenBrainz
- 🔌 Easy relay to your existing setup
- 📂 Import support for Maloja, ListenBrainz, LastFM, and Spotify

## Demo

You can view my public instance with my listening data at https://koito.mnrva.dev

## Screenshots

![screenshot one](assets/screenshot1.png)
<img width="2021" height="1330" alt="image" src="https://github.com/user-attachments/assets/956748ff-f61f-4102-94b2-50783d9ee72b" />
<img width="1505" height="1018" alt="image" src="https://github.com/user-attachments/assets/5f7e1162-f723-4e4b-a528-06cf26d1d870" />


## Installation

See the [installation guide](https://koito.io/guides/installation/), or, if you just want to cut to the chase, use this docker compose file:

```yaml
services:
  koito:
    image: gabehf/koito:latest
    container_name: koito
    depends_on:
      - db
    environment:
      - KOITO_DATABASE_URL=postgres://postgres:secret_password@db:5432/koitodb
      - KOITO_ALLOWED_HOSTS=koito.example.com,192.168.0.100:4110
    ports:
      - "4110:4110"
    volumes:
      - ./koito-data:/etc/koito
    restart: unless-stopped

  db:
    image: postgres:16
    container_name: psql
    restart: unless-stopped
    environment:
      POSTGRES_DB: koitodb
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: secret_password
    volumes:
      - ./db-data:/var/lib/postgresql/data
```

Be sure to replace `secret_password` with a random password of your choice, and set `KOITO_ALLOWED_HOSTS` to include the domain name or IP address you will be accessing Koito 
from when using either of the Docker methods described above. You should also change the default username and password, if you didn't configure custom defaults.

## Importing Data

See the [data importing guide](https://koito.io/guides/importing/) in the docs.

## Full list of configuration options

See the [configuration reference](https://koito.io/reference/configuration/) in the docs.

## Contributing

Issues and pull requests to find and fix bugs are always appreciated! If you have any feature ideas, open a GitHub issue to let me know. I'm sorting through ideas to decide which data visualizations and customization options to add next.

Also consider supporting one of the **feature bounties**! Just mention one of the bounty numbers in a Ko-fi donation to have your donation counted towards the goal. Note that bounties are first and foremost a donation to me (@gabehf) in support of building Koito, and reaching a bounty does not
necessarily mean the feature will be made, or made exactly in the way the bounty describes. This is just a way for me to see exactly what
features people are most interested in having, as well as affording me more time to work on the project and get new features out for free for everybody because unfortunately, rent is due on the first.

*Bounties are updated by me manually so it may take some time for your donation to be reflected here. Dollar values are in USD.*

### Bounty #11 - Multi-User Support & OIDC - $120 out of $500 reached

Multi-user support would allow for one Koito instance to track and display statistics for multiple users instead of just one.
New features could include global rewinds, global home page statistics, per-user relays, user pages with artist overlap % with you, etc.

**Why a bounty?** Enabling multi-user support will require very significant refactoring, testing, and restructuring of many parts of the code, in addition to adding the new features.
All future features will also have to take into account user-scope, as well as user roles and priviledges.

![](https://geps.dev/progress/24)

## Star History

<a href="https://www.star-history.com/#gabehf/koito&type=date&legend=top-left">
 <picture>
   <source media="(prefers-color-scheme: dark)" srcset="https://api.star-history.com/svg?repos=gabehf/koito&type=date&theme=dark&legend=top-left" />
   <source media="(prefers-color-scheme: light)" srcset="https://api.star-history.com/svg?repos=gabehf/koito&type=date&legend=top-left" />
   <img alt="Star History Chart" src="https://api.star-history.com/svg?repos=gabehf/koito&type=date&legend=top-left" />
 </picture>
</a>

## Albums that fueled development + notes

More relevant here than any of my other projects...

Not just during development, you can see my complete listening data on my [live demo instance](https://koito.mnrva.dev).

#### Random notes

- I find it a little annoying when READMEs use emoji but everyone else is doing it so I felt like I had to...
- About 50% of the reason I built this was minor/not-so-minor greivances with Maloja. Could I have just contributed to Maloja? Maybe, but I like building stuff and I like Koito's UI a lot more anyways.
