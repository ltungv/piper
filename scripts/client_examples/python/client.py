#!/usr/local/bin/python3

import websocket # Documentations https://github.com/websocket-client/websocket-client
import json

# Địa chỉ IP của server
HOST = "35.247.153.42"

# Function này được gọi khi nhận được 1 tin nhắn từ server
def on_message(ws, message):
    # Biến đổi tin nhắn dạng string với format JSON
    # thành một variable có dạng dictionary
    data = json.loads(message)

    print(data)

# Function này được gọi khi có lỗi xảy ra khi kết nối với server
def on_error(ws, error):
    print(error)

# Function này được gọi khi kết nối với server bị ngắt
def on_close(ws):
    print("### closed ###")

if __name__ == "__main__":
    # Tạo kết nối websocket với server
    websocket.enableTrace(True)
    ws = websocket.WebSocketApp("ws://%s:8000/ws" % (HOST),
                              on_message = on_message,
                              on_error = on_error,
                              on_close = on_close)

    # Giữ kết nối vĩnh viễn với server
    ws.run_forever()
