#!/bin/bash

openssl genrsa -out ca.key.pem 2048

openssl req -x509 -new -nodes -key ca.key.pem -subj "/CN=CARoot" -days 3650 -out ca.cert.pem

openssl genrsa -out server.key.pem 2048

openssl req -new -key server.key.pem -subj "/CN=yourserver.domain.com" -out server.csr.pem

openssl x509 -req -in server.csr.pem -CA ca.cert.pem -CAkey ca.key.pem -CAcreateserial -out server.cert.pem -days 3650 -sha256

openssl genrsa -out client.key.pem 2048

openssl req -new -key client.key.pem -subj "/CN=Client" -out client.csr.pem

openssl x509 -req -in client.csr.pem -CA ca.cert.pem -CAkey ca.key.pem -CAcreateserial -out client.cert.pem -days 3650 -sha256

chmod 644 *.pem
