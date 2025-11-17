import { useCallback, useEffect, useRef } from 'react';
import type { WsMessage } from '@/types/common.ts';

export type MessageHandler = (msg: WsMessage) => void;

type WsEntry = {
  socket: WebSocket | null;
  handlers: Record<string, Set<MessageHandler>>;
  reconnectTimeout: NodeJS.Timeout | null;
  reconnectAttempts: number;
  consumers: number;
};
const wsRegistry: Record<string, WsEntry> = {};

export function useWebsocket({ ws_url }: { ws_url: string }) {
  if (!wsRegistry[ws_url]) {
    wsRegistry[ws_url] = {
      socket: null,
      handlers: {},
      reconnectTimeout: null,
      reconnectAttempts: 0,
      consumers: 0,
    };
  }

  const entry = wsRegistry[ws_url];
  const socketRef = useRef<WebSocket | null>(entry.socket);
  const handlersRef = useRef(entry.handlers);
  const maxReconnectAttempts = 5;
  const monitorRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const connect = useCallback(() => {
    //  Nếu đã có socket đang OPEN hoặc CONNECTING → KHÔNG tạo mới
    const state = entry.socket?.readyState;
    if (state === WebSocket.OPEN || state === WebSocket.CONNECTING) return;

    const ws = new WebSocket(ws_url);
    console.log('[WS] Creating new connection:', ws_url);
    entry.socket = ws;
    socketRef.current = ws;

    ws.onopen = () => {
      console.log('[WS] Connected');
      entry.reconnectAttempts = 0;
      // Start bufferedAmount monitor for debug if flood message
      // if (!monitorRef.current) {
      //   monitorRef.current = setInterval(() => {
      //     const wsCur = socketRef.current;
      //     if (!wsCur || wsCur.readyState !== WebSocket.OPEN) return;
      //     // bufferedAmount shows unsent outgoing bytes queued on the client
      //     // Useful to detect backpressure when sending
      //     console.log('[WS] bufferedAmount:', wsCur.bufferedAmount);
      //   }, 1000);
      // }
    };

    ws.onclose = (event) => {
      console.log('[WS] Closed:', event.code, event.reason);
      entry.socket = null;
      // Stop monitor on close
      if (monitorRef.current) {
        clearInterval(monitorRef.current);
        monitorRef.current = null;
      }

      if (event.code !== 1000 && entry.reconnectAttempts < maxReconnectAttempts) {
        const delay = Math.min(1000 * 2 ** entry.reconnectAttempts, 10000);
        console.log(`[WS] Reconnecting in ${delay}ms`);
        entry.reconnectTimeout = setTimeout(() => {
          entry.reconnectAttempts++;
          connect();
        }, delay);
      }
    };

    ws.onerror = (err) => console.error('[WS] Error:', err);

    ws.onmessage = (event) => {
      try {
        const msg: WsMessage = JSON.parse(event.data);
        const handlers = handlersRef.current[msg.type];
        handlers?.forEach((h) => h(msg));
      } catch (err) {
        console.error('[WS] Invalid message:', event.data, err);
      }
    };
  }, [ws_url]);

  useEffect(() => {
    entry.consumers++;
    connect();

    const handleBeforeUnload = () => {
      entry.socket?.close(1000, 'Page closing');
    };
    window.addEventListener('beforeunload', handleBeforeUnload);

    return () => {
      entry.consumers--;
      if (entry.consumers <= 0) {
        console.log('[WS] Closing shared connection');
        if (monitorRef.current) {
          clearInterval(monitorRef.current);
          monitorRef.current = null;
        }
        entry.socket?.close(1000, 'All consumers unmounted');
        entry.socket = null;
        entry.handlers = {};
        if (entry.reconnectTimeout) clearTimeout(entry.reconnectTimeout);
      }
      window.removeEventListener('beforeunload', handleBeforeUnload);
    };
  }, [connect]);

  const sendMsg = useCallback((msg: WsMessage) => {
    const ws = socketRef.current;
    if (ws?.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify(msg));
    } else {
      console.warn('[WS] Cannot send, socket not open');
    }
  }, []);

  const msgHandlerRegister = useCallback((type: string, handler: MessageHandler) => {
    if (!handlersRef.current[type]) handlersRef.current[type] = new Set();
    handlersRef.current[type].add(handler);
    return () => handlersRef.current[type].delete(handler);
  }, []);

  return { socketRef, sendMsg, msgHandlerRegister };
}
