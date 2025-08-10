#!/bin/sh

# Use rsync to copy files while excluding .git folder and other unnecessary files
rsync -avz --progress \
  --exclude=".git/" \
  --exclude=".gitignore" \
  --exclude=".DS_Store" \
  --exclude="Caddyfile" \
  --exclude="publish.sh" \
  ./ root@10.6.0.1:/www/wwwroot/hbjs/

echo "Deployment complete!"