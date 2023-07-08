# SaltBot2.0

This is a fun discord bot written in ~~JavaScript~~ ~~Python~~ Go (yes this is the 3rd time I've re-implemented this bot). To see a list of commands, you can either look at the [handler.go file](./handler/handler.go) or type `!help` in a channel that SaltBot is listening to. To add SaltBot to a server, contact me at `davidgreeson13@gmail.com` for an [OAuth2 url](https://discordpy.readthedocs.io/en/latest/discord.html).

## Saltbot Prerequisites

To run Saltbot, you need to have 3 tokens:
 - [`BOT_TOKEN`](https://discordpy.readthedocs.io/en/latest/discord.html) - A discord developer bot token for connecting to the discord servers.
 - [`YOUTUBE_AUTH`](https://developers.google.com/youtube/v3/getting-started) - A YouTube API token for retrieving YouTube videos.
 - [`GIPHY_AUTH`](https://developers.giphy.com) - A Giphy API token for retrieving Giphy gifs.

Create each of those tokens, and then create an auth file called `auth.env` like below but with your tokens:
```
BOT_TOKEN=<YOUR-BOT-TOKEN>
GIPHY_AUTH=<YOUR-GIPHY-AUTH>
YOUTUBE_AUTH=<YOUR-YOUTUBE-AUTH>
```

## Running SaltBot

Run saltbot directly with the golang interpreter:
```bash
go run saltbot.go
```

Saltbot will first attempt to reach out to a kubernetes server using a kubernetes Service Account. If it can't, then it will attemp to use `~/.kube/config`.

### Running SaltBot in a Kubernetes Cluster

I published SaltBot on a public docker hub repo at `highsaltlevels/saltbot`. If you would like to deploy this into a kubernetes cluster, you're free to use the namespace and deployment files in the `k8s` folder.

1. Replace the placeholders with actual credentials

```bash
sed -i s/__BOT_TOKEN__/<YOUR-BOT-TOKEN>/g k8s/deployment.yaml
sed -i s/__GIPHY_AUTH__/<YOUR-GIPHY-AUTH>/g k8s/deployment.yaml
sed -i s/__YOUTUBE_AUTH__/<YOUR-YOUTUBE-AUTH>/g k8s/deployment.yaml
```

2. (Optional) Create a `saltbot` Namespace

```bash
kubectl create -f k8s/namespace.yaml
```

3. (Optional) Log into Dockerhub to Avoid Anonymous Pull Rate Limiting

```bash
docker login -u <username> -p <password>
kubectl -n saltbot create secret generic regcred --from-file=.dockerconfigjson=/path/to/.docker/config.json --type=kubernetes.io/dockerconfigjson
```

4. Deploy Saltbot

```bash
kubectl -n saltbot apply -f k8s/deployment.yaml
```
