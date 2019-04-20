# Outline

- Tổng quan hệ thống server
  - HTTPS
  - User login
  - Websocket
- Ví dụ cách kết nối với hệ thống (Demo)

# Tổng quan

## HTTPS - SSL/TLS

Server của ban tổ chức sẽ được mã hoá bằng giao thức HTTPS nhằm hạn chế gian lận xảy ra trong quá trình thi đấu

Mỗi đội sẽ được BTC cung cấp 3 tập tin, sử dụng để xác thức và mã hoá thông tin được trao đổi giữa server và mỗi đội

- `cacert.pem`: Tập tin này được dùng để xác thực danh tính của hệ thống, nhằm đảm bảo máy của mỗi đội đang kết nối với đúng hệ thống của BTC
- `clientcert.pem`: Tập tin này sẽ được gửi đến cho hệ thống để thực hiện 2 công việc sau:
  - Xác nhận danh tính của mỗi đội
  - Mã hoá thông tin được gửi về máy mỗi đội
- `clientkey.pem`: Tập tin này được dùng để giải mã thông tin được gửi về từ hệ thống

## User login

- Mỗi đội sẽ được BTC cung cấp tài khoản và mật khẩu để đăng nhập vào hệ thống
- Khi đăng nhập thành công hệ thống sẽ trả về một mã xác thực để đội có thể sử dụng và bắt đầu nhận dữ liệu
- Mỗi mã xác thực sẽ có thời hạn sử dụng 10 phút

## Websocket

- Mã xác thực được nhận từ hệ thống sẽ được sử dụng để thiết lập kết nối websocket
- Sau khi thiết lập kết nối, dữ liệu từ hệ thống nhận diện vị trí sẽ được gửi về máy các đội
- Mỗi đội chỉ được sử dụng cùng một lúc tối đa 2 máy để kết nối với hệ thống, hệ thống sẽ tự động ngắt những kết nối tiếp theo
- Dữ liệu được gửi từ hệ thống về mỗi đội sẽ có định dạng JSON và được mã hoá với dạng base64 bytes

# Ví dụ cách kết nối với hệ thống (Demo)

## Ngôn ngữ và thư viện

- Ngôn ngữ lập trình Python3.
- Thư viện:
  - [json](https://docs.python.org/3/library/json.html#module-json) (Đọc và dịch chuỗi json)
  - [base64](https://docs.python.org/3/library/base64.html#module-base64) (Đọc và giải mã base64 bytes)
  - [ssl](https://docs.python.org/3/library/ssl.html#module-ssl) (Cài đặt hệ thống bảo mật)
  - [time](https://docs.python.org/3/library/time.html#module-time) (Sử lý dữ liệu thời gian)
  - [urllib.request](https://docs.python.org/3/library/urllib.request.html#module-urllib.request) (Gửi yêu cầu HTTP)
  - [websocket-client](https://github.com/websocket-client/websocket-client) (Thiết lập kết nối websocket)
    ```bash
    pip install websocket-client
    ```

## Demo

### Khai báo các thư viện được sử dụng

```python
import websocket
import json
import base64
import ssl
import time
import urllib.request
```

### Khai báo các hằng số được sử dụng

```python
# Tên địa chỉ và cổng kết nối của hệ thống
HOST = "tungle.local"
PORT = 4433

# Các giao thức được sử dụng
HTTPS = "https"
WSS = "wss"

# Các tập tin dùng để mã hoá, giải mã và xác minh dữ liệu
CA_CRT = "./keys/certs/pub/cacert.pem"
CRT = "./keys/certs/pub/clientcert.pem"
KEY = "./keys/certs/priv/clientkey.pem"

# Địa chỉ đăng nhập
SUB_URL = "%s://%s:%d/sub" % (HTTPS, HOST, PORT)

# Thông tin đăng nhập của đội
CREDS = {
    "username": "user2",
    "password": "password2"
}
```

### Thiết lập thông tin xác thực và mã hoá

```python
# Cài đặt thông tin xác thực và mã hoá
sslCTX = ssl.create_default_context(
    purpose=ssl.Purpose.SERVER_AUTH # Giao thức xác thức được sử dụng,
    cafile=CA_CRT,                  # Chìa khoá xác thực
)

# Cài đặt chìa khoá mã hoá và giải mã
sslCTX.load_cert_chain(CRT, KEY)
```

### Đăng nhập vào hệ thống

```python
# Chuyển đổi thông tin đăng nhập thành chuỗi ký tự JSON
# với định dạng utf-8 bytes
credsJSON = json.dumps(creds).encode("utf-8")

# Thiết lập yêu cầu HTTP
req = urllib.request.Request(url)
# Dạng tin nhắn được gửi đi
req.add_header('Content-Type', 'application/json; charset=utf-8')
# Độ dài tin nhắn được gửi đi
req.add_header('Content-Length', len(credsJSONBytes))

# Gửi yêu cầu HTTP và nhận kết quả trả về
resp = urllib.request.urlopen(
    req,               # Dạng yêu cầu
    credsJSONBytes,    # Thông tin được gửi đi
    context=sslContext # Hệ thống xác thực được sử dụng
)

# Đọc và chuyển dổi kết quả trả về
# thành tham số dạng từ điển trong python
respBody = resp.read()
respBodyJSON = json.loads(respBody)

# Nhận mã xác thực từ kết quả trả về
token = respBodyJSON["token"]
```

### Thiết lập giao thức kết nối websocket với hệ thống

#### Khai báo các hàm số được sử dụng

```python
# Hàm này được gọi khi nhận được 1 tin nhắn từ hệ thống
def on_message(ws, message):
    # Biến đổi tin nhắn dạng JSON
    # thành tham số dạng từ điển trong python
    data = json.loads(message)
    print(data)

# Hàm này được gọi khi lỗi xảy ra trong quá trình kết nối với hệ thống
def on_error(ws, error):
    print(error)

# Hàm này được gọi khi kết nối với hệ thống bị ngắt
def on_close(ws):
    print("### closed ###")
```

#### Bắt đầu nhận dữ liệu

```python
# Thêm mã xác thực vào thông tin của yêu cầu HTTP
header = {
    "Authorization": respBodyJSON['token']
}

# Cài đặt thông tin xác thực và mã hoá cho kết nối websocket
sslopt = {
    'cert_reqs': ssl.PROTOCOL_SSLv23,
    'keyfile': clientkey,
    'certfile': clientcert,
    'ca_certs': cacert,
}

# Cài đặt thông tin cho kết nối websocket
websocket.enableTrace(True)
ws = websocket.WebSocketApp(
    # Địa chỉ nhận data
    "%s://%s:%s/data" % (WSS, HOST, PORT),
    # Thông tin của yêu cầu
    header = header
    # Cài đặt các hàm số được dùng
    on_message = on_message,
    on_error = on_error,
    on_close = on_close,
)

# Chạy và giữ kết nối với hệ thống
ws.run_forever(sslopt=sslopt)
```
