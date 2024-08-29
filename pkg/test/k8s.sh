#!/bin/bash
systemctl enable corosync pacemaker pcsd
pcs host auth -u hacluster -p hacluster 172.16.100.141 172.16.100.142 

pcs cluster setup  cluster 172.16.100.141 172.16.100.142  --force

pcs cluster start --all
pcs cluster enable --all
systemctl start corosync pacemaker pcsd

pcs property set no-quorum-policy=ignore
pcs property set stonith-enabled=false

pcs resource create mgs ocf:wistor:WISTOR target=/dev/xi_raid6A1p1 mountpoint=/wi-stor/mgs
pcs constraint location mgs prefers 172.16.100.142=100
pcs constraint location mgs prefers 172.16.100.141=50

pcs resource create mdt0000 ocf:wistor:WISTOR target=/dev/xi_raid6A1p2 mountpoint=/wi-stor/mdt0000
pcs constraint location mdt0000 prefers 172.16.100.141=50
pcs constraint location mdt0000 prefers 172.16.100.142=100

pcs resource create mdt0001 ocf:wistor:WISTOR target=/dev/xi_raid6A2p2 mountpoint=/wi-stor/mdt0001
pcs constraint location mdt0001 prefers 172.16.100.141=50
pcs constraint location mdt0001 prefers 172.16.100.142=100

pcs resource create mdt0002 ocf:wistor:WISTOR target=/dev/xi_raid6B1p2 mountpoint=/wi-stor/mdt0002
pcs constraint location mdt0002 prefers 172.16.100.141=100
pcs constraint location mdt0002 prefers 172.16.100.142=50

pcs resource create mdt0003 ocf:wistor:WISTOR target=/dev/xi_raid6B2p2 mountpoint=/wi-stor/mdt0003
pcs constraint location mdt0003 prefers 172.16.100.141=100
pcs constraint location mdt0003 prefers 172.16.100.142=50

pcs resource create ost0000 ocf:wistor:WISTOR target=/dev/xi_raid6A1p3 mountpoint=/wi-stor/ost0000
pcs constraint location ost0000 prefers 172.16.100.141=50
pcs constraint location ost0000 prefers 172.16.100.142=100

pcs resource create ost0001 ocf:wistor:WISTOR target=/dev/xi_raid6A2p3 mountpoint=/wi-stor/ost0001
pcs constraint location ost0001 prefers 172.16.100.141=50
pcs constraint location ost0001 prefers 172.16.100.142=100

pcs resource create ost0002 ocf:wistor:WISTOR target=/dev/xi_raid6B1p3 mountpoint=/wi-stor/ost0002
pcs constraint location ost0001 prefers 172.16.100.141=100
pcs constraint location ost0002 prefers 172.16.100.142=50

pcs resource create ost0003 ocf:wistor:WISTOR target=/dev/xi_raid6B2p3 mountpoint=/wi-stor/ost0003
pcs constraint location ost0003 prefers 172.16.100.141=100
pcs constraint location ost0003 prefers 172.16.100.142=50



pcs resource create healthLNET ocf:wistor:healthLNET lctl=true multiplier=1001 device=ib0 host_list="10.10.8.41@o2ib 10.10.8.42@o2ib " clone
pcs constraint location mgs rule score=-INFINITY pingd lte 0 or not_defined pingd
pcs constraint location mdt0000 rule score=-INFINITY pingd lte 0 or not_defined pingd
pcs constraint location mdt0001 rule score=-INFINITY pingd lte 0 or not_defined pingd
pcs constraint location mdt0002 rule score=-INFINITY pingd lte 0 or not_defined pingd
pcs constraint location mdt0003 rule score=-INFINITY pingd lte 0 or not_defined pingd
pcs constraint location ost0000 rule score=-INFINITY pingd lte 0 or not_defined pingd
pcs constraint location ost0001 rule score=-INFINITY pingd lte 0 or not_defined pingd
pcs constraint location ost0002 rule score=-INFINITY pingd lte 0 or not_defined pingd
pcs constraint location ost0003 rule score=-INFINITY pingd lte 0 or not_defined pingd

