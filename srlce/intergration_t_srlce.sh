#!/bin/bash
# This script used to execute simple integration test for SRLCE

# Set the environment variables to replace 
LAB_CA_DIR=~/clab/2node/clab-2nd
USER=admin
PASSWORD=admin
TARGET=clab-2nd-srl2

echo ">>>>>> SSH scraping only"
go run srlce.go -target $TARGET -username $USER -password $PASSWORD -d -logSSH
if [ $? -ne 0 ]; then
    echo "SSH scraping only failed"
fi
echo ">>>>>> SSH scraping with gNOI skip verify"
go run srlce.go -target $TARGET -username $USER -password $PASSWORD -gNOIdld -SkipVerify -d
if [ $? -ne 0 ]; then
    echo "SSH scraping with gNOI skip verify failed"
fi
echo ">>>>>> SSH scraping with gNOI and root CA certificate, key and certificate"
go run srlce.go -target $TARGET -username $USER -password $PASSWORD -gNOIdld -rootCA ${LAB_CA_DIR}/ca/root/root-ca.pem -key ${LAB_CA_DIR}/ca/srl2/srl2-key.pem  -cert ${LAB_CA_DIR}/ca/srl2/srl2.pem -d
if [ $? -ne 0 ]; then
    echo "SSH scraping with gNOI and root CA certificate, key and certificate failed"
fi
echo ">>>>>> SSH scraping with gNOI and root CA certificate, key and certificate + clab cleanup"
go run srlce.go -target $TARGET -username $USER -password $PASSWORD -gNOIdld -rootCA ${LAB_CA_DIR}/ca/root/root-ca.pem -key ${LAB_CA_DIR}/ca/srl2/srl2-key.pem  -cert ${LAB_CA_DIR}/ca/srl2/srl2.pem -cclab -d
if [ $? -ne 0 ]; then
    echo "SSH scraping with gNOI and root CA certificate, key and certificate + clab cleanup failed"
fi
echo ">>>>>> JSON RPC scraping only"
go run srlce.go -target $TARGET -jsonrpc -username $USER -password $PASSWORD -d
if [ $? -ne 0 ]; then
    echo "JSON RPC scraping only failed"
fi
echo ">>>>>> JSON RPC scraping only + clab cleanup"
go run srlce.go -target $TARGET -jsonrpc -username $USER -password $PASSWORD -cclab -d
if [ $? -ne 0 ]; then
    echo "JSON RPC scraping only + clab cleanup failed"
fi
echo "JSON RPC with gNOI skip verify"
go run srlce.go -target $TARGET -jsonrpc -username $USER -password $PASSWORD -gNOIdld -SkipVerify -d
if [ $? -ne 0 ]; then
    echo "JSON RPC with gNOI skip verify failed"
fi
echo ">>>>>> JSON RPC with gNOI skip verify"
go run srlce.go -target $TARGET -jsonrpc -username $USER -password $PASSWORD -gNOIdld -SkipVerify -cclab -d
if [ $? -ne 0 ]; then
    echo "JSON RPC with gNOI skip verify + clab cleanup failed"
fi
echo ">>>>>> JSON RPC with gNOI and root CA certificate, key and certificate" 
go run srlce.go -target $TARGET -jsonrpc -username $USER -password $PASSWORD -gNOIdld -rootCA ${LAB_CA_DIR}/ca/root/root-ca.pem -key ${LAB_CA_DIR}/ca/srl2/srl2-key.pem  -cert ${LAB_CA_DIR}/ca/srl2/srl2.pem -d
if [ $? -ne 0 ]; then
    echo "JSON RPC with gNOI and root CA certificate, key and certificate failed"
fi
echo ">>>>>> JSON RPC with gNOI and root CA certificate, key and certificate + clab cleanup"
go run srlce.go -target $TARGET -jsonrpc -username $USER -password $PASSWORD -gNOIdld -rootCA ${LAB_CA_DIR}/ca/root/root-ca.pem -key ${LAB_CA_DIR}/ca/srl2/srl2-key.pem  -cert ${LAB_CA_DIR}/ca/srl2/srl2.pem -cclab -d
if [ $? -ne 0 ]; then
    echo "JSON RPC with gNOI and root CA certificate, key and certificate + clab cleanup failed"
fi
echo ">>>>>> JSON RPC with gNOI and root CA certificate, key and certificate + clab cleanup + debug"
go run srlce.go -target $TARGET -jsonrpc -username $USER -password $PASSWORD -gNOIdld -rootCA ${LAB_CA_DIR}/ca/root/root-ca.pem -key ${LAB_CA_DIR}/ca/srl2/srl2-key.pem  -cert ${LAB_CA_DIR}/ca/srl2/srl2.pem -cclab -d
if [ $? -ne 0 ]; then
    echo "JSON RPC with gNOI and root CA certificate, key and certificate + clab cleanup + debug failed"
fi
echo ">>>>>> JSON RPC with gNOI and root CA certificate, key and certificate + clab cleanup + debug + log file"
go run srlce.go -target $TARGET -jsonrpc -username $USER -password $PASSWORD -gNOIdld -rootCA ${LAB_CA_DIR}/ca/root/root-ca.pem -key ${LAB_CA_DIR}/ca/srl2/srl2-key.pem  -cert ${LAB_CA_DIR}/ca/srl2/srl2.pem -cclab -d -logFile ./srlce_jsonrpc.log
if [ $? -ne 0 ]; then
    echo "JSON RPC with gNOI and root CA certificate, key and certificate + clab cleanup + debug + log file failed"
fi
echo ">>>>>> SSH scraping with gNOI and root CA certificate, key and certificate + clab cleanup + debug + log file"
go run srlce.go -target $TARGET -username $USER -password $PASSWORD -gNOIdld -rootCA ${LAB_CA_DIR}/ca/root/root-ca.pem -key ${LAB_CA_DIR}/ca/srl2/srl2-key.pem  -cert ${LAB_CA_DIR}/ca/srl2/srl2.pem -cclab -d -logFile ./srlce_ssh.log -logSSH
if [ $? -ne 0 ]; then
    echo "SSH scraping with gNOI and root CA certificate, key and certificate + clab cleanup + debug + log file failed"
fi





