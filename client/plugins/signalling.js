class Signalling {
  constructor(websocketURL) {
    if (!!Signalling.instance) {
      return Signalling.instance;
    }

    this._websocketUrl = websocketURL;
    this._websocket = null;

    return this
  }

  open(authToken) {
    if (this._websocket) {
      console.error('[Signalling] Websocket connection already established.');
      return;
    }

    this._websocket = new WebSocket(`${this._websocketUrl}?auth_token=${authToken}`)

    this._websocket.onopen = () => {
      console.log('[Signalling] Websocket connection opened.')
    };

    this._websocket.onerror = (err) => {
      console.log('[Signalling] Websocket error: ', err)
    };

    this._websocket.onclose = (evt) => {
      console.log('[Signalling] Websocket connection closed: ', evt);
      this._websocket = null
    }

    this._websocket.onmessage = (evt) => {
      console.log('[Signalling] Received websocket message: ', evt.data)
    }
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
}

export default function ({ app }, inject) {
  const signalling = new Signalling(process.env.WS_URL);
  app.$Signalling = signalling
  inject('Signalling', signalling)
}
