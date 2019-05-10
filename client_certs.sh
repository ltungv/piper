#!/bin/zsh
#
KEYS_PATH="keys"
CLIENT_PATH="keys/clients"
CACERT_PATH="${KEYS_PATH}/ca"

echo "\n\nGENERATING CLIENT CERT"
openssl genrsa -out ${CLIENT_PATH}/clientkey.pem 2048

openssl req -new -nodes -key ${CLIENT_PATH}/clientkey.pem \
        -out ${CLIENT_PATH}/clientcert.csr \
        -subj /C="VN"/ST="Binh Duong"/L="Thu Dau Mot"/O="Contestants"/OU="Client"/CN="VGU Robocon 2019 Contestant"

openssl x509 -req -in ${CLIENT_PATH}/clientcert.csr \
        -CA ${CACERT_PATH}/cacert.pem -CAkey ${CACERT_PATH}/cakey.pem \
        -CAcreateserial -out ${CLIENT_PATH}/clientcert.pem
