#!/usr/local/bin/python3

import websocket # Documentations https://github.com/websocket-client/websocket-client
import json
import base64
import ssl
import time
import urllib.request
from datetime import datetime

# Địa chỉ IP của server
HOST = "tungle.local"

# Function này được gọi khi nhận được 1 tin nhắn từ server
def on_message(ws, message):
    # Biến đổi tin nhắn dạng string với format JSON
    # thành một variable có dạng dictionary
    data = json.loads(message)
    print('Now:', time.time_ns())
    print('Sent:', data['time'])
    print('Elapsed:', "%fms" % ((time.time_ns() - data['time']) * 1e-6))

# Function này được gọi khi có lỗi xảy ra khi kết nối với server
def on_error(ws, error):
    print(error)

# Function này được gọi khi kết nối với server bị ngắt
def on_close(ws):
    print("### closed ###")

if __name__ == "__main__":
    # SSL certificates locations
    cacert = "./keys/certs/pub/cacert.pem"
    clientcert = "./keys/certs/pub/clientcert.pem"
    clientkey = "./keys/certs/priv/clientkey.pem"

    # Login url
    url = "https://tungle.local:4433/sub"

    # SSL context
    sslContext = ssl.create_default_context(
        purpose=ssl.Purpose.SERVER_AUTH,
        cafile=cacert,
    )
    sslContext.load_cert_chain(clientcert, clientkey)

    # User login credentials
    creds = {
        "username": "user2",
        "password": "password2"
    }
    credsJSON = json.dumps(creds)
    credsJSONBytes = credsJSON.encode('utf-8')   # needs to be bytes

    # Make request
    req = urllib.request.Request(url)
    req.add_header('Content-Type', 'application/json; charset=utf-8')
    req.add_header('Content-Length', len(credsJSONBytes))

    # Send request
    resp = urllib.request.urlopen(req, credsJSONBytes, context=sslContext)
    respBody = resp.read()
    respBodyJSON = json.loads(respBody)

    header = {
        "Authorization": respBodyJSON['token']
    }
    # Tạo kết nối websocket với server
    websocket.enableTrace(True)
    sslopt = {
      'cert_reqs': ssl.PROTOCOL_SSLv23,
      'keyfile': clientkey,
      'certfile': clientcert,
      'ca_certs': cacert,
    }

    ws = websocket.WebSocketApp("wss://%s:4433/ws" % (HOST),
                              on_message = on_message,
                              on_error = on_error,
                              on_close = on_close,
                              header = header
                              )

    # Giữ kết nối vĩnh viễn với server
    ws.run_forever(sslopt=sslopt)
