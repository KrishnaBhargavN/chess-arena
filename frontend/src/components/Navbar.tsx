import React from "react";
import { Link } from "react-router-dom";
const Navbar: React.FC = () => {
  return (
    <nav className="navbar">
      <Link to="/" className="nav-brand">
        ♞ <span>Chess</span>Arena
      </Link>

      <div className="nav-links">
        <Link to="/play" className="nav-item">
          Play
        </Link>
        <button className="btn btn-login">Log In</button>
        <button className="btn btn-signup">Sign Up</button>
      </div>
    </nav>
  );
};

export default Navbar;
