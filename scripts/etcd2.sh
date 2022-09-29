#!/usr/bin/env sh

ETCD_VER=v2.0.9

# choose either URL
GITHUB_URL=https://github.com/etcd-io/etcd/releases/download
DOWNLOAD_URL=${GITHUB_URL}

rm -f ${PWD}/third_party/etcd-${ETCD_VER}-linux-amd64.tar.gz
rm -rf ${PWD}/third_party/etcd-download-test && mkdir -p ${PWD}/third_party/etcd-download-test

curl -L ${DOWNLOAD_URL}/${ETCD_VER}/etcd-${ETCD_VER}-linux-amd64.tar.gz -o ${PWD}/third_party/etcd-${ETCD_VER}-linux-amd64.tar.gz
tar xzvf ${PWD}/third_party/etcd-${ETCD_VER}-linux-amd64.tar.gz -C ${PWD}/third_party/etcd-download-test --strip-components=1
rm -f ${PWD}/third_party/etcd-${ETCD_VER}-linux-amd64.tar.gz

${PWD}/third_party/etcd-download-test/etcd --version
${PWD}/third_party/etcd-download-test/etcdctl version