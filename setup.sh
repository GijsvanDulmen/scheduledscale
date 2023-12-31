#!/bin/bash

kind create cluster --name scheduledscale

kubectl cluster-info --context kind-scheduledscale