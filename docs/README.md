# deploy-vultr — docs

**Vultr deploy.** Provision a Vultr instance (cloud-init Docker) and run the app image.

## Install

```bash
togo install togo-framework/deploy-vultr
```

Registers on the [`deploy`](https://github.com/togo-framework/deploy) base; select it with **deploy.provider in togo.yaml (or DEPLOY_PROVIDER)**, then use **`togo deploy`**.

## Interface

`Deployer` — `Provision`/`Deploy`/`Destroy`/`Status` over a `Spec{App,Dir,BuildCmd,Host,User,Image,Region,Domain}` built from your `togo.yaml`.

## Configuration

| Env var | Description |
|---|---|
| `VULTR_API_KEY` | Vultr API key (required). |

## Usage & notes

Uses govultr to create an instance whose cloud-init runs `spec.Image`. `Destroy` deletes it.

## Example

```bash
togo deploy --provider vultr --dry-run   # preview the plan
togo deploy --provider vultr
```

## Links

- [govultr](https://github.com/vultr/govultr)
- [Marketplace](https://to-go.dev/marketplace)
- [Source](https://github.com/togo-framework/deploy-vultr)
