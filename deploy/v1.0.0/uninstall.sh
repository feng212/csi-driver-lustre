#!/bin/bash

echo "Uninstalling lustre csi driver,version 2.15.5"
kubectl delete -f controller-node-daemonset.yaml
kubectl delete -f controller-deployment.yaml
kubectl delete -f csi-driver.yaml
kubectl delete -f rbac.yaml
echo 'Uninstalled lustre csi driver  successfully.'
