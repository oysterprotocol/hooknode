#!/bin/bash

#install php and dependencies
sudo apt-get update
sudo apt-get -y install php
sudo apt-get -y install nginx
sudo apt-get -y install php-fpm
sudo apt-get -y install php-curl

#install node.js (for Nelson)
curl -sL https://deb.nodesource.com/setup_8.x | sudo -E bash -
sudo apt-get -y install -y nodejs

#install IRI
#pre-check Java agreements
echo debconf shared/accepted-oracle-license-v1-1 select true | sudo debconf-set-selections
echo debconf shared/accepted-oracle-license-v1-1 seen true | sudo debconf-set-selections
sudo apt-get -y install software-properties-common -y && sudo add-apt-repository ppa:webupd8team/java -y && sudo apt update && sudo apt install oracle-java8-installer curl wget jq git -y && sudo apt install oracle-java8-set-default -y
sudo sh -c 'echo JAVA_HOME="/usr/lib/jvm/java-8-oracle" >> /etc/environment' && source /etc/environment
sudo useradd -s /usr/sbin/nologin -m iota
sudo -u iota mkdir -p /home/iota/node /home/iota/node/ixi /home/iota/node/mainnetdb
sudo -u iota wget -O /home/iota/node/iri-1.4.2.0.jar https://github.com/iotaledger/iri/releases/download/v1.4.2.0/iri-1.4.2.0.jar

#find RAM, in MB
phymem=$(awk -F":" '$1~/MemTotal/{print $2}' /proc/meminfo )
phymem=${phymem:0:-2}
#allot about 75% of RAM to java
phymem=$((($phymem/1333) + ($phymem % 1333 > 0)))
xmx="Xmx"
xmx_end="m"
xmx=$xmx$phymem$xmx_end

#set up Systemd service
cat <<EOF | sudo tee /lib/systemd/system/iota.service
[Unit]
Description=IOTA (IRI) full node
After=network.target

[Service]
WorkingDirectory=/home/iota/node
User=iota
PrivateDevices=yes
ProtectSystem=full
Type=simple
ExecReload=/bin/kill -HUP $MAINPID
KillMode=mixed
KillSignal=SIGTERM
TimeoutStopSec=60
ExecStart=/usr/bin/java -$xmx -Djava.net.preferIPv4Stack=true -jar iri-1.4.1.7.jar -c iota.ini
SyslogIdentifier=IRI
Restart=on-failure
RestartSec=30

[Install]
WantedBy=multi-user.target
Alias=iota.service

EOF
#configure IRI
cat << "EOF" | sudo -u iota tee /home/iota/node/iota.ini
[IRI]
PORT = 14265
UDP_RECEIVER_PORT = 14600
TCP_RECEIVER_PORT = 15600
API_HOST = 0.0.0.0
IXI_DIR = ixi
HEADLESS = true
DEBUG = false
TESTNET = false
DB_PATH = mainnetdb
RESCAN_DB = false

REMOTE_LIMIT_API = "interruptAttachingToTangle, attachToTangle, setApiRateLimit"
#We don't need to add normal neighbors as we're going to be using Nelson
EOF
#Download the last known Tangle database
cd /tmp/ && curl -LO http://db.iota.partners/IOTA.partners-mainnetdb.tar.gz && sudo -u iota tar xzfv /tmp/IOTA.partners-mainnetdb.tar.gz -C /home/iota/node/mainnetdb && rm /tmp/IOTA.partners-mainnetdb.tar.gz
#install Nelson
sudo npm install -g nelson.cli
#start the IOTA service
sudo service iota start
systemctl enable iota.service
#configure auto updates for IRI 
echo '*/15 * * * * root bash -c "bash <(curl -s https://gist.githubusercontent.com/zoran/48482038deda9ce5898c00f78d42f801/raw)"' | sudo tee /etc/cron.d/iri_updater > /dev/null

#Start Nelson with pm2
sudo npm install pm2 -g
sudo pm2 startup
sudo pm2 start nelson -- --getNeighbors
sudo pm2 save

#Install hooknode service
sudo mkdir -p /home/oyster/hooknode
sudo git clone https://github.com/oysterprotocol/hooknode.git /home/oyster/hooknode
sudo rm -rf /var/www/html
sudo ln -s /home/oyster/hooknode/html /var/www/html
sudo cp /home/oyster/hooknode/nginx.conf /etc/nginx/
sudo service nginx restart

#get public ip
ips="$(dig +short myip.opendns.com @resolver1.opendns.com)"


#prepare and show confirmation message
endmsg1="Installation finished, your hooknode is set up at http://"
endmsg2=":250/HookNode.php"
echo $endmsg1$ips$endmsg2

# - we need to move the node to https
