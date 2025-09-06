
// Update the import path below if WSMessage is located elsewhere in your project
// Update the import path below if WSMessage is located elsewhere in your project
import type { WSMessage } from "../types/common";
import { createContext, type ReactNode, useCallback, useContext, useEffect, useRef, useState } from "react";

const API_URL = process.env.NEXT_PUBLIC_API_URL;


export type SendMessageFn = (msg: WSMessage) => void;
export type ReceiveMessageFn = (handler: (msg: WSMessage) => void) => () => void;

type WSContextType = {
  
  socket: WebSocket | null;
  sendMessage: SendMessageFn;
  subscribe: ReceiveMessageFn;
};
const WebSocketContext = createContext<WSContextType | null>(null);
export const WebSocketProvider = ({
  roomId,
  playerId,
  children,
}: {
  roomId: string;
  playerId: string;
  children: ReactNode;
}) => {
  const socketRef = useRef<WebSocket | null>(null); //tránh bị closure
  const [socket, setSocket] = useState<WebSocket | null>(null);
  const messageHandlersRef = useRef<Set<(rawMsg: WSMessage) => void>>(new Set());
  useEffect(() => {
    roomId = "123456"; //fixed test
    const url  = API_URL || "" 
    const wsUrl = url.replace(/^http/, "ws") + `/ws?room=${roomId}&player=${playerId}`;
    const ws = new WebSocket(wsUrl);
    setSocket(ws);
    socketRef.current = ws;
    ws.onopen = () => console.log("WebSocket connected");
    ws.onclose = () => console.log("WebSocket closed");
    ws.onerror = (err) => console.error("WebSocket error", err);
    const handleMessage = (event:MessageEvent)=>{
      let msg: WSMessage;
      try {
        msg = JSON.parse(event.data);
      } catch {
        console.warn("Invalid WS message:", event.data);
        return;
      }
       //send msg to all handler -- handler filter msg later
       messageHandlersRef.current.forEach((h) => h(msg));
    }
    ws.onmessage = handleMessage;
    const handleBeforeUnload = () => {
      if (socketRef.current && socketRef.current.readyState === WebSocket.OPEN) {
        socketRef.current.close(); // Đóng WebSocket
      }
    };

    window.addEventListener('beforeunload', handleBeforeUnload);
    //cleanup function
    return () => {
      ws.onmessage = null;
      ws.close();
      socketRef.current = null;
      setSocket(null);
      messageHandlersRef.current.clear();
      window.removeEventListener('beforeunload', handleBeforeUnload);
    };
  }, [roomId, playerId]);

  const sendMessage = useCallback<SendMessageFn>((msg:WSMessage) => {
    if (socketRef.current?.readyState === WebSocket.OPEN) {
      socketRef.current.send(JSON.stringify(msg));
    }
  }, []);
  const subscribe= useCallback<ReceiveMessageFn>((handler)=>{
    messageHandlersRef.current.add(handler)
    //return cleanup unsub callback
    return ()=>{
      messageHandlersRef.current.delete(handler)
    }
  },[])


  return <WebSocketContext.Provider value={{socket,sendMessage,subscribe}}>{children}</WebSocketContext.Provider>;
};
export const useWebSocket = ():WSContextType => {
  const ctx = useContext(WebSocketContext);
  if (!ctx) throw new Error("useWebSocket must be inside WebSocketProvider");
  return ctx;
};
