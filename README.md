# miku

A Discord bot to convert media links (e.g., Spotify, Apple Music)
between each other.

## Usage

1. Create a new Discord bot with the following scopes:
 - `bot`
 - `Send Messages`
 - `Read Messages`
2. Invite the bot to your server using the following URL (change the
   client ID to your bot's client ID):
   <https://discord.com/api/oauth2/authorize?client_id=<client_id>&permissions=3072&scope=bot>

Apple Music:

**TODO**

```bash
go run \
  github.com/minchao/go-apple-music/examples/token-generator@latest \
  -l 15777000 -t "<Team_ID>" -pf "$HOME/Downloads/AuthKey_<Key_ID>.p8" \
  -k "<Key_ID>"
```

## Development

Export env vars from `.env.development`:

```bash
set -o allexport && source .env.development && set +o allexport
```

## License

GPL-3.0-only
