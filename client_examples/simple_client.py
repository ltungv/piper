import websocket
import json
import ssl
import time
import urllib.request
from pathlib import Path
from threading import Thread


# Cài đặt và trả về thông tin của giao thức mã hoá
def makeSSLContext(ca, crt, key):
    sslCTX = ssl.create_default_context(
        purpose=ssl.Purpose.SERVER_AUTH,
        cafile=ca
    )

    sslCTX.load_cert_chain(crt, key)

    return sslCTX


# Trả về chuỗi json từ thông tin đăng nhâp
def makeJSONCredentials(username, password):
    creds = {
        "username": username,
        "password": password
    }

    return json.dumps(creds).encode("utf-8")


# Cài đặt và trả về thông tin của yêu cầu
def makeRequestHeader(url, contentType, content):
    req = urllib.request.Request(url)

    req.add_header('Content-Type', contentType)
    req.add_header('Content-Length', len(content))

    return req


# Gửi yêu cầu dăng nhập và trả về mã xác thực
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


def run(ws, user, i):
    # Nhận dữ liệu
    messages = 0
    totalMsg = 0
    count = 0
    delay = 0
    mark = time.time()

    while True:
        msg = ws.recv()

        packet = json.loads(msg.decode('utf-8'))
        totalMsg += 1
        count += 1
        delay += (time.time_ns() - packet['time']) * 1e-6

        if (count % 120 == 0):
            print("elapsed: %d | %d messages read | avg ping: %.3fms" %
                    (time.time() - mark, totalMsg, delay / count))
            delay = 0
            count = 0
            mark = time.time()

        ws.send(json.dumps({'finished': True}).encode('utf-8'))


def createUserInstances(username, nInstances):
    # Nhận mã xác thực
    # và thêm mã xác thực vào thông tin yêu cầu websocket
    token = getToken('https://%s:%s/subscribe' % (HOST, PORT),
                     user, users[user],
                     CA_CRT, CRT, KEY)
    header = {
        'Authorization': 'Bearer %s' % (token)
    }

    for i in range(INSTANCE_PER_USER):
        print("%s-%d" % (username, i))
        ws = websocket.create_connection('wss://%s:%s/data' % (HOST, PORT),
                                         header=header,
                                         sslopt=sslopt)
        t = Thread(target=run, args=(ws, user, i))
        t.start()


if __name__ == '__main__':
    CA_CRT = Path("../keys/ca/cacert.pem")
    CRT = Path("../keys/clients/TheQueen/clientcert.pem")
    KEY = Path("../keys/clients/TheQueen/clientkey.pem")
    print(CRT)
    print(KEY)
    HOST = "192.168.1.100"
    PORT = 4433

    INSTANCE_PER_USER = 1

    # Cài đặt và thông tin của giao thức mã hoá cho websocket
    sslopt = {
        'cert_reqs': ssl.PROTOCOL_SSLv23,
        'keyfile': KEY,
        'certfile': CRT,
        'ca_certs': CA_CRT,
    }

    with open('../.creds.json', 'r') as f:
        usersInfo = json.load(f)

    users = {}
    for username in usersInfo:
        users[username] = usersInfo[username]["password"]

    threads = []

    for user in users:
        t = Thread(createUserInstances(user, INSTANCE_PER_USER))
        t.start()
        threads.append(t)

    for t in threads:
        t.join()
