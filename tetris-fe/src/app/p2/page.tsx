

import GameScreen from "@/components/Gameplay/GameScreen";
import { WebSocketProvider } from "@/components/WebSocketContext";

export default function Player2Page() {
  return (
    <WebSocketProvider roomId="123456" playerId="p2">
      <GameScreen ></GameScreen>
    </WebSocketProvider>
  );
}
