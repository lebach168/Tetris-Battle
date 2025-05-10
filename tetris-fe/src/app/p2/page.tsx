"use client";
import dynamic from "next/dynamic";
//const GameScreen = dynamic(() => import("@/components/Gameplay/GameScreen"), { ssr: false });
import { WebSocketProvider } from "@/components/WebSocketContext";
import GameScreen from "@/components/Gameplay/GameScreen";

export default function Player2Page() {
  return (
    <WebSocketProvider roomId="123456" playerId="p2">
      <GameScreen></GameScreen>
    </WebSocketProvider>
  );
}
