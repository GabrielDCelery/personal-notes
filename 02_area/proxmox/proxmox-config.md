# Proxmox configuration

This documentation is heavily influenced by [Christian Lempa's Don't run Proxmox without these settings!](https://www.youtube.com/watch?v=VAJWUZ3sTSI&ab_channel=ChristianLempa) video

## Configuring updates

> [!NOTE]
> During the configuration there will be pop-ups saying that `You do not have a valid subscription for this server. Please visit www.proxmox.com to get a list of available options.`, you can ignore that

1. Go to the `Updates` -> `Repositories` section

2. `Disable` every repostory that has the name `enterprise` in them.

3. Using the `Add` button add a `No-Subscription` repository.

> [!NOTE]
> After adding the `No_Subscription` repository you will see a `pve-no-subscription` warning, you can ignore that

4. Go to `Updates` -> `Refresh` and that will check all the configured repositories if there are updates for them.

5. Go to any of the Proxmox nodes that you wish to update and under `Updates` click on `Upgrade`

> [!IMPORTANT]
> You have to do the updates as `root`, this might be important if you are using an identity provider that might not have the necessary permissions

5. (alternative) You can also SSH into the Proxmox server and run `apt update && apt dist-upgrade`

## Enable notifications
