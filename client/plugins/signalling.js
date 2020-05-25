class Signalling {
  constructor(websocketURL) {
    if (!!Signalling.instance) {
      return Signalling.instance;
    }

    this._websocketUrl = websocketURL;
    this._websocket = null;
    this._onSignallingMessage = null;
    this._hearbeatId = null;

    Signalling.instance = this
    return this
  }

  open(authToken) {
    if (this._websocket) {
      console.error('[Signalling] Websocket connection already established.');
      return;
    }

    return new Promise((resolve, reject) => {
      this._websocket = new WebSocket(`${this._websocketUrl}?auth_token=${authToken}`)

      this._websocket.onopen = () => {
        resolve();
        this.heartbeat();
        console.log('[Signalling] Websocket connection opened.')
      };

      this._websocket.onerror = (err) => {
        reject();
        console.log('[Signalling] Websocket error: ', err)
      };

      this._websocket.onclose = (evt) => {
        console.log('[Signalling] Websocket connection closed: ', evt);
        clearInterval(this._hearbeatId);
        this._websocket = null
      }

      this._websocket.onmessage = (evt) => {
        if (this._onSignallingMessage) {
          this._onSignallingMessage(evt.data)
        }
      }
    });
  }

  send(type, payload) {
    const wsMessage = {
      type,
      payload
    };

    if (this._websocket && this._websocket.readyState === WebSocket.OPEN) {
      this._websocket.send(JSON.stringify(wsMessage))
    }
  }

  heartbeat() {
    this._hearbeatId = setInterval(() => {
      this.send('HEARTBEAT', {
        data: 'PING'
      });
    }, 30000);
  }

  registerOnSignallingMessageHandler(handler) {
    this._onSignallingMessage = handler
  }
}

export default function ({ app }, inject) {
  const websocketUrl = process.env.WS_URL || 'wss://konference-api.herokuapp.com/ws';
  const signalling = new Signalling(websocketUrl);
  app.$Signalling = signalling;
  inject('Signalling', signalling)
}
