# extract from: https://gist.github.com/einthusan/f72c6dc7e0eff88b8bfe

echo "Remove existing go installation"
sudo apt-get purge golang -y
echo "Download and install Go, as well as create GOPATH directory"
cd ~
wget https://storage.googleapis.com/golang/go1.5.2.linux-amd64.tar.gz 
tar -xf go1.5.2.linux-amd64.tar.gz && rm go1.5.2.linux-amd64.tar.gz
sudo mv go /usr/local && sudo mkdir -p ~/gopath 
echo "set enviornment variables required for Go"
export GOPATH=~/gopath
cat <<EOF >> ~/.bashrc
export GOPATH=~/gopath
export PATH=$GOPATH/bin:/usr/local/go/bin:$PATH
export GORACE=log_path=/usr/local/go/racereport
export w=$GOPATH/src/github.com
EOF
