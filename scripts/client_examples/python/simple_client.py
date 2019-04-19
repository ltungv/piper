import websocket
import json
import base64
import ssl
import time
import urllib.request


## Khai báo các hằng số được sử dụng


### Tên địa chỉ và cổng kết nối của hệ thống
HOST = "vgurobocon2019.local"
PORT = 4433

### Các giao thức được sử dụng
HTTPS = "https"
WSS = "wss"


### Các tập tin dùng để mã hoá, giải mã và xác minh dữ liệu
CA_CRT = "./keys/certs/pub/cacert.pem"
CRT = "./keys/certs/pub/clientcert.pem"
KEY = "./keys/certs/priv/clientkey.pem"


### Địa chỉ đăng nhập
SUB_URL = "%s://%s:%d/subscribe" % (HTTPS, HOST, PORT)

### Thông tin đăng nhập của đội
CREDS = {
    "username": "user2",
    "password": "password2"
}


## Thiết lập thông tin xác thực và mã hoá


### Cài đặt thông tin xác thực và mã hoá
sslCTX = ssl.create_default_context(
    purpose=ssl.Purpose.SERVER_AUTH,
    cafile=CA_CRT
)
sslCTX.load_cert_chain(CRT, KEY)


## Đăng nhập vào hệ thống


### Chuyển đổi thông tin đăng nhập thành chuỗi ký tự JSON với định dạng utf-8 bytes
credsJSON = json.dumps(CREDS).encode("utf-8")

### Thiết lập yêu cầu HTTP
req = urllib.request.Request(SUB_URL)
req.add_header('Content-Type', 'application/json; charset=utf-8')
req.add_header('Content-Length', len(credsJSON))


## Đăng nhập vào hệ thống

### Gửi yêu cầu HTTP và nhận kết quả trả về
resp = urllib.request.urlopen(
    req,               # Dạng yêu cầu
    credsJSON,    # Thông tin được gửi đi
    context=sslCTX # Hệ thống xác thực được sử dụng
)

### Đọc và nhận mã xác thực
respBody = resp.read()
respBodyJSON = json.loads(respBody)

token = respBodyJSON["token"]


## Khai báo các hàm số được sử dụng


### Nhận tin nhắn từ hệ thống
def on_message(ws, message):
    # Biến đổi tin nhắn dạng JSON
    # thành tham số dạng từ điển trong python
    data = json.loads(message)
    print((time.time_ns() - data['time']) * 1e-6)
    resp = {"finished": True}
    ws.send(json.dumps(resp).encode('utf-8'))

### Xử lý lỗi xảy ra
def on_error(ws, error):
    print(error)

### Xử lý kết nối bị ngắt
def on_close(ws):
    print("### closed ###")


## Bắt đầu nhận dữ liệu


### Thêm mã xác thực vào thông tin của yêu cầu HTTP
header = {
    "Authorization": respBodyJSON['token']
}

### Cài đặt thông tin xác thực và mã hoá cho kết nối websocket
sslopt = {
    'cert_reqs': ssl.PROTOCOL_SSLv23,
    'keyfile': KEY,
    'certfile': CRT,
    'ca_certs': CA_CRT,
}


## Nhận dữ liệu


### Cài đặt thông tin cho kết nối websocket
ws = websocket.WebSocketApp(
    # Địa chỉ nhận data
    "%s://%s:%s/data" % (WSS, HOST, PORT),
    # Thông tin của yêu cầu
    header = header,
    # Cài đặt các hàm số được dùng
    on_message = on_message,
    on_error = on_error,
    on_close = on_close,
)

### Chạy và giữ kết nối với hệ thống
ws.run_forever(sslopt=sslopt)
