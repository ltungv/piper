#!/bin/zsh

KEYS_PATH="keys"

CERTS_PATH="keys/certs"
CERTS_PRIV_PATH="${CERTS_PATH}/priv"
CERTS_PUB_PATH="${CERTS_PATH}/pub"
JWT_KEYS_PATH="keys/jwt"

CONFIG_PATH="configs"

rm -rf ${KEYS_PATH}
mkdir -p ${CERTS_PRIV_PATH} ${CERTS_PUB_PATH} ${JWT_KEYS_PATH}


# generate jwt rsa keys pair
echo "\n\nGENERATING JWT KEYS"
openssl genrsa -out ${JWT_KEYS_PATH}/rsa.key 2048

openssl rsa -in ${JWT_KEYS_PATH}/rsa.key -pubout > ${JWT_KEYS_PATH}/rsa.pub


# generate cakey.pem
echo "\n\nGENERATING ROOT KEYS"
openssl genrsa -out ${CERTS_PRIV_PATH}/cakey.pem 2048

openssl req -new -key ${CERTS_PRIV_PATH}/cakey.pem \
        -x509 -days 3650 -out ${CERTS_PUB_PATH}/cacert.pem \
        -passout pass:@Brisingr5013 \
        -subj /C="VN"/ST="Binh Duong"/L="Thu Dau Mot"/O="VGU"/OU="REC"/CN="Robotics and Electronics Club"


# generate serverkey.pem
echo "\n\nGENERATING SERVER KEYS"
openssl genrsa -out ${CERTS_PRIV_PATH}/serverkey.pem 2048
j
openssl req -new -nodes -key ${CERTS_PRIV_PATH}/serverkey.pem \
        -out ${CERTS_PRIV_PATH}/servercert.csr \
        -config ${CONFIG_PATH}/san.cnf -extensions req_ext -extensions x509_ext \
        -subj /C="VN"/ST="Binh Duong"/L="Thu Dau Mot"/O="VGU"/OU="REC"/CN="0.0.0.0"

openssl x509 -req -in ${CERTS_PRIV_PATH}/servercert.csr \
        -CA ${CERTS_PUB_PATH}/cacert.pem -CAkey ${CERTS_PRIV_PATH}/cakey.pem \
        -extfile ${CONFIG_PATH}/san.cnf -extensions req_ext -extensions x509_ext \
        -CAcreateserial -out ${CERTS_PUB_PATH}/servercert.pem


# generate clientkey.pem
echo "\n\nGENERATING CLIENT KEYS"
openssl genrsa -out ${CERTS_PRIV_PATH}/clientkey.pem 2048

openssl req -new -nodes -key ${CERTS_PRIV_PATH}/clientkey.pem \
        -out ${CERTS_PRIV_PATH}/clientcert.csr \
        -subj /C="VN"/ST="Ho Chi Minh"/L="Ho Chi Minh"/O="Contestants"/OU="Client"/CN="VGU Robocon 2019 Client"

openssl x509 -req -in ${CERTS_PRIV_PATH}/clientcert.csr \
        -CA ${CERTS_PUB_PATH}/cacert.pem -CAkey ${CERTS_PRIV_PATH}/cakey.pem \
        -CAcreateserial -out ${CERTS_PUB_PATH}/clientcert.pem
