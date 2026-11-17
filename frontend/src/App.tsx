import { useEffect, useState } from "react";
import { Routes, Route, Link, NavLink } from "react-router-dom";
import DashboardPage from "./pages/DashboardPage";
import EndpointsPage from "./pages/EndpointsPage";
import EndpointDetailPage from "./pages/EndpointDetailPage";

export default function App() {
  const [dark, setDark] = useState(() => localStorage.getItem("theme") === "dark");
  useEffect(() => {
    document.documentElement.classList.toggle("dark", dark);
    localStorage.setItem("theme", dark ? "dark" : "light");
  }, [dark]);

  const navCls = ({ isActive }: { isActive: boolean }) =>
    `px-3 py-1.5 rounded-md text-sm ${isActive ? "bg-gray-100 dark:bg-gray-800 font-medium" : "text-gray-600 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800"}`;

  return (
    <div className="min-h-screen">
      <header className="border-b border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900">
        <div className="max-w-6xl mx-auto px-4 h-14 flex items-center justify-between">
          <div className="flex items-center gap-6">
            <Link to="/" className="font-semibold text-lg">WebhookLab</Link>
            <nav className="flex items-center gap-1">
              <NavLink to="/" end className={navCls}>Dashboard</NavLink>
              <NavLink to="/endpoints" className={navCls}>Endpoints</NavLink>
            </nav>
          </div>
          <button className="btn" onClick={() => setDark((d) => !d)}>
            {dark ? "Light" : "Dark"}
          </button>
        </div>
      </header>
      <main className="max-w-6xl mx-auto px-4 py-6">
        <Routes>
          <Route path="/" element={<DashboardPage />} />
          <Route path="/endpoints" element={<EndpointsPage />} />
          <Route path="/endpoints/:id" element={<EndpointDetailPage />} />
        </Routes>
      </main>
    </div>
  );
}
