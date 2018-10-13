FROM ubuntu

RUN apt-get update -qq && apt-get upgrade -y && apt-get install -y sudo unattended-upgrades curl
RUN echo -e "APT::Periodic::Update-Package-Lists \"1\";\nAPT::Periodic::Unattended-Upgrade \"1\";\n" > /etc/apt/apt.conf.d/20auto-upgrades

RUN /bin/bash -c 'bash <(curl https://get.parity.io -L)'

RUN /bin/bash -c 'curl https://dist.ipfs.io/go-ipfs/v0.4.17/go-ipfs_v0.4.17_linux-amd64.tar.gz -L > go-ipfs_v0.4.17_linux-amd64.tar.gz'
RUN /bin/bash -c 'tar xvvf go-ipfs_v0.4.17_linux-amd64.tar.gz && pushd go-ipfs && ./install.sh && popd'

RUN mkdir -p /opt/provide.network
RUN touch /opt/spec.json
RUN touch /opt/bootnodes.txt

ADD start-node.sh /opt/start-node.sh

EXPOSE 5001
EXPOSE 8050
EXPOSE 8051
EXPOSE 8080
EXPOSE 30300

WORKDIR /opt

ENTRYPOINT ["./start-node.sh"]
