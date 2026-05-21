export interface ReconnectingWebSocketOptions {
  onOpen?: (isReconnect: boolean) => void;
  onMessage?: (event: MessageEvent) => void;
  onClose?: () => void;
  initialBackoffMs?: number;
  maxBackoffMs?: number;
}

export class ReconnectingWebSocket {
  private url: string;
  private onOpen: (isReconnect: boolean) => void;
  private onMessage: (event: MessageEvent) => void;
  private onClose: () => void;
  private initialBackoffMs: number;
  private maxBackoffMs: number;

  private socket: WebSocket | null = null;
  private intentionallyClosed = false;
  private hasConnectedBefore = false;
  private attempt = 0;
  private reconnectTimer: number | null = null;

  constructor(url: string, opts: ReconnectingWebSocketOptions = {}) {
    this.url = url;
    this.onOpen = opts.onOpen ?? (() => {});
    this.onMessage = opts.onMessage ?? (() => {});
    this.onClose = opts.onClose ?? (() => {});
    this.initialBackoffMs = opts.initialBackoffMs ?? 500;
    this.maxBackoffMs = opts.maxBackoffMs ?? 30000;
    this.connect();
  }

  private connect() {
    this.socket = new WebSocket(this.url);

    this.socket.onopen = () => {
      const wasReconnect = this.hasConnectedBefore;
      this.hasConnectedBefore = true;
      this.attempt = 0;
      this.onOpen(wasReconnect);
    };

    this.socket.onmessage = (event) => this.onMessage(event);

    this.socket.onclose = () => {
      if (this.intentionallyClosed) {
        this.onClose();
        return;
      }
      this.scheduleReconnect();
    };

    this.socket.onerror = (err) => {
      console.error("ws error:", err);
    };
  }

  private scheduleReconnect() {
    const delay = Math.min(
      this.initialBackoffMs * Math.pow(2, this.attempt),
      this.maxBackoffMs
    );
    this.attempt++;
    console.log(`[ws] reconnecting in ${delay}ms (attempt ${this.attempt})`);
    this.reconnectTimer = window.setTimeout(() => this.connect(), delay);
  }

  send(data: string) {
    if (this.socket && this.socket.readyState === WebSocket.OPEN) {
      this.socket.send(data);
    } else {
      console.warn("[ws] send dropped: socket not open");
    }
  }

  close() {
    this.intentionallyClosed = true;
    if (this.reconnectTimer !== null) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
    this.socket?.close();
  }
}
