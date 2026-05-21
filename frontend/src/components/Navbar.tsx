import React from "react";
import { Link, useNavigate } from "react-router-dom";
import { useAuth } from "../context/AuthContext";

const Navbar: React.FC = () => {
  const { user, logout } = useAuth();
  const navigate = useNavigate();

  const handleLogout = async () => {
    await logout();
    navigate("/login");
  };

  return (
    <nav className="navbar">
      <Link to="/" className="nav-brand">
        ♞ <span>Chess</span>Arena
      </Link>

      <div className="nav-links">
        {user ? (
          <>
            <Link to="/games" className="nav-item">My Games</Link>
            <span className="nav-item nav-username">{user.username}</span>
            <button className="btn btn-login" onClick={handleLogout}>
              Log Out
            </button>
          </>
        ) : (
          <>
            <Link to="/login" className="btn btn-login">
              Log In
            </Link>
            <Link to="/register" className="btn btn-signup">
              Sign Up
            </Link>
          </>
        )}
      </div>
    </nav>
  );
};

export default Navbar;
