#!/usr/bin/bash

KEYS_PATH="./keys"

CERTS_PATH="./keys/certs"
CERTS_PRIV_PATH="${CERTS_PATH}/priv"
CERTS_PUB_PATH="${CERTS_PATH}/pub"

JWT_KEYS_PATH="./keys/jwt"

rm -rf ${KEYS_PATH}
mkdir -p ${CERTS_PRIV_PATH} ${CERTS_PUB_PATH} ${JWT_KEYS_PATH}

# generate jwt rsa keys pair
echo "\n\nGENERATING JWT KEYS"
openssl genrsa -out ${JWT_KEYS_PATH}/rsa.key 4096
openssl rsa -in ${JWT_KEYS_PATH}/rsa.key -pubout > ${JWT_KEYS_PATH}/rsa.pub

# generate cakey.pem
echo "\n\nGENERATING ROOT KEYS"
openssl genrsa -out ${CERTS_PRIV_PATH}/cakey.pem 4096
openssl req -new -key ${CERTS_PRIV_PATH}/cakey.pem -x509 -days 3650 -out ${CERTS_PUB_PATH}/cacert.pem -passout pass:@Brisingr5013 -subj /C=VN/ST="Binh Duong"/L="Thu Dau Mot"/O="VGU"/OU="REC"/CN="Robotics and Electronics Club"

# generate serverkey.pem
echo "\n\nGENERATING SERVER KEYS"
openssl genrsa -out ${CERTS_PRIV_PATH}/serverkey.pem 4096
openssl req -new -nodes -key ${CERTS_PRIV_PATH}/serverkey.pem -out ${CERTS_PRIV_PATH}/servercert.csr -subj /C=VN/ST="Binh Duong"/L="Thu Dau Mot"/O="VGU"/OU="REC"/CN="vgurobocon2019.local"
openssl x509 -req -in ${CERTS_PRIV_PATH}/servercert.csr -CA ${CERTS_PUB_PATH}/cacert.pem -CAkey ${CERTS_PRIV_PATH}/cakey.pem -CAcreateserial -out ${CERTS_PUB_PATH}/servercert.pem

# generate clientkey.pem
echo "\n\nGENERATING CLIENT KEYS"
openssl genrsa -out ${CERTS_PRIV_PATH}/clientkey.pem 4096
openssl req -new -nodes -key ${CERTS_PRIV_PATH}/clientkey.pem -out ${CERTS_PRIV_PATH}/clientcert.csr -subj /C=VN/ST="Binh Duong"/L="Thu Dau Mot"/O="Client"/OU="Contestant"/CN="client.local"
openssl x509 -req -in ${CERTS_PRIV_PATH}/clientcert.csr -CA ${CERTS_PUB_PATH}/cacert.pem -CAkey ${CERTS_PRIV_PATH}/cakey.pem -CAcreateserial -out ${CERTS_PUB_PATH}/clientcert.pem
