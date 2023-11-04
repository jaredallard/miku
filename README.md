# miku

A Discord bot to convert media links (e.g., Spotify, Apple Music)
between each other.

## Usage

While each provider has its own authenticiation requirements, the bot
will at minimum require a Discord bot. To create one, follow the steps
below:

1. Create a new Discord bot with the following scopes:
 - `bot`
 - `Send Messages`
 - `Read Messages`
 - Enable the `Message Content` toggle under `Privileged Gateway Intents`
2. Invite the bot to your server using the following URL (change the
   client ID to your bot's client ID).
    ```
    https://discord.com/api/oauth2/authorize?client_id=<client_id>&permissions=3072&scope=bot
    ```
3. Generate a Bot Token and take note of it.

Set the following environment variables:

```bash
MIKU_DISCORD_TOKEN="<Discord Bot Token From Step 3>"
# Optional: Limit to single channel.
MIKU_DISCORD_CHANNEL_ID="<Discord Channel ID>"
```

## Enabling Providers

Below is specific instructions/requirements for a provider to be
enabled.

### Spotify

1. Create a new Spotify app following the instructions
   [here](https://developer.spotify.com/documentation/general/guides/app-settings/#register-your-app).
2. Take note of the Client ID and Client Secret.

Set the following environment variables:

```bash
MIKU_SPOTIFY_CLIENT_ID="<Client ID>"
MIKU_SPOTIFY_CLIENT_SECRET="<Client Secret>"
```

### Apple Music

**Note**: Currently the API token will expire every 6 months and need
to be regenerated. This will eventually be automated.

1. Create a bew media identifier following the instructions
   [here](https://developer.apple.com/help/account/configure-app-capabilities/create-a-media-identifier-and-private-key/).
2. Ensure you downloaded a `.p8` and have your Team ID and Key ID ready.
3. Run the following command to generate a token:
    ```bash
    go run github.com/minchao/go-apple-music/examples/token-generator@latest \
      -l 15777000 -t "<Team_ID>" -pf "$HOME/Downloads/AuthKey_<Key_ID>.p8" \
      -k "<Key_ID>"
    ```

Set the following environment variables:

```bash
MIKU_APPLE_MUSIC_API_TOKEN="<Generated Token From Step 3>"
```

## Development

Setup a `.env.development` using the provider documentation above.
Reference `.env.example` to see all available options.

Export env vars from `.env.development`:

```bash
set -o allexport && source .env.development && set +o allexport
```

## License

GPL-3.0-only
