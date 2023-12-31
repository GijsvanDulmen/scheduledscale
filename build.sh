#!/bin/bash
docker build --progress=plain -t scheduledscale .

docker tag scheduledscale ghcr.io/gijsvandulmen/scheduledscale:latest
docker tag scheduledscale ghcr.io/gijsvandulmen/scheduledscale:1.0

docker push ghcr.io/gijsvandulmen/scheduledscale:latest
docker push ghcr.io/gijsvandulmen/scheduledscale:1.0