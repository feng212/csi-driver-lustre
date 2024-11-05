#!/bin/bash

echo "Installing lustre csi driver,version 2.15.5"
kubectl apply -f rbac.yaml
kubectl apply -f csi-driver.yaml
kubectl apply -f controller-deployment.yaml
kubectl apply -f controller-node-daemonset.yaml
echo 'lustre csi driver installed successfully.'
