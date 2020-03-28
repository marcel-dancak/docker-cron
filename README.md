
## Features
- YAML configuration format similar to Docker Compose
- Running tasks in new containers (docker run) or in existing containers (docker exec)
- Reload of configuration file on save
- File logging of every task run
- API server
- Optional web app server with real time info through websocket

## Example

### docker-compose.yml
```yaml
version: "2.3"

services:

  postgres:
    image: postgres:10-alpine
    volumes:
      - db-data:/var/lib/postgresql/data
      - db-backup:/backup/
    ports:
      - 5432:5432

  nginx:
    image: nginx
    volumes:
      - letsencrypt:/etc/letsencrypt/
    ports:
      - 80:80
      - 443:443

  dcron:
    image: dancakm/docker-cron
    depends_on:
      - nginx
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - ./dcron/tasks.yml:/etc/dcron/tasks.yml
      - letsencrypt:/etc/letsencrypt/
    environment:
      - DCRON_CONFIG_FILE=/etc/dcron/tasks.yml
      - DCRON_COMPOSE_PROJECT=example
      - DCRON_SSL_CERT=/etc/letsencrypt/live/example.com/fullchain.pem
      - DCRON_SSL_CERT_KEY=/etc/letsencrypt/live/example.com/privkey.pem
      - DCRON_WEB_PORT=7443
      - DCRON_API_PORT=7000
    expose:
      - 7000
    ports:
      - 7443:7443

volumes:
  certbot:
  letsencrypt:
  db-data:
  db-backup:
    driver: local
    driver_opts:
      type: none
      device: /mnt/volume/db-backups
      o: bind
```

### dcron/tasks.yml
```yaml
run:
  certbot:
    schedule: "30 23 * * *"
    image: certbot/certbot
    volumes:
      - certbot:/var/www/certbot/
      - letsencrypt:/etc/letsencrypt/
    command: ["renew", "--post-hook", "sh -c 'wget -qO- --post-data= http://dcron:7000/api/services/kill/nginx?signal=SIGHUP'"]

  db-backup:
    schedule: "45 23 * * *"
    service: postgres
    user: postgres
    command: ["sh", "-c", "pg_dump -Fc dbname -f /backup/db_`date +%d-%m-%y`.dump && ls -l /backup/"]

```
