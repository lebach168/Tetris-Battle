
import { Link } from '@tanstack/react-router';

export default function Navbar() {
  return (
    <nav className="w-full bg-blue-600 text-white px-4 py-2 flex justify-between items-center">
      <div className="flex items-center space-x-4">
        <Link to="/" className="font-bold text-lg hover:text-blue-300">
          MyTetris
        </Link>
      </div>
      <div className="flex items-center space-x-4">
        <Link to="/login" className="hover:text-blue-300">
          Login
        </Link>
        <Link to="/register" className="hover:text-blue-300">
          Register
        </Link>
      </div>
    </nav>
  );
}
