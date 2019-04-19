<!-- $size: 4:3 -->

# Outline

- Tổng quan hệ thống
  - Secure Hypertext Transfer Protool (HTTPS)
  - JSON Web Token (JWT)
  - Websocket
- Phương thức hoạt động
  - Đăng nhập
  - Nhận dữ liệu
- Demo
- Một vài lưu ý

---

# Tổng quan hệ thống

---

## Secure Hyper Text Transfer Protocol (HTTPS)

- Xác thực kết nối giữa 2 bên
- Mã hoá thông tin trao đổi

---

## JSON Web Token (JWT)

- Xác thực người dùng
- Hạn chế thời gian đăng nhập
- Hạn chết số lượng đăng nhập

---

## Websocket

- Hỗ trợ kết nối thời gian thực
- Đảm bảo độ trễ thấp
- Được hỗ trợ bởi nhiều nền tảng

---

# Phương thức hoạt động

---

## Hệ thống mã hoá

Mỗi đội sẽ được BTC cung cấp 3 tập tin, sử dụng để xác thức và mã hoá thông tin được trao đổi giữa server và mỗi đội
- `cacert.pem`: Xác nhận danh tính
- `clientcert.pem`: Mã hoá thông tin
- `clientkey.pem`: Giải mã thông tin

---

## Đăng nhập

- Đăng nhập bằng tài khoản được cung cấp
- Nhận mã xác thực khi đăng nhập thành công
    - Thời hạn sử dụng 10 phút

---

## Nhận dữ liệu

- Thiết lập kết nối bằng mã xác thực
- Dữ liệu được gửi về ngay sau khi kết nối

---

# Demo

---

## Ngôn ngữ và thư viện

- Ngôn ngữ lập trình Python3.
- Thư viện:
  - [json](https://docs.python.org/3/library/json.html#module-json) (Đọc và dịch chuỗi json)
  - [base64](https://docs.python.org/3/library/base64.html#module-base64) (Đọc và giải mã base64 bytes)
  - [ssl](https://docs.python.org/3/library/ssl.html#module-ssl) (Cài đặt hệ thống bảo mật)
  - [time](https://docs.python.org/3/library/time.html#module-time) (Xử lý dữ liệu thời gian)
  - [urllib.request](https://docs.python.org/3/library/urllib.request.html#module-urllib.request) (Gửi yêu cầu HTTP)
  - [websocket-client](https://github.com/websocket-client/websocket-client) (Thiết lập kết nối websocket)
    ```bash
    pip install websocket-client
    ```

---

### Khai báo các thư viện được sử dụng

```python
import websocket
import json
import base64
import ssl
import time
import urllib.request
```

---

## Khai báo các hằng số được sử dụng

---

### Tên địa chỉ và cổng kết nối của hệ thống
```python
HOST = "tungle.local"
PORT = 4433
```

### Các giao thức được sử dụng
```python
HTTPS = "https"
WSS = "wss"
```

---

### Các tập tin dùng để mã hoá, giải mã và xác minh dữ liệu
```python
CA_CRT = "./keys/certs/pub/cacert.pem"
CRT = "./keys/certs/pub/clientcert.pem"
KEY = "./keys/certs/priv/clientkey.pem"
```

---

### Địa chỉ đăng nhập
```python
SUB_URL = "%s://%s:%d/subscribe" % (HTTPS, HOST, PORT)
```

### Thông tin đăng nhập của đội
```python
CREDS = {
    "username": "user2",
    "password": "password2"
}
```

---

## Thiết lập thông tin xác thực và mã hoá

---

### Cài đặt thông tin xác thực và mã hoá
```python
sslCTX = ssl.create_default_context(
    purpose=ssl.Purpose.SERVER_AUTH,
    cafile=CA_CRT
)
sslCTX.load_cert_chain(CRT, KEY)
```

---

## Đăng nhập vào hệ thống

---

### Chuyển đổi thông tin đăng nhập thành chuỗi ký tự JSON với định dạng utf-8 bytes
```python
credsJSON = json.dumps(CREDS).encode("utf-8")
```

### Thiết lập yêu cầu HTTP
```python
req = urllib.request.Request(SUB_URL)
req.add_header('Content-Type', 'application/json; charset=utf-8')
req.add_header('Content-Length', len(credsJSON))
```

---

## Đăng nhập vào hệ thống

### Gửi yêu cầu HTTP và nhận kết quả trả về
```python
resp = urllib.request.urlopen(
    req,               # Dạng yêu cầu
    credsJSON,    # Thông tin được gửi đi
    context=sslCTX # Hệ thống xác thực được sử dụng
)
```

### Đọc và nhận mã xác thực
```python
respBody = resp.read()
respBodyJSON = json.loads(respBody)

token = respBodyJSON["token"]
```

---

## Khai báo các hàm số được sử dụng

---

### Nhận tin nhắn từ hệ thống
```python
def on_message(ws, message):
    # Biến đổi tin nhắn dạng JSON
    # thành tham số dạng từ điển trong python
    data = json.loads(message)
    print(data)
```

### Xử lý lỗi xảy ra
```python
def on_error(ws, error):
    print(error)
```

### Xử lý kết nối bị ngắt
```python
def on_close(ws):
    print("### closed ###")
```

---

## Bắt đầu nhận dữ liệu

---

### Thêm mã xác thực vào thông tin của yêu cầu HTTP
```python
header = {
    "Authorization": respBodyJSON['token']
}
```

### Cài đặt thông tin xác thực và mã hoá cho kết nối websocket
```python
sslopt = {
    'cert_reqs': ssl.PROTOCOL_SSLv23,
    'keyfile': KEY,
    'certfile': CRT,
    'ca_certs': CA_CRT,
}
```

---

## Nhận dữ liệu

---

### Cài đặt thông tin cho kết nối websocket
```python
websocket.enableTrace(True)
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
```

### Chạy và giữ kết nối với hệ thống
```python
ws.run_forever(sslopt=sslopt)
```

---

# Một vài lưu ý

- Tối đa 2 kết nối cho mỗi tài khoản
- Mọi dữ liệu trao đổi đều được mã hoá dưới dạng base64
- Ngắt kết nối khi không nhận kịp dữ liệu
