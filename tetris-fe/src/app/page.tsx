"use client";
import { redirect } from "next/navigation";
import dynamic from "next/dynamic";
const GameplayWindow = dynamic(() => import("../components/Gameplay/PlayerSide"), { ssr: false });
export default function Home() {
  // return (

  // );
  redirect("/p1");
  return null;
}
