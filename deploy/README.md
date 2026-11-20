# Deploy to Azure

The app runs on one Azure VM (Ubuntu 24.04, `Standard_B2ats_v2` in `eastasia`, inside the Azure free
account) with Docker Compose. GitHub Actions tests every change, builds the images, pushes them to GitHub Container Registry and
updates the VM. The design is constitution Template Exception 15.

```
push to main ──► CI (backend + frontend checks)
             ──► build api and web images ──► ghcr.io/<owner>/smart-grid-asset-registry-{api,web}:<sha>
             ──► Azure login (OIDC) ──► az vm run-command ──► deploy/remote-deploy.sh on the VM
             ──► smoke test https://<site>/api/health

VM  /opt/smart-grid
  caddy :80 :443  (only public ports; HTTPS from Let's Encrypt)
    └─ web (nginx: SPA, /api → api:8080) ─ api ─ db (PostgreSQL 16, volume db-data)
```

| File | Purpose |
|---|---|
| [`../.github/workflows/ci.yml`](../.github/workflows/ci.yml) | Checks on every pull request; on `main` the deploy workflow runs it first |
| [`../.github/workflows/deploy.yml`](../.github/workflows/deploy.yml) | Build, push and deploy on push to `main` (or by hand) |
| [`docker-compose.prod.yml`](docker-compose.prod.yml) | The production stack, copied to the VM on every deploy |
| [`Caddyfile`](Caddyfile) | HTTPS and the reverse proxy in front of `web` |
| [`remote-deploy.sh`](remote-deploy.sh) | What runs on the VM during a deploy |
| [`cloud-init.yml`](cloud-init.yml) | First boot: Docker, Compose, swap |

The VM keeps `/opt/smart-grid/.env` (mode 600). The database password is generated there on the first
deploy and never leaves the VM. No secret is stored in GitHub: Azure is reached through OIDC, and the
images are pulled with the deploy job's own short-lived token.

## One-time setup

