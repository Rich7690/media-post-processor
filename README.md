# Media Post Processor

## Summary

This project is a web service and transcode job scheduler that runs as a result of download webhook requests from [Sonarr](https://sonarr.tv/) and [Radarr](https://radarr.video/). It is currently alpha level software that was built as my own personal desire to have all of my video files transcoded to a standard x264 codec and mp4 file format. The idea came mainly from my usage of the mp4 automator [here](https://github.com/mdhiggins/sickbeard_mp4_automator). Currently the service is hard coded with these transcode options, but will soon add a configuration UI to support many more things. Feel free to add feature requests to the issue tracker in this repo


## Usage
### Docker
This service can be used in docker-compose as follows:

```
  redis:
   image: redis:5-alpine
   container_name: redis
   command: --appendonly yes
   ports:
    - 6379:6379 Optional:if you want redis access on your host
   networks:
    - app-net
   volumes:
    - /tmp/redis:/data # point this somewhere persistent if you want jobs to live accross container restarts
  web:
    image: unknowndev7690/web:latest
    networks:
     - app-net
    ports:
     - 8080:8080 # Optional: If you want to explore the api manually
    container_name: web
    volumes:
     - /media:/media # You'll want to mount paths the same as radarr and sonarr see them here until path mapping is supported
    depends_on:
     - redis
    environment:
     - REDIS_ADDRESS=redis:6379
     - RADARR_API_KEY=API_KEY # Copy your Radarr API key here
     - SONARR_API_KEY=PI_KEY # Copy your Sonarr API key here
     - FFMPEG_PATH=/usr/bin/ffmpeg # Optional: Defaults to alpine linux install location. 
     - FFPROBE_PATH=/usr/bin/ffprobe # Optional: Defaults to alpine linux install location. 
     - SONARR_BASE_ENDPOINT=http://some-path-to-sonarr.com # Optional: Only enable if you want Sonarr integration
     - RADARR_BASE_ENDPOINT=https://some-path-radarr.com # Optional: Only enable if you want Radarr integration
     - ENABLE_RADARR_SCANNER=true # Use to enable individual components of the app. 
     - ENABLE_WEB=true # This enables the webhook web service. 
     - ENABLE_WORKER=true # This enables background transcoder. This allows you to deploy them in separate containers
```

Radarr or Sonarr should then be configured on the Connect section as follows:
![](https://i.imgur.com/b5AqAlJ.png)

You can do a `docker logs -f web` to validate that is receiving requests correctly.

### Non-Docker
Currently, I don't cross-compile builds for native setups, but if you prefer to run apps on your OS directly, you should be able to just compile with `go build ./...` once you have installed golang 1.13 or above on that OS. You then can setup the binary yourself.
