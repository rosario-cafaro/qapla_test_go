## GO

```bash
docker-compose build
```

```bash
docker-compose up
```

Valid parameters JSON response\
[http://127.0.0.1:8080/?tracking=ITXXXXXXXXXX&json=1](http://127.0.0.1:8080/?tracking=ITXXXXXXXXXX&json=1)

Valid parameters HTML response\
[http://127.0.0.1:8080/?tracking=ITXXXXXXXXXX](http://127.0.0.1:8080/?tracking=ITXXXXXXXXXX)

Missing parameter JSON response\
[http://127.0.0.1:8080/?tracking=&json=1](http://127.0.0.1:8080/?tracking=&json=1)

Missing parameter HTML response\
[http://127.0.0.1:8080/?tracking=&json=](http://127.0.0.1:8080/?tracking=&json=)