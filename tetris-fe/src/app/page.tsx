'use client'
import dynamic from "next/dynamic";
const GameplayWindow = dynamic(() => import("../components/Gameplay/GameplayWindow"), { ssr: false });
export default function Home() {
  return (
    <GameplayWindow/>
  );
}
