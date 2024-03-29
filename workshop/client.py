#!/usr/bin/env python3

## Khai báo các thư viện được sử dụng
import websocket
import json
import ssl
import time
import urllib.request
from pathlib import Path

# http://vgurobocon2019.local/

# Các hàm được sử dụng


## Cài đặt và trả về thông tin cho giao thức mã hoá HTTPS
def makeSSLContext(ca, crt, key):
    sslCTX = ssl.create_default_context(
        purpose=ssl.Purpose.SERVER_AUTH,
        cafile=ca
    )

    sslCTX.load_cert_chain(crt, key)

    return sslCTX


## Trả về chuỗi JSON chứa thông tin đăng nhâp
def makeJSONCredentials(username, password):
    creds = {
        "username": username,
        "password": password
    }

    return json.dumps(creds).encode("utf-8")


## Cài đặt và trả về thông tin của yêu cầu HTTPS
def makeRequestHeader(url, contentType, content):
    req = urllib.request.Request(url)

    req.add_header('Content-Type', contentType)
    req.add_header('Content-Length', len(content))

    return req


## Gửi yêu cầu đăng nhập và trả về mã xác thực
def getToken(url, username, password,
             ca, crt, key):
    reqSSLContext = makeSSLContext(ca, crt, key)
    reqContent = makeJSONCredentials(username, password)
    req = makeRequestHeader(
        url,
        'application/json; charset=utf-8',
        reqContent
    )

    # Gửi yêu cầu và nhận kết quả trả về
    resp = urllib.request.urlopen(
        req, data=reqContent, context=reqSSLContext)

    # Đọc và trả về mã xác thực
    respBody = resp.read()
    respBodyJSON = json.loads(respBody.decode('utf-8'))

    return respBodyJSON["token"]


# Đăng nhập và nhận dữ liệu


## Cài đặt thông tin của giao thức mã hoá cho Websocket
CA_CRT = str(Path("../keys/certs/pub/cacert.pem"))
CRT = str(Path("../keys/certs/pub/clientcert.pem"))
KEY = str(Path("../keys/certs/priv/clientkey.pem"))

sslopt = {
    'cert_reqs': ssl.PROTOCOL_SSLv23,
    'keyfile': KEY,
    'certfile': CRT,
    'ca_certs': CA_CRT,
}


## Nhận mã xác thực và thêm mã xác thực vào thông tin yêu cầu Websocket
HOST = "127.0.0.1"
PORT = 4433

url = 'https://%s:%s/subscribe' % (HOST, PORT)
token = getToken(url,
                 'user', 'password',
                 CA_CRT, CRT, KEY)

header = {
    'Authorization': 'Bearer %s' % (token)
}


## Thiết lập kết nối Websocket và bắt đầu nhận dữ liệu
url = 'wss://%s:%s/data' % (HOST, PORT)
ws = websocket.create_connection(url,
                                 header=header,
                                 sslopt=sslopt)

while True:
    msg = ws.recv()
    packet = json.loads(msg.decode('utf-8'))
    print(packet)
    ws.send(json.dumps({'finished': True}).encode('utf-8'))
