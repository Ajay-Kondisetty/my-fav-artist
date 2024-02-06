# my-fav-artist
A small microservice which gives the top track of the given country.

## Description

A small microservice which takes region as input and return the top track of the country, lyrics of the track, details of the artist, and recommendations based on track and artist.

## Getting Started

### Dependencies

* Docker
* Docker Compose
* Create a file named `local_env` in the `geomelody` folder and the following variables with appropriate values.
* The `country` input param should follow ISO 3166-1-Alpha-2 code format.
```
ENVIRONMENT=local

# LAST API credentials
LAST_API_KEY=<YOUR_LAST_API_KEY>
LAST_API_SECRET=<YOUR_LAST_API_SECRET>
LAST_API_URL=https://ws.audioscrobbler.com/2.0/

# MUSIC MIX API credentials
MUSIC_MIX_API_KEY=<YOUR_MUSIC_MIX_API_KEY<
MUSIC_MIX_URL=https://api.musixmatch.com/ws/1.1/

# HTTP Request config
HTTP_RESPONSE_HEADER_TIMEOUT=60s

COUNTRIES_JSON_FILE_NAME=countries.json

# Redis database
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_DEFAULT_EXPIRY=3600
```

### Installing

* Clone the repo
```
git clone https://github.com/Ajay-Kondisetty/my-fav-artist
```

### Executing program

* Change to root of the application which has Dockerfile
* Execute docker compose command to spin-up the app(if you are running it on windows machine then make sure launch Docker Desktop app first)
```
docker-compose up
```
* The app should be up and ready to handle connections within few seconds

## Authors
Ajay Kondisetty 
[@ajaykondisetty](https://www.linkedin.com/in/i-am-ajay/)

## Version History
* 1.0
    * Initial Release

## License

This project is licensed under the MIT License by Ajay Kondisetty - see the LICENSE.md file for details