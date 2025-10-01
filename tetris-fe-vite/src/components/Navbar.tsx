/* eslint-disable no-unused-vars */
import { Link } from '@tanstack/react-router';
import { useEffect, useState } from 'react';

export default function Navbar() {
  const [isDark, setIsDark] = useState<boolean>(false);

  useEffect(() => {
    const hasDark = document.documentElement.classList.contains('dark');
    setIsDark(hasDark);
  }, []);

  const toggleTheme = () => {
    const nextIsDark = !isDark;
    setIsDark(nextIsDark);
    if (nextIsDark) {
      document.documentElement.classList.add('dark');
      localStorage.setItem('theme', 'dark');
    } else {
      document.documentElement.classList.remove('dark');
      localStorage.setItem('theme', 'light');
    }
  };
  return (
    <nav className="w-full bg-primary text-white px-4 py-2 flex justify-between items-center">
      <div className="flex items-center space-x-4">
        <Link to="/" className="font-bold text-lg hover:text-blue-300">
          MyTetris
        </Link>
        <Link to="/room" className="hover:text-blue-300">
          Play
        </Link>
        {/* <Link to="/leaderboard" className="hover:text-blue-300">
          Leaderboard
        </Link> */}
        <Link to="/about" className="hover:text-blue-300">
          About
        </Link>
      </div>
      <div className="flex items-center space-x-4">
        <Link to="/login" className="hover:text-blue-300">
          Login
        </Link>
        {/* <button
          type="button"
          onClick={toggleTheme}
          className="inline-flex items-center gap-2 bg-secondary text-secondary-foreground hover:bg-muted border border-border rounded-md px-3 py-1 text-sm transition-colors"
          aria-label="Toggle theme"
        >
          {isDark ? 'Light' : 'Dark'}
        </button> */}
        
      </div>
    </nav>
  );
}
