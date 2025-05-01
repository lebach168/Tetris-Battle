'use client';

import Link from 'next/link';

export default function Navbar() {
  return (
    <nav className="w-full bg-blue-600 text-white px-4 py-2 flex justify-between items-center">
      <div className="flex items-center space-x-4">
        <Link href="/" className="font-bold text-lg hover:text-blue-300">
          MyTetris
        </Link>
      </div>
      <div className="flex items-center space-x-4">
        <Link href="/login" className="hover:text-blue-300">
          Login
        </Link>
        <Link href="/register" className="hover:text-blue-300">
          Register
        </Link>
      </div>
    </nav>
  );
}