Run from the repository root in Git Bash, with the [Azure CLI](https://learn.microsoft.com/cli/azure/)
installed and `az login` done. Adjust the names first; `DNS_LABEL` must be unique in the region.

The commands also work in Azure Cloud Shell (Bash), which is already signed in; there, set `MY_IP` by hand
to your own address, because `api.ipify.org` would return Cloud Shell's. With Cloud Shell, pass your own
public key with `--ssh-key-values` instead of `--generate-ssh-keys`, or the key is lost with the session.

```sh
LOCATION=eastasia
RG=rg-smart-grid
VM=vm-smart-grid
DNS_LABEL=smart-grid-registry
GITHUB_REPO=<owner>/<repository>
MY_IP=$(curl -fsS https://api.ipify.org)
```

### 1. The VM

```sh
az group create --name "$RG" --location "$LOCATION"

# --nsg-rule NONE: no default "SSH from anywhere" rule; the rules follow below.
az vm create \
  --resource-group "$RG" --name "$VM" \
  --image Ubuntu2404 --size Standard_B2ats_v2 \
  --admin-username azureuser --generate-ssh-keys \
  --storage-sku Premium_LRS --os-disk-size-gb 64 \
  --public-ip-sku Standard --public-ip-address-dns-name "$DNS_LABEL" \
  --nsg-rule NONE \
  --custom-data deploy/cloud-init.yml

# SkuNotAvailable means the region has no capacity for that size right now (common on free accounts):
# try the other free size (Standard_B1s) or another region with --location; the resource group
# may stay where it is. In November 2026 southeastasia had neither size; eastasia had B2ats_v2.

# HTTP and HTTPS for everyone (HTTP only redirects, and lets Let's Encrypt check the domain).
az network nsg rule create --resource-group "$RG" --nsg-name "${VM}NSG" \
  --name allow-web --priority 100 --direction Inbound --access Allow \
  --protocol '*' --destination-port-ranges 80 443 --source-address-prefixes Internet

# SSH only from your own address (for maintenance; deploys do not use SSH).
az network nsg rule create --resource-group "$RG" --nsg-name "${VM}NSG" \
  --name allow-ssh-me --priority 110 --direction Inbound --access Allow \
  --protocol Tcp --destination-port-ranges 22 --source-address-prefixes "$MY_IP/32"

SITE_ADDRESS=$(az network public-ip show --resource-group "$RG" --name "${VM}PublicIP" \
  --query dnsSettings.fqdn --output tsv)
echo "$SITE_ADDRESS"
```

### 2. Azure login for GitHub Actions (OIDC)

The deploy job may only run commands on this one VM.

```sh
APP_ID=$(az ad app create --display-name gh-smart-grid-deploy --query appId --output tsv)
az ad sp create --id "$APP_ID"

# Newer repositories send a subject with immutable IDs; copy it from the
# "subject claim" line of a failed azure/login step. The older form is
# repo:<owner>/<repository>:environment:production.
SUBJECT='repo:<owner>@<owner-id>/<repository>@<repository-id>:environment:production'

az ad app federated-credential create --id "$APP_ID" --parameters "{
  \"name\": \"github-smart-grid-production-ids\",
  \"issuer\": \"https://token.actions.githubusercontent.com\",
  \"subject\": \"$SUBJECT\",
  \"audiences\": [\"api://AzureADTokenExchange\"]
}"

az role assignment create --assignee "$APP_ID" \
  --role "Virtual Machine Contributor" \
  --scope "$(az vm show --resource-group "$RG" --name "$VM" --query id --output tsv)"
```

The subject must match what GitHub sends exactly. This repository's token carries the owner and repository
IDs (`repo:oscarsiregar@116549248/smart-grid-asset-registry@1464917673:environment:production`), so a rename
does not break the login, but a deleted and recreated repository gets a new ID and needs a new credential.
To change a subject, create a new credential and delete the old one (`az ad app federated-credential
delete`); `federated-credential update` did not change it.

### 3. GitHub

Create the `production` environment (Settings → Environments), then set the repository variables, with the
[GitHub CLI](https://cli.github.com/) or in Settings → Secrets and variables → Actions → Variables:

```sh
gh variable set AZURE_CLIENT_ID       --repo "$GITHUB_REPO" --body "$APP_ID"
gh variable set AZURE_TENANT_ID       --repo "$GITHUB_REPO" --body "$(az account show --query tenantId --output tsv)"
gh variable set AZURE_SUBSCRIPTION_ID --repo "$GITHUB_REPO" --body "$(az account show --query id --output tsv)"
gh variable set AZURE_RESOURCE_GROUP  --repo "$GITHUB_REPO" --body "$RG"
gh variable set AZURE_VM_NAME         --repo "$GITHUB_REPO" --body "$VM"
gh variable set SITE_ADDRESS          --repo "$GITHUB_REPO" --body "$SITE_ADDRESS"
```

### 4. First deploy

Push to `main`, or run the **Deploy** workflow by hand (Actions → Deploy → Run workflow). The app is then at
`https://$SITE_ADDRESS`.

Set a budget alert in the Azure portal (Cost Management → Budgets), so anything outside the free account is
noticed early.

## Operations

```sh
# Status and logs (no SSH needed)
az vm run-command invoke -g "$RG" -n "$VM" --command-id RunShellScript \
  --scripts "cd /opt/smart-grid && docker compose ps && docker compose logs --tail 100 api"

# Database backup to your machine (SSH from your own address)
ssh azureuser@"$SITE_ADDRESS" \
  "cd /opt/smart-grid && sudo docker compose exec -T db pg_dump -U postgres asset_registry" > backup.sql

# Roll back: run the Deploy workflow by hand on an older commit, or re-run an older run's deploy job.

# Remove everything
az group delete --name "$RG"
az ad app delete --id "$APP_ID"
```

The database lives in the `db-data` volume on the VM's disk. There are no automatic backups; take one before
anything risky.
