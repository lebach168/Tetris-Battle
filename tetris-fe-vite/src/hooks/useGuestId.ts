"use client";

import { useEffect, useState } from "react";

export function useGuestId() {
  const [guestId, setGuestId] = useState<string | null>(null);

  useEffect(() => {
    if (typeof window !== "undefined") {
      let id = sessionStorage.getItem("guest_id");
      if (!id) {
        const characters = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789";
        id = "guest_";
        for (let i = 0; i < 4; i++) {
          id += characters.charAt(Math.floor(Math.random() * characters.length));
        }
        sessionStorage.setItem("guest_id", id);
      }
      setGuestId(id);
    }
  }, []);

  return guestId;
}
