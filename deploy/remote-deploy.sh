# Runs as root on the VM through Azure Run Command. The deploy workflow
# prepends these settings: IMAGE_PREFIX, IMAGE_TAG, SITE_ADDRESS, GHCR_USER,
# GHCR_TOKEN, COMPOSE_B64, CADDYFILE_B64. Prints DEPLOY_OK as its last line on
# success.
set -eu

APP_DIR=/opt/smart-grid
ENV_FILE="$APP_DIR/.env"

if ! command -v docker >/dev/null 2>&1; then
	echo "docker is not installed yet (cloud-init still running?)"
	exit 1
fi

mkdir -p "$APP_DIR"
cd "$APP_DIR"

echo "$COMPOSE_B64" | base64 -d >docker-compose.yml
echo "$CADDYFILE_B64" | base64 -d >Caddyfile

# The database password is generated on the VM the first time and never
# leaves it.
touch "$ENV_FILE"
chmod 600 "$ENV_FILE"
if ! grep -q '^DB_PASSWORD=' "$ENV_FILE"; then
	echo "DB_PASSWORD=$(openssl rand -hex 24)" >>"$ENV_FILE"
fi

# Rewrite the release settings, keeping every other line.
grep -vE '^(IMAGE_PREFIX|IMAGE_TAG|SITE_ADDRESS)=' "$ENV_FILE" >"$ENV_FILE.new" || true
{
	echo "IMAGE_PREFIX=$IMAGE_PREFIX"
	echo "IMAGE_TAG=$IMAGE_TAG"
	echo "SITE_ADDRESS=$SITE_ADDRESS"
} >>"$ENV_FILE.new"
chmod 600 "$ENV_FILE.new"
mv "$ENV_FILE.new" "$ENV_FILE"

# Pull with the job's short-lived token, then forget it.
export DOCKER_CONFIG="$APP_DIR/.docker"
echo "$GHCR_TOKEN" | docker login ghcr.io -u "$GHCR_USER" --password-stdin
trap 'docker logout ghcr.io >/dev/null 2>&1 || true' EXIT
docker compose pull
docker logout ghcr.io

docker compose up -d --remove-orphans --wait --wait-timeout 180
docker compose ps

# Keep the disk from filling with old releases.
docker image prune -af --filter "until=168h" >/dev/null

echo "DEPLOY_OK"
