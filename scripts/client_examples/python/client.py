# websocket-client
import websocket
import json

def on_message(ws, message):
    data = json.loads(message)
    print(data[0]['type'])
    print(data[0]['pos'])
    print(data[0]['dim'])


def on_error(ws, error):
    print(error)

def on_close(ws):
    print("### closed ###")

if __name__ == "__main__":
    websocket.enableTrace(True)
    ws = websocket.WebSocketApp("ws://35.247.153.42:8000/ws",
                              on_message = on_message,
                              on_error = on_error,
                              on_close = on_close)
    ws.run_forever()
