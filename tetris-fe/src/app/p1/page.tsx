import GameScreen from "@/components/Gameplay/GameScreen";
import PlayerSide from "@/components/Gameplay/PlayerSide";
import { WebSocketProvider } from "@/components/WebSocketContext";

export default function Player1Page() {
  return (
    <WebSocketProvider roomId="123456" playerId="p1">
      <GameScreen></GameScreen>
    </WebSocketProvider>
  );
}
