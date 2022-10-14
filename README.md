# How to use this image
`docker run --name w2g-bot -d -e BOT_TOKEN=<your discord bot token> jonashoen/w2g-bot`

**You can also use compose:**

```
services:
  w2g:
    image: jonashoen/w2g-bot
    restart: always
    container_name: w2g-bot
    environment:
      - BOT_TOKEN=${BOT_TOKEN}
```

## Bot commands
`!w2g` creates an empty watch2gether room  
`!w2g <YouTube link>` creates a watch2gether room with the provided YouTube video playing

The bot will send the link to the created rooms in the same text channel the commands were sent.
