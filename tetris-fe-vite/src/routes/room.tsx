
import ListRoom from '@/components/Room/ListRoom';
import { createFileRoute } from '@tanstack/react-router';

export const Route = createFileRoute("/room")({
  component: ListRoom,
});