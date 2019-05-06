#!/bin/zsh
#
CONFIG_PATH="configs"
KEYS_PATH="keys"

USERS_LIST="${CONFIG_PATH}/users.txt"

CLIENTS_PATH="keys/clients"
CACERT_PATH="${KEYS_PATH}/ca"
SVRCERT_PATH="${KEYS_PATH}/server"
BROWSERCERT_PATH="${KEYS_PATH}/browser"

JWT_KEYS_PATH="keys/jwt"

rm -rf ${KEYS_PATH}
mkdir -p ${CACERT_PATH} ${SVRCERT_PATH} ${JWT_KEYS_PATH} ${BROWSERCERT_PATH}

# generate jwt rsa keys pair
echo "\n\nGENERATING JWT KEYS"
openssl genrsa -out ${JWT_KEYS_PATH}/rsa.key 2048
openssl rsa -in ${JWT_KEYS_PATH}/rsa.key -pubout > ${JWT_KEYS_PATH}/rsa.pub

# generate cakey.pem
echo "\n\nGENERATING ROOT CERT"
openssl genrsa -out ${CACERT_PATH}/cakey.pem 2048
openssl req -new -nodes -key ${CACERT_PATH}/cakey.pem \
        -x509 -days 3650 -out ${CACERT_PATH}/cacert.pem \
        -subj /C="VN"/ST="Binh Duong"/L="Thu Dau Mot"/O="VGU"/OU="REC"/CN="Robotics and Electronics Club Root"

# generate serverkey.pem
echo "\n\nGENERATING SERVER CERT"
openssl genrsa -out ${SVRCERT_PATH}/serverkey.pem 2048
openssl req -new -nodes -key ${SVRCERT_PATH}/serverkey.pem \
        -out ${SVRCERT_PATH}/servercert.csr \
        -config ${CONFIG_PATH}/san.cnf -extensions req_ext -extensions x509_ext \
        -subj /C="VN"/ST="Binh Duong"/L="Thu Dau Mot"/O="VGU"/OU="REC"/CN="192.168.1.100"
openssl x509 -req -in ${SVRCERT_PATH}/servercert.csr \
        -CA ${CACERT_PATH}/cacert.pem -CAkey ${CACERT_PATH}/cakey.pem \
        -extfile ${CONFIG_PATH}/san.cnf -extensions req_ext -extensions x509_ext \
        -CAcreateserial -out ${SVRCERT_PATH}/servercert.pem

echo "\n\nGENERATING BROWSER CERT"
openssl genrsa -out ${BROWSERCERT_PATH}/browserkey.pem 2048
openssl req -new -nodes -key ${BROWSERCERT_PATH}/browserkey.pem \
        -out ${BROWSERCERT_PATH}/browsercert.csr \
        -subj /C="VN"/ST="Binh Duong"/L="Thu Dau Mot"/O="VGU"/OU="REC"/CN="Robotics and Electronics Club Browser"
openssl x509 -req -in ${BROWSERCERT_PATH}/browsercert.csr \
        -CA ${CACERT_PATH}/cacert.pem -CAkey ${CACERT_PATH}/cakey.pem \
        -CAcreateserial -out ${BROWSERCERT_PATH}/browsercert.pem
openssl pkcs12 -export -clcerts -in ${BROWSERCERT_PATH}/browsercert.pem -inkey ${BROWSERCERT_PATH}/browserkey.pem -out ${BROWSERCERT_PATH}/browsercert.p12

# generate clientkey.pem
while IFS="" read -r p || [ -n "$p" ]
do
    CLIENTCERT_PATH="${CLIENTS_PATH}/${p}"
    mkdir -p ${CLIENTCERT_PATH}

    echo "\n\nGENERATING CERT FOR '${p}'"
    openssl genrsa -out ${CLIENTCERT_PATH}/clientkey.pem 2048
    openssl req -new -nodes -key ${CLIENTCERT_PATH}/clientkey.pem \
            -out ${CLIENTCERT_PATH}/clientcert.csr \
            -subj /C="VN"/ST="Binh Duong"/L="Thu Dau Mot"/O="Contestants"/OU="Client"/CN="${p}"
    openssl x509 -req -in ${CLIENTCERT_PATH}/clientcert.csr \
            -CA ${CACERT_PATH}/cacert.pem -CAkey ${CACERT_PATH}/cakey.pem \
            -CAcreateserial -out ${CLIENTCERT_PATH}/clientcert.pem
done < ${USERS_LIST}

