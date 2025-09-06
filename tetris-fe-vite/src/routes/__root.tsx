import { Outlet } from '@tanstack/react-router';
import Navbar from '../components/Navbar';
//import { useGuestId } from '@/hooks/useGuestId';
import { createRootRoute } from '@tanstack/react-router';
import { TanStackRouterDevtools } from '@tanstack/react-router-devtools';
import './globals.css'; 

const RootLayout = () => {
  //const guestId = useGuestId();

  return (
    <div className="min-h-screen flex flex-col">
      <Navbar />
      <main className="flex-1 p-4 bg-gray-100 w-[75%] max-w-screen-2xl mx-auto px-4">
        <Outlet />
      </main>
      <TanStackRouterDevtools />
    </div>
  );
};

export const Route = createRootRoute({
  component: RootLayout,
});
