#!/usr/bin/env bash
helm -n scheduledscale install --set LOG_LEVEL=DEBUG --create-namespace scheduledscale helm/