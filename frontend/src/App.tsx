import React from "react";
import { BrowserRouter as Router, Routes, Route } from "react-router-dom";
import Navbar from "./components/Navbar";
import Home from "./components/Home";
import Play from "./components/Play";
import Login from "./components/Login";
import Register from "./components/Register";
import ProtectedRoute from "./components/ProtectedRoute";
import "./App.css";
import { AuthProvider } from "./context/AuthContext";
import { WSProvider } from "./context/WSContext";
import MatchMaking from "./components/MatchMaking";

const App: React.FC = () => {
  return (
    <Router>
      <AuthProvider>
        <WSProvider>
          <div className="app-container">
            <Navbar />
            <Routes>
              <Route path="/" element={<Home />} />
              <Route path="/login" element={<Login />} />
              <Route path="/register" element={<Register />} />
              <Route
                path="/matchmaking"
                element={
                  <ProtectedRoute>
                    <MatchMaking />
                  </ProtectedRoute>
                }
              />
              <Route
                path="/play/:gameId"
                element={
                  <ProtectedRoute>
                    <Play />
                  </ProtectedRoute>
                }
              />
            </Routes>
          </div>
        </WSProvider>
      </AuthProvider>
    </Router>
  );
};

export default App;
