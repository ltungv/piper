const WebSocket = require('ws');

let host = process.argv[2] || '0.0.0.0';
let port = process.argv[3] || '8000';
let n = process.argv[4] || 1;

function runClient() {
  for (let i = 0; i < n; i++) {
    const ws = new WebSocket(`http://${host}:${port}/ws`);

    ws.on('message', function (msg) {
        packet = JSON.parse(msg)
        ws.ping(packet.time)
      //console.log(msg);
    });

    ws.on('pong', function(msg) {
      console.log(msg.toString())
    })

    ws.on('close', function() {
        runClient();
    })
  }
}

runClient();
