export const Navbar = () => {
  return (
    <nav className="navbar">
      <div className="nav-brand">
        ♞ <span>Chess</span>Arena
      </div>
      <div className="nav-links">
        <a href="#play" className="nav-item">
          Play
        </a>
        <a href="#puzzles" className="nav-item">
          Puzzles
        </a>
        <button className="btn btn-login">Log In</button>
        <button className="btn btn-signup">Sign Up</button>
      </div>
    </nav>
  );
};
